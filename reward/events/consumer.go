package events

import (
	"context"
	"github.com/IBM/sarama"
	"strings"
	"time"
	"webook/pkg/logger"
	"webook/pkg/saramax"
	"webook/reward/domain"
	"webook/reward/service"
)

type PaymentEvent struct {
	BizTradeNo string
	Status     uint8
}

func (p PaymentEvent) ToDomainStatus() domain.RewardStatus {
	// 	PaymentStatusInit
	//	PaymentStatusSuccess
	//	PaymentStatusFailed
	//	PaymentStatusRefund
	switch p.Status {
	// 这里不能引用 payment 里面的定义，只能手写
	case 1:
		return domain.RewardStatusInit
	case 2:
		return domain.RewardStatusPayed
	case 3, 4:
		return domain.RewardStatusFailed
	default:
		return domain.RewardStatusUnknown
	}
}

type PaymentEventConsumer struct {
	client sarama.Client
	l      logger.V1
	svc    service.RewardService
}

// Start 这边就是自己启动 goroutine 了
func (r *PaymentEventConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("reward",
		r.client)
	if err != nil {
		return err
	}
	go func() {
		err := cg.Consume(context.Background(),
			[]string{"payment_events"},
			saramax.NewHandler[PaymentEvent](r.l, r.Consume))
		if err != nil {
			r.l.Error("退出了消费循环异常", logger.Error(err))
		}
	}()
	return err
}

func (r *PaymentEventConsumer) Consume(msg *sarama.ConsumerMessage, evt PaymentEvent) error {
	if !strings.HasPrefix(evt.BizTradeNo, "reward") {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	return r.svc.UpdateReward(ctx, evt.BizTradeNo, evt.ToDomainStatus())
}
