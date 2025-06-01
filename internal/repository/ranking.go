package repository

import (
	"context"
	"webook/internal/domain"
	"webook/internal/repository/cache"
)

type RankingRepository interface {
	// ReplaceTopN 将热榜数据存入缓存
	ReplaceTopN(ctx context.Context, arts []domain.Article) error
	GetTopN(ctx context.Context) ([]domain.Article, error)
}

type CachedRankingRepository struct {
	cache cache.RankingCache

	// 下面是给 v1 用的
	redisCache *cache.RankingRedisCache
	localCache *cache.RankingLocalCache
}

// NewCachedRankingRepositoryV1 结合本地缓存
// 为什么结合本地缓存？
// 因为像热榜这类需求，对于所有用户必然都会调用这个接口
// 所以双缓存更安全，更有效
// 查找的时候，本地缓存->redis->数据库
// 更新的时候，数据库->本地缓存->redis
// 核心在于本地缓存操作几乎不可能失败
func NewCachedRankingRepositoryV1(redisCache *cache.RankingRedisCache, localCache *cache.RankingLocalCache) *CachedRankingRepository {
	return &CachedRankingRepository{redisCache: redisCache, localCache: localCache}
}

func (c *CachedRankingRepository) GetTopNV1(ctx context.Context) ([]domain.Article, error) {
	res, err := c.localCache.Get(ctx)
	if err == nil {
		return res, nil
	}
	res, err = c.redisCache.Get(ctx)
	if err != nil {
		// 解决可用性问题，设置一个兜底，当redis崩溃时，再去本地缓存获取数据
		// 这个时候就不要考虑过期时间，直接拿数据，因为有数据总比没数据的好
		return c.localCache.ForceGet(ctx)
	}
	// 回写缓存
	_ = c.localCache.Set(ctx, res)
	return res, nil
}

func (c *CachedRankingRepository) GetTopN(ctx context.Context) ([]domain.Article, error) {
	return c.cache.Get(ctx)
}

func NewCachedRankingRepository(cache cache.RankingCache) RankingRepository {
	return &CachedRankingRepository{cache: cache}
}

// ReplaceTopNV1 在本地和redis都缓存
func (c *CachedRankingRepository) ReplaceTopNV1(ctx context.Context, arts []domain.Article) error {
	_ = c.localCache.Set(ctx, arts)
	return c.redisCache.Set(ctx, arts)
}

// ReplaceTopN 只在redis缓存
func (c *CachedRankingRepository) ReplaceTopN(ctx context.Context, arts []domain.Article) error {
	return c.cache.Set(ctx, arts)
}
