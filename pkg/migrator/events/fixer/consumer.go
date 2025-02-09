package fixer

import (
	"context"
	"errors"
	"github.com/IBM/sarama"
	"gorm.io/gorm"
	"time"
	"webook/pkg/logger"
	"webook/pkg/migrator"
	"webook/pkg/migrator/events"
	"webook/pkg/migrator/fixer"
	"webook/pkg/samarax"
)

type Consumer[T migrator.Entity] struct {
	client   sarama.Client
	l        logger.V1
	srcFirst *fixer.OverrideFixer[T]
	dstFirst *fixer.OverrideFixer[T]
	topic    string
}

func NewConsumer[T migrator.Entity](
	client sarama.Client,
	l logger.V1, src *gorm.DB,
	dst *gorm.DB, topic string) (*Consumer[T], error) {
	srcFirst, err := fixer.NewOverrideFixer[T](src, dst)
	if err != nil {
		return nil, err
	}
	dstFirst, err := fixer.NewOverrideFixer[T](src, dst)
	if err != nil {

	}
	return &Consumer[T]{
		client:   client,
		l:        l,
		srcFirst: srcFirst,
		dstFirst: dstFirst,
		topic:    topic}, nil
}

func (c *Consumer[T]) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("migrator-fix", c.client)
	if err != nil {
		return err
	}
	go func() {
		er := cg.Consume(context.Background(), []string{c.topic}, samarax.NewHandler[events.InconsistentEvent](c.l, c.Consume))
		// 带监控
		//er := cg.Consume(context.Background(), []string{TopicReadEvent}, samarax.NewHandlerV1[ReadEvent]("consumer_prom", i.l, i.Consume))
		if er != nil {
			c.l.Error("退出消费", logger.Error(er))
		}
	}()
	return err
}

func (c *Consumer[T]) Consume(msg *sarama.ConsumerMessage, t events.InconsistentEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	switch t.Direction {
	case "SRC":
		return c.srcFirst.Fix(ctx, t.ID)
	case "DST":
		return c.dstFirst.Fix(ctx, t.ID)
	}
	return errors.New("未知的校验方向")
}
