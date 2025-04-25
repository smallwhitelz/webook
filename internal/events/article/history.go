package article

import (
	"context"
	"github.com/IBM/sarama"
	"time"
	"webook/internal/domain"
	"webook/internal/repository"
	"webook/pkg/logger"
	"webook/pkg/saramax"
)

type HistoryRecordConsumer struct {
	repo   repository.HistoryRecordRepository
	client sarama.Client
	l      logger.V1
}

func (h *HistoryRecordConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("interactive", h.client)
	if err != nil {
		return err
	}
	go func() {
		er := cg.Consume(context.Background(), []string{TopicReadEvent}, saramax.NewHandler[ReadEvent](h.l, h.Consume))
		if er != nil {
			h.l.Error("退出消费", logger.Error(er))
		}
	}()
	return err
}

func (h *HistoryRecordConsumer) Consume(msg *sarama.ConsumerMessage, event ReadEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return h.repo.AddRecord(ctx, domain.HistoryRecord{
		BizId: event.Aid,
		Biz:   "article",
		Uid:   event.Uid,
	})
}
