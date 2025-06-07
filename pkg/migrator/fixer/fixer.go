package fixer

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"webook/pkg/migrator"
	"webook/pkg/migrator/events"
)

type OverrideFixer[T migrator.Entity] struct {
	base   *gorm.DB
	target *gorm.DB

	columns []string
}

func NewOverrideFixer[T migrator.Entity](base *gorm.DB, target *gorm.DB) (*OverrideFixer[T], error) {
	row, err := base.Model(new(T)).Order("id").Rows()
	if err != nil {
		return nil, err
	}
	columns, err := row.Columns()
	// columns 也可以让外界去传
	return &OverrideFixer[T]{base: base, target: target, columns: columns}, err
}

// Fix 最最粗暴的，直接覆盖的写法
// 实际上，在修复数据的时候，根本不需要去管校验出来的不一致是啥
// 直接覆盖式写法
func (f *OverrideFixer[T]) Fix(ctx context.Context, id int64) error {
	var t T
	err := f.base.WithContext(ctx).Where("id = ?", id).First(&t).Error
	switch err {
	case gorm.ErrRecordNotFound:
		// base 没有数据，target则执行delete
		return f.target.WithContext(ctx).Model(&t).Delete("id = ?", id).Error
	case nil:
		// 找到了，
		// 这里因为可能有并发问题，所以用Upsert语义，
		return f.target.WithContext(ctx).Clauses(clause.OnConflict{
			DoUpdates: clause.AssignmentColumns(f.columns),
		}).Create(&t).Error
	default:
		return err
	}
}

// FixV1 多余写法
func (f *OverrideFixer[T]) FixV1(evt events.InconsistentEvent) error {
	switch evt.Type {
	case events.InconsistentEventTypeNeq, events.InconsistentEventTypeTargetMissing:
		var t T
		err := f.base.Where("id = ?", evt.ID).First(&t).Error
		switch err {
		case gorm.ErrRecordNotFound:
			// base 没有数据，target则执行delete
			return f.target.Model(&t).Delete("id = ?", evt.ID).Error
		case nil:
			// 找到了，
			// 这里因为可能有并发问题，所以用Upsert语义，
			return f.target.Clauses(clause.OnConflict{
				DoUpdates: clause.AssignmentColumns(f.columns),
			}).Create(&t).Error
		default:
			return err
		}
	case events.InconsistentEventTypeBaseMissing:
		return f.target.Model(new(T)).Delete("id = ?", evt.ID).Error
	}
	return nil
}
