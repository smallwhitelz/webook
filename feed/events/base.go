package events

import (
	"context"
	"github.com/IBM/sarama"
	"time"
	"webook/feed/domain"
	"webook/feed/service"
	"webook/pkg/logger"
	"webook/pkg/saramax"
)

const topicFeedEvent = "feed_event"

// FeedEvent 业务方就按照这个格式，将放到Feed里面的数据，丢到feed_event这个topic下
type FeedEvent struct {
	Type string
	// 一定是string
	// map[string]any
	// 传过来的是int64，再反解析回来，就变成float64
	Metadata map[string]string
}

type FeedEventConsumer struct {
	client sarama.Client
	l      logger.V1
	svc    service.FeedService
}

func NewFeedEventConsumer(client sarama.Client, l logger.V1, svc service.FeedService) *FeedEventConsumer {
	return &FeedEventConsumer{client: client, l: l, svc: svc}
}

func (r *FeedEventConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("feed_event", r.client)
	if err != nil {
		return err
	}
	go func() {
		err := cg.Consume(context.Background(), []string{topicFeedEvent}, saramax.NewHandler[FeedEvent](r.l, r.Consume))
		if err != nil {
			r.l.Error("退出消费循环一场", logger.Error(err))
		}
	}()
	return err
}

func (r *FeedEventConsumer) Consume(msg *sarama.ConsumerMessage, evt FeedEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return r.svc.CreateFeedEvent(ctx, domain.FeedEvent{
		Type: evt.Type,
		Ext:  evt.Metadata,
	})
}
