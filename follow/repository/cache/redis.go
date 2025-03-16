package cache

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"strconv"
	"webook/follow/domain"
)

var ErrKeyNotExist = redis.Nil

type RedisFollowCache struct {
	client redis.Cmdable
}

func (r *RedisFollowCache) StaticsInfo(ctx context.Context, uid int64) (domain.FollowStatics, error) {
	data, err := r.client.HGetAll(ctx, r.staticsKey(uid)).Result()
	if err != nil {
		return domain.FollowStatics{}, err
	}
	if len(data) == 0 {
		return domain.FollowStatics{}, ErrKeyNotExist
	}
	var res domain.FollowStatics
	res.Followers, _ = strconv.ParseInt(data[fieldFollowerCnt], 10, 64)
	res.Followees, _ = strconv.ParseInt(data[fieldFolloweeCnt], 10, 64)
	return res, nil
}

func (r *RedisFollowCache) SetStaticsInfo(ctx context.Context, uid int64, statics domain.FollowStatics) error {
	return r.client.HMSet(ctx, r.staticsKey(uid), fieldFollowerCnt, statics.Followers, fieldFolloweeCnt, statics.Followees).Err()
}

func (r *RedisFollowCache) Follow(ctx context.Context, follower, followee int64) error {
	return r.updateStaticsInfo(ctx, follower, followee, 1)
}

func (r *RedisFollowCache) CancelFollow(ctx context.Context, follower, followee int64) error {
	return r.updateStaticsInfo(ctx, follower, followee, -1)
}

func (r *RedisFollowCache) staticsKey(uid int64) string {
	return fmt.Sprintf("follow:statics:%d", uid)
}

func (r *RedisFollowCache) updateStaticsInfo(ctx context.Context, follower int64, followee int64, delta int64) error {
	tx := r.client.TxPipeline()
	// 这两个操作只是记录一下，还没有发过去redis
	tx.HIncrBy(ctx, r.staticsKey(follower), fieldFolloweeCnt, delta)
	tx.HIncrBy(ctx, r.staticsKey(followee), fieldFollowerCnt, delta)
	// 发过去redis执行，并且返回了结果
	_, err := tx.Exec(ctx)
	return err
}

func NewRedisFollowCache(client redis.Cmdable) FollowCache {
	return &RedisFollowCache{client: client}
}

const (
	// 被多少人关注
	fieldFollowerCnt = "follower_cnt"
	// 关注了多少人
	fieldFolloweeCnt = "followee_cnt"
)
