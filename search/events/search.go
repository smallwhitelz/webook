package events

import (
	"context"
	"github.com/IBM/sarama"
	"time"
	"webook/pkg/logger"
	"webook/pkg/samarax"
	"webook/search/service"
)

// SyncDataEvent 通用的event
// 所有的业务方都可以通过这个event 来同步数据
type SyncDataEvent struct {
	IndexName string
	DocID     string
	// 这里应该是BizTags
	Data string
}
type SyncDataEventConsumer struct {
	svc    service.SyncService
	client sarama.Client
	l      logger.V1
}

func (s *SyncDataEventConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("search_sync_data", s.client)
	if err != nil {
		return err
	}
	go func() {
		err := cg.Consume(context.Background(), []string{"sync_search_data"},
			samarax.NewHandler[SyncDataEvent](s.l, s.Consume))
		if err != nil {
			s.l.Error("退出消费循环异常", logger.Error(err))
		}
	}()
	return err
}

func (s *SyncDataEventConsumer) Consume(sg *sarama.ConsumerMessage, evt SyncDataEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return s.svc.InputAny(ctx, evt.IndexName, evt.DocID, evt.Data)
}
