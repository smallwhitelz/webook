package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
	"webook/internal/domain"
)

type ArticleCache interface {
	GetFirstPage(ctx context.Context, uid int64) ([]domain.Article, error)
	SetFirstPage(ctx context.Context, uid int64, arts []domain.Article) error
	DelFirstPage(ctx context.Context, uid int64) error
	Get(ctx context.Context, id int64) (domain.Article, error)
	Set(ctx context.Context, art domain.Article) error
	GetPub(ctx context.Context, id int64) (domain.Article, error)
	SetPub(ctx context.Context, res domain.Article) error
}

type ArticleRedisCache struct {
	client redis.Cmdable
}

func (a *ArticleRedisCache) Get(ctx context.Context, id int64) (domain.Article, error) {
	val, err := a.client.Get(ctx, a.key(id)).Bytes()
	if err != nil {
		return domain.Article{}, err
	}
	var res domain.Article
	err = json.Unmarshal(val, &res)
	return res, err
}

func (a *ArticleRedisCache) Set(ctx context.Context, art domain.Article) error {
	val, err := json.Marshal(art)
	if err != nil {
		return err
	}
	return a.client.Set(ctx, a.key(art.Id), val, time.Minute*10).Err()
}

func (a *ArticleRedisCache) GetPub(ctx context.Context, id int64) (domain.Article, error) {
	val, err := a.client.Get(ctx, a.pubKey(id)).Bytes()
	if err != nil {
		return domain.Article{}, err
	}
	var res domain.Article
	err = json.Unmarshal(val, &res)
	return res, err
}

func (a *ArticleRedisCache) SetPub(ctx context.Context, art domain.Article) error {
	val, err := json.Marshal(art)
	if err != nil {
		return err
	}
	return a.client.Set(ctx, a.pubKey(art.Id), val, time.Minute*10).Err()
}

func (a *ArticleRedisCache) GetFirstPage(ctx context.Context, uid int64) ([]domain.Article, error) {
	key := a.firstKey(uid)
	// 返回string
	//val, err := a.client.Get(ctx, firstKey).Result()
	// 返回byte数组
	val, err := a.client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}
	var res []domain.Article
	err = json.Unmarshal(val, &res)
	return res, err
}

func (a *ArticleRedisCache) SetFirstPage(ctx context.Context, uid int64, arts []domain.Article) error {
	// 这里只缓存摘要，列表显示没必要显示内容
	for i := 0; i < len(arts); i++ {
		arts[i].Content = arts[i].Abstract()
	}
	key := a.firstKey(uid)
	val, err := json.Marshal(arts)
	if err != nil {
		return err
	}
	return a.client.Set(ctx, key, val, time.Minute*10).Err()

}

func (a *ArticleRedisCache) DelFirstPage(ctx context.Context, uid int64) error {
	key := a.firstKey(uid)
	return a.client.Del(ctx, key).Err()
}

func NewArticleRedisCache(client redis.Cmdable) ArticleCache {
	return &ArticleRedisCache{
		client: client,
	}
}

func (a *ArticleRedisCache) pubKey(id int64) string {
	return fmt.Sprintf("article:pub:detail:%d", id)
}

func (a *ArticleRedisCache) key(id int64) string {
	return fmt.Sprintf("article:detail:%d", id)
}

func (a *ArticleRedisCache) firstKey(uid int64) string {
	return fmt.Sprintf("article:first_page:%d", uid)
}
