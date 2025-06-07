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

// validateBaseToTarget 源表到目标表的校验
// 源表有，目标表没有
func (v *Validator[T]) validateBaseToTarget(ctx context.Context) error {
	// offset在最后肯定要++，在err!=nil也要++，所以干脆让他一开始就++
	// 从-1开始，进去就是第0个
	offset := -1
	for {
		// 校验要选择合适的时机，比如进来就看看负载是否高，高的话可以睡一会再进来看看
		offset++
		var src T
		err := v.base.WithContext(ctx).Order("id").Offset(offset).First(&src).Error
		if err == gorm.ErrRecordNotFound {
			// 这个就是没有数据了
			return nil
		}
		if err != nil {
			// 查询出错
			v.l.Error("base -> target 查询base失败", logger.Error(err))
			// 在这里offset是要+1，因为不能因为一条错误数据中断整个校验
			// 或者可以用更优雅的方式，同一个offset错3次再不管他，也就是引入了重试机制，但是实在没有必要
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
	}
}

// validateTargetToBase 源表没有，目标表有
// 这种情况很少见，唯一有的情况就是源表的数据同步到目标表后，在这期间，源表的某个数据被硬删除了，导致目标表有这个数据，源表没有
func (v *Validator[T]) validateTargetToBase(ctx context.Context) error {
	offset := -v.batchSize
	for {
		offset += v.batchSize
		var ts []T
		// 这里因为我们主要是比对源表如果有硬删除的，就把目标表的也删除掉，所以只用查id
		// 然后去源表里看这个id还在不在，不在就删除掉目标表的这个数据
		err := v.target.WithContext(ctx).Select("id").
			Order("id").Offset(offset).Limit(v.batchSize).Find(&ts).Error
		if err == gorm.ErrRecordNotFound || len(ts) == 0 {
			return nil
		}
		if err != nil {
			v.l.Error("target => base 查询target失败", logger.Error(err))
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
			continue
		}
		if err != nil {
			v.l.Error("target => base 查询 base 失败", logger.Error(err))
			// 保守起见，我都认为base里面没有数据，但是没多大必要
			// v.notifyBaseMissing(ts) 再调一次这个
			continue
		}
		// 找差集，diff里面的，就是target有，但是base没有的
		diff := slice.DiffSetFunc(ts, srcTs, func(src, dst T) bool {
			return src.ID() == dst.ID()
		})
		v.notifyBaseMissing(diff)
		// 没有找够一批，说明也没数据了
		if len(ts) < v.batchSize {
			return nil
		}
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
