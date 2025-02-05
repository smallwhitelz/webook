package samarax

import (
	"encoding/json"
	"github.com/IBM/sarama"
	"github.com/prometheus/client_golang/prometheus"
	"strconv"
	"time"
	"webook/pkg/logger"
)

type HandlerV1[T any] struct {
	l      logger.V1
	fn     func(msg *sarama.ConsumerMessage, event T) error
	vector *prometheus.SummaryVec
}

func NewHandlerV1[T any](
	consumer string,
	l logger.V1,
	fn func(msg *sarama.ConsumerMessage, event T) error) *HandlerV1[T] {
	// 整个 options 都可以考虑做成参数
	vector := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: "saramax",
		Subsystem: "consumer_handler",
		Name:      consumer,
	}, []string{"topic", "error"})
	return &HandlerV1[T]{l: l, fn: fn, vector: vector}
}

func (h *HandlerV1[T]) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (h *HandlerV1[T]) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (h *HandlerV1[T]) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	msgs := claim.Messages()
	for msg := range msgs {
		// 在这里调用业务处理逻辑

		h.consumeClaim(msg)
		session.MarkMessage(msg, "")
	}
	return nil
}

func (h *HandlerV1[T]) consumeClaim(msg *sarama.ConsumerMessage) {
	start := time.Now()
	var err error
	defer func() {
		errInfo := strconv.FormatBool(err != nil)
		duration := time.Since(start).Milliseconds()
		h.vector.WithLabelValues(msg.Topic, errInfo).Observe(float64(duration))
	}()
	var t T
	err = json.Unmarshal(msg.Value, &t)
	if err != nil {
		// 你也可以在这里引入重试的逻辑
		h.l.Error("反序列消息体失败",
			logger.String("topic", msg.Topic),
			logger.Int32("partition", msg.Partition),
			logger.Int64("offset", msg.Offset),
			logger.Error(err))
	}
	err = h.fn(msg, t)
	if err != nil {
		h.l.Error("处理消息失败",
			logger.String("topic", msg.Topic),
			logger.Int32("partition", msg.Partition),
			logger.Int64("offset", msg.Offset),
			logger.Error(err))
	}
}
