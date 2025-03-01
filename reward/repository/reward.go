package repository

import (
	"context"
	"webook/reward/domain"
	"webook/reward/repository/cache"
	"webook/reward/repository/dao"
)

type rewardRepository struct {
	dao   dao.RewardDAO
	cache cache.RewardCache
}

func (repo *rewardRepository) CachedCodeURL(ctx context.Context, cu domain.CodeURL, r domain.Reward) error {
	return repo.cache.CachedCodeURL(ctx, cu, r)
}

func (repo *rewardRepository) GetCachedCodeURL(ctx context.Context, reward domain.Reward) (domain.CodeURL, error) {
	return repo.cache.GetCachedCodeURL(ctx, reward)
}

func NewRewardRepository(dao dao.RewardDAO, c cache.RewardCache) RewardRepository {
	return &rewardRepository{dao: dao, cache: c}
}

func (repo *rewardRepository) CreateReward(ctx context.Context, reward domain.Reward) (int64, error) {
	return repo.dao.Insert(ctx, repo.toEntity(reward))
}

func (repo *rewardRepository) toEntity(r domain.Reward) dao.Reward {
	return dao.Reward{
		Status:    r.Status.AsUint8(),
		Biz:       r.Target.Biz,
		BizName:   r.Target.BizName,
		BizId:     r.Target.BizId,
		TargetUid: r.Target.Uid,
		Uid:       r.Uid,
		Amount:    r.Amt,
	}
}

func (repo *rewardRepository) toDomain(r dao.Reward) domain.Reward {
	return domain.Reward{
		Id:  r.Id,
		Uid: r.Uid,
		Target: domain.Target{
			Biz:     r.Biz,
			BizId:   r.BizId,
			BizName: r.BizName,
			Uid:     r.Uid,
		},
		Amt:    r.Amount,
		Status: domain.RewardStatus(r.Status),
	}
}
