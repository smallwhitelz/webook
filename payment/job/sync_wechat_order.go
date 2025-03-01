package job

import (
	"context"
	"time"
	"webook/payment/services/wechat"
	"webook/pkg/logger"
)

type SyncWechatOrderJob struct {
	svc *wechat.NativePaymentService
	l   logger.V1
}

func (s *SyncWechatOrderJob) Name() string {
	return "sync_wechat_order_job"
}

// 这个定时任务多久运行一次？
// 不必特别频繁，一分钟一次就可以
func (s *SyncWechatOrderJob) Run() error {
	// 定时找到超时的微信支付订单，然后发起同步
	// 针对过期订单
	t := time.Now().Add(-time.Minute * 31)
	offset := 0
	const limit = 100
	for {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		pmts, err := s.svc.FindExpiredPayment(ctx, offset, limit, t)
		cancel()
		if err != nil {
			return err
		}
		for _, pmt := range pmts {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
			err = s.svc.SyncWechatInfo(ctx, pmt.BizTradeNO)
			cancel()
			if err != nil {
				s.l.Error("同步微信订单状态失败", logger.Error(err),
					logger.String("biz_trade_no", pmt.BizTradeNO))
			}
		}
		if len(pmts) < limit {
			return nil
		}
		offset = offset + limit
	}
}
