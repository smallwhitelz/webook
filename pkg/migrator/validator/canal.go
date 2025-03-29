package validator

import (
	"context"
	"gorm.io/gorm"
	"webook/pkg/logger"
	"webook/pkg/migrator"
	"webook/pkg/migrator/events"
)

type CanalIncrValidator[T migrator.Entity] struct {
	baseValidator
}

func NewCanalIncrValidator[T migrator.Entity](
	base *gorm.DB,
	target *gorm.DB,
	direction string,
	l logger.V1,
	producer events.Producer,
) *CanalIncrValidator[T] {
	return &CanalIncrValidator[T]{
		baseValidator: baseValidator{
			base:      base,
			target:    target,
			direction: direction,
			l:         l,
			producer:  producer,
		},
	}
}

// Validate 一次校验一条
// id是被修改的数据的主键
func (v *CanalIncrValidator[T]) Validate(ctx context.Context, id int64) error {
	var base T
	err := v.base.WithContext(ctx).Where("id = ?", id).First(&base).Error
	switch err {
	case nil:
		// 找到了
		var target T
		err = v.target.WithContext(ctx).Where("id = ?", id).First(&target).Error
		switch err {
		case nil:
			// 两边都找到了
			if !base.CompareTo(target) {
				v.notify(id, events.InconsistentEventTypeNEQ)
				return nil
			}
		case gorm.ErrRecordNotFound:
			// base有 target没有
			v.notify(id, events.InconsistentEventTypeTargetMissing)
			return nil
		default:
			return err
		}
	case gorm.ErrRecordNotFound:
		// base 没找到
		var target T
		err = v.target.WithContext(ctx).Where("id = ?", id).First(&target).Error
		switch err {
		case nil:
			// target找到了
			v.notify(id, events.InconsistentEventTypeBaseMissing)
		case gorm.ErrRecordNotFound:
			return nil
		default:
			return err
		}
	default:
		// 不知道啥错误
		return err
	}
	return nil
}
