package repository

import (
	"context"
	"webook/follow/domain"
	"webook/follow/repository/cache"
	"webook/follow/repository/dao"
	"webook/pkg/logger"
)

type FollowRepository interface {
	// AddFollowRelation 创建关注关系
	AddFollowRelation(ctx context.Context, f domain.FollowRelation) error
	// InactiveFollowRelation 取消关注
	InactiveFollowRelation(ctx context.Context, follower int64, followee int64) error
	GetFollowee(ctx context.Context, follower int64, offset int64, limit int64) ([]domain.FollowRelation, error)
	FollowInfo(ctx context.Context, follower int64, followee int64) (domain.FollowRelation, error)
}

type CachedRelationRepository struct {
	dao   dao.FollowRelationDAO
	cache cache.FollowCache
	l     logger.V1
}

func (c *CachedRelationRepository) FollowInfo(ctx context.Context, follower int64, followee int64) (domain.FollowRelation, error) {
	fr, err := c.dao.FollowRelationDetail(ctx, follower, followee)
	if err != nil {
		return domain.FollowRelation{}, err
	}
	return c.toDomain(fr), nil
}

func (c *CachedRelationRepository) GetFollowee(ctx context.Context, follower int64, offset int64, limit int64) ([]domain.FollowRelation, error) {
	// 你可以考虑在这里缓存关注者列表的第一页
	followerList, err := c.dao.FollowRelationList(ctx, follower, offset, limit)
	if err != nil {
		return nil, err
	}
	return c.genFollowRelationList(followerList), nil
}

func (c *CachedRelationRepository) InactiveFollowRelation(ctx context.Context, follower int64, followee int64) error {
	err := c.dao.UpdateStatus(ctx, follower, followee, dao.FollowRelationStatusInactive)
	if err != nil {
		return err
	}
	return c.cache.CancelFollow(ctx, follower, followee)
}

func (c *CachedRelationRepository) AddFollowRelation(ctx context.Context, f domain.FollowRelation) error {
	err := c.dao.CreateFollowRelation(ctx, c.toEntity(f))
	if err != nil {
		return err
	}
	// 更新缓存里面的关注了多少人，以及有多少粉丝的计数， +1
	return c.cache.Follow(ctx, f.Follower, f.Followee)
}

func (c *CachedRelationRepository) toEntity(f domain.FollowRelation) dao.FollowRelation {
	return dao.FollowRelation{
		Follower: f.Follower,
		Followee: f.Followee,
	}
}

func (c *CachedRelationRepository) genFollowRelationList(followerList []dao.FollowRelation) []domain.FollowRelation {
	res := make([]domain.FollowRelation, 0, len(followerList))
	for _, f := range followerList {
		res = append(res, c.toDomain(f))
	}
	return res
}

func (c *CachedRelationRepository) toDomain(f dao.FollowRelation) domain.FollowRelation {
	return domain.FollowRelation{
		Followee: f.Followee,
		Follower: f.Follower,
	}
}

func NewCachedRelationRepository(dao dao.FollowRelationDAO, cache cache.FollowCache, l logger.V1) FollowRepository {
	return &CachedRelationRepository{dao: dao, cache: cache, l: l}
}
