package cache

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/hash"
	"strconv"
	"time"
	"webook/interactive/domain"
)

var (
	//go:embed lua/incr_cnt.lua
	luaIncrCnt string
	//go:embed lua/interactive_ranking_incr.lua
	luaRankingIncr string
	//go:embed lua/interactive_ranking_set.lua
	luaRankingSet string
)

const fieldReadCnt = "read_cnt"
const fieldLikeCnt = "like_cnt"
const fieldCollectCnt = "collect_cnt"

var RankingUpdateErr = errors.New("指定的元素不存在")

type InteractiveCache interface {
	IncrReadCntIfPresent(ctx context.Context, biz string, bizId int64) error
	IncrLikeCntIfPresent(ctx context.Context, biz string, bizId int64) error
	DecrLikeCntIfPresent(ctx context.Context, biz string, bizId int64) error
	IncrCollectCntIfPresent(ctx context.Context, biz string, bizId int64) error
	Get(ctx context.Context, biz string, bizId int64) (domain.Interactive, error)
	Set(ctx context.Context, biz string, bizId int64, res domain.Interactive) error
	// IncrRankingIfPresent 如果排名数据存在就+1
	IncrRankingIfPresent(ctx context.Context, biz string, bizId int64) error
	// SetRankingScore 如果排名数据不存在就把数据库中读取到的更新到缓存，如果更新过就+1
	SetRankingScore(ctx context.Context, biz string, bizId int64, count int64) error
	// LikeTop 基本实现，是借助 zset
	// 1. 前 100 名是一个高频数据，你可以结合本地缓存。
	//    你可以定时刷新本地缓存，比如说每 5s 调用 LikeTop，放进去本地缓存
	// 2. 如果你有一亿的数据，你怎么实时维护？zset 放 一亿个元素，你的 Redis 撑不住
	// 		2.1 不是真的维护一亿，而是维护近期的数据的点赞数，比如说三天内的
	//      2.2 你要分 key。这是 Redis 解决大数据结构常见的方案
	// 3. 借助定时任务，我每分钟计算一次。如果你有很多数据，一分钟不够你遍历一遍
	// 4. 我定时计算，算 1000 名；而后我借助 zset 来实时维护者 1000 名的分数
	LikeTop(ctx context.Context, biz string) ([]domain.Interactive, error)
}

type InteractiveRedisCache struct {
	client redis.Cmdable
}

func (r *InteractiveRedisCache) Get(ctx context.Context, biz string, bizId int64) (domain.Interactive, error) {
	key := r.key(biz, bizId)
	res, err := r.client.HGetAll(ctx, key).Result()
	if err != nil {
		return domain.Interactive{}, err
	}
	if len(res) == 0 {
		return domain.Interactive{}, ErrKeyNotExist
	}
	var intr domain.Interactive
	intr.BizId = bizId
	intr.CollectCnt, _ = strconv.ParseInt(res[fieldCollectCnt], 10, 64)
	intr.LikeCnt, _ = strconv.ParseInt(res[fieldLikeCnt], 10, 64)
	intr.ReadCnt, _ = strconv.ParseInt(res[fieldReadCnt], 10, 64)
	return intr, nil
}

func (r *InteractiveRedisCache) Set(ctx context.Context, biz string, bizId int64, res domain.Interactive) error {
	key := r.key(biz, bizId)
	err := r.client.HSet(ctx, key,
		fieldCollectCnt, res.CollectCnt,
		fieldReadCnt, res.ReadCnt,
		fieldLikeCnt, res.LikeCnt).Err()
	if err != nil {
		return err
	}
	return r.client.Expire(ctx, key, time.Minute*15).Err()
}

func (r *InteractiveRedisCache) IncrCollectCntIfPresent(ctx context.Context, biz string, bizId int64) error {
	key := r.key(biz, bizId)
	_, err := r.client.Eval(ctx, luaIncrCnt, []string{key}, fieldCollectCnt, 1).Int()
	return err
}

func (r *InteractiveRedisCache) IncrLikeCntIfPresent(ctx context.Context, biz string, bizId int64) error {
	key := r.key(biz, bizId)
	_, err := r.client.Eval(ctx, luaIncrCnt, []string{key}, fieldLikeCnt, 1).Int()
	return err
}

