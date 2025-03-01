package service

import (
	"context"
	"fmt"
	pmtv1 "webook/api/proto/gen/payment/v1"
	"webook/pkg/logger"
	"webook/reward/domain"
	"webook/reward/repository"
)

type WechatNativeRewardService struct {
	client pmtv1.WechatPaymentServiceClient
	repo   repository.RewardRepository
	l      logger.V1
}

func (w *WechatNativeRewardService) PreReward(ctx context.Context, r domain.Reward) (domain.CodeURL, error) {
	// 缓存，可选的步骤
	res, err := w.repo.GetCachedCodeURL(ctx, r)
	if err == nil {
		return res, nil
	}
	r.Status = domain.RewardStatusInit
	rid, err := w.repo.CreateReward(ctx, r)
	if err != nil {
		return domain.CodeURL{}, err
	}
	pmtResp, err := w.client.NativePrePay(ctx, &pmtv1.PrePayRequest{
		Amt: &pmtv1.Amount{
			Total:    r.Amt,
			Currency: "CNY",
		},
		BizTradeNo:  fmt.Sprintf("reward-%d", rid),
		Description: fmt.Sprintf("打赏-%s", r.Target.BizName),
	})
	if err != nil {
		return domain.CodeURL{}, err
	}
	cu := domain.CodeURL{
		Rid: rid,
		URL: pmtResp.CodeUrl,
	}
	err1 := w.repo.CachedCodeURL(ctx, cu, r)
	if err1 != nil {
		w.l.Error("缓存二维码失败", logger.Error(err1), logger.Int64("rid", rid))
	}
	return cu, nil
}

func (w *WechatNativeRewardService) GetReward(ctx context.Context, rid, uid int64) (domain.Reward, error) {
	//TODO implement me
	panic("implement me")
}

func (w *WechatNativeRewardService) UpdateReward(ctx context.Context, bizTradeNO string, status domain.RewardStatus) error {
	//TODO implement me
	panic("implement me")
}
