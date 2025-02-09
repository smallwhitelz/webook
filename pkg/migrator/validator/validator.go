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
	utime     int64

	// <=0 就认为中断
	// >0 就认为睡眠
	sleepInterval time.Duration
	fromBase      func(ctx context.Context, offset int) (T, error)
}

func NewValidator[T migrator.Entity](
	base *gorm.DB,
	target *gorm.DB, l logger.V1, producer events.Producer, direction string) *Validator[T] {
	res := &Validator[T]{
		base: base, target: target,
		l: l, producer: producer, direction: direction, batchSize: 100}
	res.fromBase = res.fullFromBase
	return res
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
	offset := 0
	for {
		src, err := v.fromBase(ctx, offset)
		if err == context.DeadlineExceeded || err == context.Canceled {
			return nil
		}
		if err == gorm.ErrRecordNotFound {
			// 增量校验要考虑一直运行
			// 这个就是没有数据
			if v.sleepInterval <= 0 {
				return nil
			}
			time.Sleep(v.sleepInterval)
			continue
		}
		if err != nil {
			// 查询出错
			v.l.Error("base -> target 查询 base 失败", logger.Error(err))
			offset++
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
		offset++
	}
}

// baseToTarget 批量写法
func (v *Validator[T]) baseToTargetV1(ctx context.Context) error {
	offset := 0
	const limit = 100
	for {
		var srcs []T
		dbCtx, cancel := context.WithTimeout(ctx, time.Second)
		err := v.base.WithContext(dbCtx).Order("id").Where("utime > ?", v.utime).Offset(offset).
			Limit(limit).Find(&srcs).Error
		cancel()
		switch err {
		// 在 find 里面其实不会有这个错误
		//case gorm.ErrRecordNotFound:
		case context.DeadlineExceeded, context.Canceled:
			// 超时你可以继续，也可以返回。一般超时都是因为数据库有了问题
			return err
		case nil:
			if len(srcs) == 0 {
				// 结束，没有数据
				return nil
			}
			err := v.dstDiffV1(srcs)
			if err != nil {
				// 直接中断，你也可以考虑继续重试
				return err
			}
		default:
			v.l.Error("src => dst 查询源表失败", logger.Error(err))
		}
		if len(srcs) < limit {
			// 没有数据了
			return nil
		}
		offset += len(srcs)
	}
}

func (v *Validator[T]) dstDiffV1(srcs []T) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	ids := slice.Map(srcs, func(idx int, src T) int64 {
		return src.ID()
	})
	var dsts []T
	err := v.target.WithContext(ctx).Where("id IN ?", ids).Find(&dsts).Error
	// 让调用者来决定
	if err != nil {
		return err
	}
	dstMap := v.toMap(dsts)
	for _, src := range srcs {
		dst, ok := dstMap[src.ID()]
		if !ok {
			v.notify(src.ID(), events.InconsistentEventTypeTargetMissing)
		}
		if !src.CompareTo(dst) {
			v.notify(src.ID(), events.InconsistentEventTypeNEQ)
		}
	}
	return nil
}

func (v *Validator[T]) toMap(data []T) map[int64]T {
	res := make(map[int64]T, len(data))
	for _, val := range data {
		res[val.ID()] = val
	}
	return res
}

func (v *Validator[T]) Full() *Validator[T] {
	v.fromBase = v.fullFromBase
	return v
}

func (v *Validator[T]) Incr() *Validator[T] {
	v.fromBase = v.incrFromBase
	return v
}

func (v *Validator[T]) Utime(t int64) *Validator[T] {
	v.utime = t
	return v
}

func (v *Validator[T]) SleepInterval(interval time.Duration) *Validator[T] {
	v.sleepInterval = interval
	return v
}

func (v *Validator[T]) fullFromBase(ctx context.Context, offset int) (T, error) {
	dbCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	var src T
	err := v.base.WithContext(dbCtx).Order("id").Offset(offset).First(&src).Error
	return src, err
}

func (v *Validator[T]) incrFromBase(ctx context.Context, offset int) (T, error) {
	dbCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	var src T
	err := v.base.WithContext(dbCtx).Where("utime > ?", v.utime).
		Order("utime").Offset(offset).First(&src).Error
	return src, err
}

func (v *Validator[T]) validateTargetToBase(ctx context.Context) error {
	offset := 0
	for {
		var ts []T
		err := v.target.WithContext(ctx).Select("id").Order("id").
			Offset(offset).Limit(v.batchSize).Find(&ts).Error
		if err == context.DeadlineExceeded || err == context.Canceled {
			return nil
		}
		if err == gorm.ErrRecordNotFound || len(ts) == 0 {
			if v.sleepInterval <= 0 {
				return nil
			}
			time.Sleep(v.sleepInterval)
			continue
		}
		if err != nil {
			v.l.Error("target -> base 查询 target 失败", logger.Error(err))
			offset += len(ts)
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
			offset += len(ts)
			continue
		}
		// 找差集，diff里面就是target有，但是base没有的
		diff := slice.DiffSetFunc(ts, srcTs, func(src, dst T) bool {
			return src.ID() == dst.ID()
		})
		v.notifyBaseMissing(diff)
		// 说明也没有数据了
		if len(ts) < v.batchSize {
			if v.sleepInterval <= 0 {
				return nil
			}
			time.Sleep(v.sleepInterval)
		}
		offset += len(ts)
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
	if err != nil {
		v.l.Error("发送不一致消息失败",
			logger.Error(err),
			logger.String("type", typ),
			logger.Int64("id", id))
	}
}
