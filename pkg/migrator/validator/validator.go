package validator

import (
	"context"
	"github.com/ecodeclub/ekit/slice"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
	"time"
	"webook/pkg/logger"
	"webook/pkg/migrator"
	"webook/pkg/migrator/events"
)

type Validator[T migrator.Entity] struct {
	// 数据迁移，肯定有
	base      *gorm.DB
	target    *gorm.DB
	l         logger.V1
	producer  events.Producer
	direction string
	batchSize int
}

func (v *Validator[T]) Validate(ctx context.Context) error {
	// 同步写法
	//err := v.validateBaseToTarget(ctx)
	//if err != nil {
	//	return err
	//}
	//return v.validateTargetToBase(ctx)

	// 异步写法
	var eg errgroup.Group
	eg.Go(func() error {
		return v.validateBaseToTarget(ctx)
	})
	eg.Go(func() error {
		return v.validateTargetToBase(ctx)
	})
	return eg.Wait()
}

func (v *Validator[T]) validateBaseToTarget(ctx context.Context) error {
	offset := -1
	for {
		offset++
		var src T
		err := v.base.WithContext(ctx).Order("id").Offset(offset).First(&src).Error
		if err == gorm.ErrRecordNotFound {
			// 这个就是没有数据
			return nil
		}
		if err != nil {
			// 查询出错
			v.l.Error("base -> target 查询 base 失败", logger.Error(err))
			continue
		}
		// 这边就是正常情况
		var dst T
		err = v.target.WithContext(ctx).Where("id = ?", src.ID()).First(&dst).Error
		switch err {
		case gorm.ErrRecordNotFound:
			// target没有
			// 丢一条消息到Kafka上
			v.notify(src.ID(), events.InconsistentEventTypeTargetMissing)
		case nil:
			equal := src.CompareTo(dst)
			if !equal {
				// 要丢一条消息到Kafka上
				v.notify(src.ID(), events.InconsistentEventTypeNEQ)
			}
		default:
			// 记录日志，然后继续
			v.l.Error("base -> target 查询 target 失败",
				logger.Int64("id", src.ID()),
				logger.Error(err))
		}
	}
}

func (v *Validator[T]) validateTargetToBase(ctx context.Context) error {
	offset := -v.batchSize
	for {
		offset += v.batchSize
		var ts []T
		err := v.target.WithContext(ctx).Select("id").Order("id").
			Offset(offset).Limit(v.batchSize).Find(&ts).Error
		if err == gorm.ErrRecordNotFound || len(ts) == 0 {
			return nil
		}
		if err != nil {
			v.l.Error("target -> base 查询 target 失败", logger.Error(err))
			continue
		}
		// 在这里有数据
		var srcTs []T
		ids := slice.Map(ts, func(idx int, t T) int64 {
			return t.ID()
		})
		err = v.base.WithContext(ctx).Select("id").Where("id IN ?", ids).Find(&srcTs).Error
		if err == gorm.ErrRecordNotFound || len(srcTs) == 0 {
			// 代表base里面一条对应的数据都没有
			v.notifyBaseMissing(ts)
			continue
		}
		if err != nil {
			v.l.Error("target -> base 查询 base 失败", logger.Error(err))
			continue
		}
		// 找差集，diff里面就是target有，但是base没有的
		diff := slice.DiffSetFunc(ts, srcTs, func(src, dst T) bool {
			return src.ID() == dst.ID()
		})
		v.notifyBaseMissing(diff)
		// 说明也没有数据了
		if len(ts) < v.batchSize {
			return nil
		}
	}
}

func (v *Validator[T]) notifyBaseMissing(ts []T) {
	for _, val := range ts {
		v.notify(val.ID(), events.InconsistentEventTypeBaseMissing)
	}
}

func (v *Validator[T]) notify(id int64, typ string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err := v.producer.ProduceInconsistentEvent(ctx, events.InconsistentEvent{
		ID:        id,
		Type:      typ,
		Direction: v.direction,
	})
	v.l.Error("发送不一致消息失败",
		logger.Error(err),
		logger.String("type", typ),
		logger.Int64("id", id))
}
