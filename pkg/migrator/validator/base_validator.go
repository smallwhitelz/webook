package validator

import (
	"context"
	"gorm.io/gorm"
	"time"
	"webook/pkg/logger"
	"webook/pkg/migrator/events"
)

type baseValidator struct {
	base   *gorm.DB
	target *gorm.DB
	// 这边需要告知，是以 SRC 为准，还是以 DST 为准
	// 修复数据需要知道
	direction string
	l         logger.V1
	producer  events.Producer
}

func (v *baseValidator) notify(id int64, typ string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	evt := events.InconsistentEvent{
		Direction: v.direction,
		ID:        id,
		Type:      typ,
	}
	err := v.producer.ProduceInconsistentEvent(ctx, evt)
	if err != nil {
		v.l.Error("发送消息失败", logger.Error(err),
			logger.Field{Key: "event", Val: evt})
	}
}
