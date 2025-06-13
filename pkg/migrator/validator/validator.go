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
	// 数据迁移，肯定有源头和目标
	base     *gorm.DB
	target   *gorm.DB
	l        logger.LoggerV1
	producer events.Producer

	direction string
	// 用在 target->base批量查询
	batchSize int

	// 用在增量校验与修复的字段
	// 从某个修改时间开始进行增量修复和校验
	utime int64
	// <= 0就认为中断
	// > 0 就认为睡眠
	sleepInterval time.Duration

	// 因为全量和增量的条件不同，所以这里抽出一个公共方法去解决
	// 用于源表到目标表的校验
	formBase func(ctx context.Context, offset int) (T, error)
}

func NewValidator[T migrator.Entity](base *gorm.DB, target *gorm.DB, l logger.LoggerV1,
	producer events.Producer, direction string) *Validator[T] {
	res := &Validator[T]{
		base:      base,
		target:    target,
		l:         l,
		producer:  producer,
		direction: direction,
		batchSize: 100,
	}
	res.formBase = res.fullFromBase
	return res
}

// Validate 执行校验
func (v *Validator[T]) Validate(ctx context.Context) error {
	// 同步写法
	//err := v.validateBaseToTarget(ctx)
	//if err != nil {
	//	return err
	//}
	//return v.validateTargetToBase(ctx)
	// 并发
	var eg errgroup.Group
	eg.Go(func() error {
		return v.validateBaseToTarget(ctx)
	})
	eg.Go(func() error {
		return v.validateTargetToBase(ctx)
	})
	return eg.Wait()
}

// Full 暴露给业务实现全量校验的方法
func (v *Validator[T]) Full() *Validator[T] {
	v.formBase = v.fullFromBase
	return v
}

// Incr 暴露给业务实现增量校验的方法
func (v *Validator[T]) Incr() *Validator[T] {
	v.formBase = v.incrFromBase
	return v
}

// Utime 和 SleepInterval 暴露Utime和SleepInterval来控制全量还是增量
func (v *Validator[T]) Utime(t int64) *Validator[T] {
	v.utime = t
	return v
}

func (v *Validator[T]) SleepInterval(interval time.Duration) *Validator[T] {
	v.sleepInterval = interval
	return v
}

// fullFromBase 全量校验，用于源表到目标表的校验
func (v *Validator[T]) fullFromBase(ctx context.Context, offset int) (T, error) {
	dbCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	var src T
	// 没有引入utime和sleepInterval的写法
	err := v.base.WithContext(dbCtx).Order("id").Offset(offset).First(&src).Error
	return src, err
}

// incrFromBase 增量校验，用于源表到目标表的校验
func (v *Validator[T]) incrFromBase(ctx context.Context, offset int) (T, error) {
	dbCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	var src T
	err := v.base.WithContext(dbCtx).Where("utime > ?", v.utime).
		Order("utime").Offset(offset).First(&src).Error
	return src, err
}

// validateBaseToTarget 源表到目标表的校验
// 源表有，目标表没有
func (v *Validator[T]) validateBaseToTarget(ctx context.Context) error {
	offset := 0
	for {
		// 校验要选择合适的时机，比如进来就看看负载是否高，高的话可以睡一会再进来看看
		src, err := v.formBase(ctx, offset)
		if err == context.DeadlineExceeded || err == context.Canceled {
			return nil
		}
		if err == gorm.ErrRecordNotFound {
			// 增量校验要考虑一直运行的
			if v.sleepInterval <= 0 {
				return nil
			}
			time.Sleep(v.sleepInterval)
			continue
			// 这个就是没有数据了
		}
		if err != nil {
			// 查询出错
			v.l.Error("base -> target 查询base失败", logger.Error(err))
			// 在这里offset是要+1，因为不能因为一条错误数据中断整个校验
			// 或者可以用更优雅的方式，同一个offset错3次再不管他，也就是引入了重试机制，但是实在没有必要
			// 增量写法
			offset++
			continue
		}
		// 这边就是正常情况
		var dst T
		err = v.target.WithContext(ctx).Where("id = ?", src.ID()).First(&dst).Error
		switch err {
		case gorm.ErrRecordNotFound:
			// target没有该数据
			// 丢一条休息到kafka上
			v.notify(src.ID(), events.InconsistentEventTypeTargetMissing)
		case nil:
			// target有该数据，没出错
			equal := src.CompareTo(dst)
			if !equal {
				// 要丢一条消息到kafka
				v.notify(src.ID(), events.InconsistentEventTypeNeq)
			}
		default:
			// 不知道什么错误
			// 记录日志，继续往下
			v.l.Error("base -> target 查询target失败",
				logger.Int64("id", src.ID()),
				logger.Error(err))
		}
		offset++
	}
}

