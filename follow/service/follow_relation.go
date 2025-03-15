package service

import (
	"context"
	"webook/follow/domain"
	"webook/follow/repository"
)

type FollowRelationService interface {
	// GetFollowee 获取关注列表
	GetFollowee(ctx context.Context, follower, offset, limit int64) ([]domain.FollowRelation, error)
	// FollowInfo 点到文章页面，看作者是否被关注的接口
	FollowInfo(ctx context.Context,
		follower, followee int64) (domain.FollowRelation, error)
	// Follow 关注
	Follow(ctx context.Context, follower, followee int64) error
	// CancelFollow 取消关注
	CancelFollow(ctx context.Context, follower, followee int64) error
}

type followRelationService struct {
	repo repository.FollowRepository
}

func (f *followRelationService) GetFollowee(ctx context.Context, follower, offset, limit int64) ([]domain.FollowRelation, error) {
	return f.repo.GetFollowee(ctx, follower, offset, limit)
}

func (f *followRelationService) FollowInfo(ctx context.Context, follower, followee int64) (domain.FollowRelation, error) {
	val, err := f.repo.FollowInfo(ctx, follower, followee)
	return val, err
}

func (f *followRelationService) Follow(ctx context.Context, follower, followee int64) error {
	return f.repo.AddFollowRelation(ctx, domain.FollowRelation{
		Followee: followee,
		Follower: follower,
	})
}

func (f *followRelationService) CancelFollow(ctx context.Context, follower, followee int64) error {
	return f.repo.InactiveFollowRelation(ctx, follower, followee)
}

func NewFollowRelationService(repo repository.FollowRepository) FollowRelationService {
	return &followRelationService{repo: repo}
}
