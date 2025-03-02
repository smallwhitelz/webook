package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
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
	// 快路径
	res, err := w.repo.GetReward(ctx, rid)
	if err != nil {
		return domain.Reward{}, err
	}
	// 确保是自己打赏的
	if res.Uid != uid {
		return domain.Reward{}, errors.New("非法访问别人的打赏记录")
	}

	// 触发降级或者限流的时候，不走慢路径
	if ctx.Value("limited") == "true" {
		return res, nil
	}
	if !res.Completed() {
		// 我去问一下，有可能支付那边已经处理好了，已经收到回调
		pmtRes, err := w.client.GetPayment(ctx, &pmtv1.GetPaymentRequest{
			BizTradeNo: w.bizTradeNO(rid),
		})
		if err != nil {
			w.l.Error("慢路径查询支付状态失败", logger.Error(err),
				logger.Int64("rid", rid))
			return res, nil
		}
		switch pmtRes.Status {
		case pmtv1.PaymentStatus_PaymentStatusSuccess:
			res.Status = domain.RewardStatusPayed
		case pmtv1.PaymentStatus_PaymentStatusInit:
			res.Status = domain.RewardStatusInit
		case pmtv1.PaymentStatus_PaymentStatusRefund:
			res.Status = domain.RewardStatusFailed
		case pmtv1.PaymentStatus_PaymentStatusFailed:
			res.Status = domain.RewardStatusFailed
		case pmtv1.PaymentStatus_PaymentStatusUnknown:
		}
		err = w.UpdateReward(ctx, w.bizTradeNO(rid), res.Status)
		if err != nil {
			w.l.Error("慢路径更新本地状态失败", logger.Error(err),
				logger.Int64("rid", rid))
		}
	}
	return res, nil
}

func (w *WechatNativeRewardService) UpdateReward(ctx context.Context, bizTradeNO string, status domain.RewardStatus) error {
	rid := w.toRid(bizTradeNO)
	return w.repo.UpdateStatus(ctx, rid, status)
}

func (s *WechatNativeRewardService) bizTradeNO(rid int64) string {
	return fmt.Sprintf("reward-%d", rid)
}

func (w *WechatNativeRewardService) toRid(tradeNO string) int64 {
	ridStr := strings.Split(tradeNO, "-")
	val, _ := strconv.ParseInt(ridStr[1], 10, 64)
	return val
}