// validateBaseToTargetV1 批量写法
func (v *Validator[T]) validateBaseToTargetV1(ctx context.Context) error {
	offset := 0
	const limit = 100
	for {
		var srcs []T
		dbCtx, cancel := context.WithTimeout(ctx, time.Second)
		err := v.base.WithContext(dbCtx).Where("utime > ?", v.utime).
			Order("utime").Offset(offset).Limit(limit).Find(&srcs).Error
		cancel()
		switch err {
		// 在Find中其实不会有这种错误
		//case gorm.ErrRecordNotFound:
		case context.Canceled, context.DeadlineExceeded:
			return err
		case nil:
			if len(srcs) == 0 {
				// 结束，没有数据
				return nil
			}
			err = v.dstDiffV1(srcs)
			if err != nil {
				return err
			}
		default:
			v.l.Error("base -> target 查询base失败", logger.Error(err))
		}
		if len(srcs) < limit {
			// 没有数据
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
	if err != nil {
		return err
	}
	dstMap := v.toMap(dsts)
	for _, src := range srcs {
		dst, ok := dstMap[src.ID()]
		if !ok {
			v.notify(src.ID(), events.InconsistentEventTypeTargetMissing)
			continue
		}
		if !src.CompareTo(dst) {
			v.notify(src.ID(), events.InconsistentEventTypeNeq)
		}
	}
	return nil
}

func (v *Validator[T]) toMap(dsts []T) map[int64]T {
	res := make(map[int64]T, len(dsts))
	for _, val := range dsts {
		res[val.ID()] = val
	}
	return res
}

// validateTargetToBase 源表没有，目标表有
// 这种情况很少见，唯一有的情况就是源表的数据同步到目标表后，在这期间，源表的某个数据被硬删除了，导致目标表有这个数据，源表没有
func (v *Validator[T]) validateTargetToBase(ctx context.Context) error {
	offset := 0
	for {
		var ts []T
		// 这里因为我们主要是比对源表如果有硬删除的，就把目标表的也删除掉，所以只用查id
		// 然后去源表里看这个id还在不在，不在就删除掉目标表的这个数据
		// 这里增量和全量用一个
		// 这里的检验和同步是base把数据删了，但是target的utime是不变的，所以没用utime
		err := v.target.WithContext(ctx).Select("id").
			Order("id").Offset(offset).Limit(v.batchSize).Find(&ts).Error
		if err == context.DeadlineExceeded || err == context.Canceled {
			return nil
		}
		if err == gorm.ErrRecordNotFound || len(ts) == 0 {
			// 增量校验要考虑一直运行的
			if v.sleepInterval <= 0 {
				return nil
			}
			time.Sleep(v.sleepInterval)
			continue
		}
		if err != nil {
			v.l.Error("target => base 查询target失败", logger.Error(err))
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
			// 都代表，base里面一条对应的数据都没有
			v.notifyBaseMissing(ts)
			offset += len(ts)
			continue
		}
		if err != nil {
			v.l.Error("target => base 查询 base 失败", logger.Error(err))
			// 保守起见，我都认为base里面没有数据，但是没多大必要
			// v.notifyBaseMissing(ts) 再调一次这个
			offset += len(ts)
			continue
		}
		// 找差集，diff里面的，就是target有，但是base没有的
		diff := slice.DiffSetFunc(ts, srcTs, func(src, dst T) bool {
			return src.ID() == dst.ID()
		})
		v.notifyBaseMissing(diff)
		// 没有找够一批，说明也没数据了
		if len(ts) < v.batchSize {
			if v.sleepInterval <= 0 {
				return nil
			}
			time.Sleep(v.sleepInterval)
		}
		offset += len(ts)
	}
}

// notifyBaseMissing 目标表有数据，源表没有数据的通知
func (v *Validator[T]) notifyBaseMissing(ts []T) {
	for _, val := range ts {
		v.notify(val.ID(), events.InconsistentEventTypeBaseMissing)
	}
}

// notify 通知一条消息
func (v *Validator[T]) notify(id int64, typ string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err := v.producer.ProduceInconsistentEvent(ctx, events.InconsistentEvent{
		ID:        id,
		Direction: v.direction,
		Type:      typ,
	})
	if err != nil {
		v.l.Error("发送不一致消息失败",
			logger.Int64("id", id),
			logger.String("type", typ),
			logger.Error(err))
	}
}