func (r *InteractiveRedisCache) DecrLikeCntIfPresent(ctx context.Context, biz string, bizId int64) error {
	key := r.key(biz, bizId)
	_, err := r.client.Eval(ctx, luaIncrCnt, []string{key}, fieldLikeCnt, -1).Int()
	return err
}

func NewInteractiveRedisCache(client redis.Cmdable) InteractiveCache {
	return &InteractiveRedisCache{
		client: client,
	}
}

func (r *InteractiveRedisCache) IncrReadCntIfPresent(ctx context.Context, biz string, bizId int64) error {
	key := r.key(biz, bizId)
	// 下面两种写法都可以
	//_, err := r.client.Eval(ctx, luaIncrCnt, []string{key}, fieldReadCnt, 1).Int()
	return r.client.Eval(ctx, luaIncrCnt, []string{key}, fieldReadCnt, 1).Err()
}

// IncrRankingIfPresentV1 分 key 的写入
func (r *InteractiveRedisCache) IncrRankingIfPresentV1(ctx context.Context, biz string, bizId int64) error {
	h := hash.Hash([]byte(biz))
	key := fmt.Sprintf("top_100_%d_%s", h%100, biz)
	res, err := r.client.Eval(ctx, luaRankingIncr, []string{key}, bizId).Result()
	if err != nil {
		return err
	}
	if res.(int64) == 0 {
		return RankingUpdateErr
	}
	return nil
}

func (r *InteractiveRedisCache) IncrRankingIfPresent(ctx context.Context, biz string, bizId int64) error {
	res, err := r.client.Eval(ctx, luaRankingIncr, []string{r.rankingKey(biz)}, bizId).Result()
	if err != nil {
		return err
	}
	if res.(int64) == 0 {
		return RankingUpdateErr
	}
	return nil
}

func (r *InteractiveRedisCache) SetRankingScore(ctx context.Context, biz string, bizId int64, count int64) error {
	return r.client.Eval(ctx, luaRankingSet, []string{r.rankingKey(biz)}, bizId, count).Err()
}

// BatchSetRankingScore 设置整个数据
func (r *InteractiveRedisCache) BatchSetRankingScore(ctx context.Context, biz string, interactives []domain.Interactive) error {
	members := make([]redis.Z, 0, len(interactives))
	for _, interactive := range interactives {
		members = append(members, redis.Z{
			Score:  float64(interactive.LikeCnt),
			Member: interactive.BizId,
		})
	}
	return r.client.ZAdd(ctx, r.rankingKey(biz), members...).Err()
}

// LikeTopV1 分 key 版本
func (r *InteractiveRedisCache) LikeTopV1(ctx context.Context, biz string) ([]domain.Interactive, error) {
	// 我从 100 个 key 里面，各取前 100
	// 然后，合并再取前 100
	interactives := make([]domain.Interactive, 0, 100*100)
	for i := 0; i < 100; i++ {
		var start int64 = 0
		var end int64 = 99
		key := fmt.Sprintf("top_100_%d_%s", i, biz)
		res, err := r.client.ZRevRangeWithScores(ctx, key, start, end).Result()
		if err != nil {
			return nil, err
		}
		for j := 0; j < len(res); j++ {
			id, _ := strconv.ParseInt(res[j].Member.(string), 10, 64)
			interactives = append(interactives, domain.Interactive{
				Biz:     biz,
				BizId:   id,
				LikeCnt: int64(res[j].Score),
			})
		}
	}

	// 进一步排序，然后取前 100
	return interactives, nil
}

func (r *InteractiveRedisCache) LikeTop(ctx context.Context, biz string) ([]domain.Interactive, error) {
	var start int64 = 0
	var end int64 = 99
	key := fmt.Sprintf("top_100_%s", biz)
	res, err := r.client.ZRevRangeWithScores(ctx, key, start, end).Result()
	if err != nil {
		return nil, err
	}
	interactives := make([]domain.Interactive, 0, 100)
	for i := 0; i < len(res); i++ {
		id, _ := strconv.ParseInt(res[i].Member.(string), 10, 64)
		interactives = append(interactives, domain.Interactive{
			Biz:     biz,
			BizId:   id,
			LikeCnt: int64(res[i].Score),
		})
	}
	return interactives, nil
}

func (r *InteractiveRedisCache) rankingKey(biz string) string {
	return fmt.Sprintf("top_100_%s", biz)
}

func (r *InteractiveRedisCache) key(biz string, bizId int64) string {
	return fmt.Sprintf("interactive:%s:%d", biz, bizId)
}
