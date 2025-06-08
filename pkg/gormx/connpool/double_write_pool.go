package connpool

import (
	"context"
	"database/sql"
	"errors"
	"github.com/ecodeclub/ekit/syncx/atomicx"
	"gorm.io/gorm"
	"webook/pkg/logger"
)

var errUnknownPattern = errors.New("未知的双写模式")

// DoubleWritePool 借助gorm底层接口ConnPool实现双写
type DoubleWritePool struct {
	src     gorm.ConnPool
	dst     gorm.ConnPool
	pattern *atomicx.Value[string]
	l       logger.LoggerV1
}

func NewDoubleWritePool(src *gorm.DB, dst *gorm.DB, l logger.LoggerV1) *DoubleWritePool {
	return &DoubleWritePool{
		src:     src.ConnPool,
		dst:     dst.ConnPool,
		l:       l,
		pattern: atomicx.NewValueOf(PatternSrcOnly)}
}

// UpdatePattern 动态改动pattern
func (d *DoubleWritePool) UpdatePattern(pattern string) error {
	// 不是合法的pattern
	switch pattern {
	case PatternSrcOnly, PatternSrcFirst, PatternDstOnly, PatternDstFirst:
		d.pattern.Store(pattern)
		return nil
	default:
		return errUnknownPattern
	}
}

// BeginTx 实现ConnPoolBeginner接口
// 用来支持双写中，如果存在事务的操作
func (d *DoubleWritePool) BeginTx(ctx context.Context, opts *sql.TxOptions) (gorm.ConnPool, error) {
	pattern := d.pattern.Load()
	switch pattern {
	case PatternSrcOnly:
		src, err := d.src.(gorm.TxBeginner).BeginTx(ctx, opts)
		return &DoubleWriteTx{src: src, l: d.l, pattern: pattern}, err
	case PatternSrcFirst:
		src, err := d.src.(gorm.TxBeginner).BeginTx(ctx, opts)
		if err != nil {
			return nil, err
		}
		dst, err := d.dst.(gorm.TxBeginner).BeginTx(ctx, opts)
		if err != nil {
			d.l.Error("双写目标表开启事务失败", logger.Error(err))
		}
		return &DoubleWriteTx{src: src, dst: dst, l: d.l, pattern: pattern}, nil
	case PatternDstFirst:
		dst, err := d.dst.(gorm.TxBeginner).BeginTx(ctx, opts)
		if err != nil {
			return nil, err
		}
		src, err := d.src.(gorm.TxBeginner).BeginTx(ctx, opts)
		if err != nil {
			d.l.Error("双写源表开启事务失败", logger.Error(err))
		}
		return &DoubleWriteTx{src: src, dst: dst, l: d.l, pattern: pattern}, nil
	case PatternDstOnly:
		dst, err := d.dst.(gorm.TxBeginner).BeginTx(ctx, opts)
		return &DoubleWriteTx{dst: dst, l: d.l, pattern: pattern}, err
	default:
		return nil, errUnknownPattern
	}
}

func (d *DoubleWritePool) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	// 这个没办法改写
	// 我没办法返回一个双写的 sql.Stmt
	panic("双写模式下不支持")
}

func (d *DoubleWritePool) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	switch d.pattern.Load() {
	case PatternSrcOnly:
		return d.src.ExecContext(ctx, query, args...)
	case PatternSrcFirst:
		res, err := d.src.ExecContext(ctx, query, args...)
		if err == nil {
			_, err1 := d.dst.ExecContext(ctx, query, args...)
			if err1 != nil {
				d.l.Error("双写写入 dst 失败",
					logger.String("sql", query),
					logger.Error(err1))
			}
		}
		return res, err
	case PatternDstFirst:
		res, err := d.dst.ExecContext(ctx, query, args...)
		if err == nil {
			_, err1 := d.src.ExecContext(ctx, query, args...)
			if err1 != nil {
				d.l.Error("双写写入 src 失败",
					logger.String("sql", query),
					logger.Error(err1))
			}
		}
		return res, err
	case PatternDstOnly:
		return d.dst.ExecContext(ctx, query, args...)
	default:
		return nil, errUnknownPattern
	}
}

// QueryContext 查多行
func (d *DoubleWritePool) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	switch d.pattern.Load() {
	case PatternSrcOnly, PatternSrcFirst:
		return d.src.QueryContext(ctx, query, args...)
	case PatternDstOnly, PatternDstFirst:
		return d.dst.QueryContext(ctx, query, args...)
	default:
		return nil, errUnknownPattern
	}
}

// QueryRowContext 查单行
func (d *DoubleWritePool) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	switch d.pattern.Load() {
	case PatternSrcOnly, PatternSrcFirst:
		return d.src.QueryRowContext(ctx, query, args...)
	case PatternDstOnly, PatternDstFirst:
		return d.dst.QueryRowContext(ctx, query, args...)
	default:
		// 这样返回没有带上错误信息
		//return &sql.Row{}
		panic(errUnknownPattern)
	}
}

type DoubleWriteTx struct {
	src *sql.Tx
	dst *sql.Tx
	// 事务本身可以保证pattern从开始到结束只有一个
	pattern string
	l       logger.LoggerV1
}

// Commit 提交事务 实现TxCommitter接口
func (d *DoubleWriteTx) Commit() error {
	switch d.pattern {
	case PatternSrcOnly:
		return d.src.Commit()
	case PatternSrcFirst:
		err := d.src.Commit()
		// 如果src提交失败了
		if err != nil {
			return err
		}
		// 这里是确定dst是开了事务的
		if d.dst != nil {
			err1 := d.dst.Commit()
			if err1 != nil {
				// src提交成功，dst失败，只能记录日志
				d.l.Error("目标表提交事务失败", logger.Error(err1))
			}
		}
		return nil
	case PatternDstFirst:
		err := d.dst.Commit()
		// 如果dst提交失败了
		if err != nil {
			return err
		}
		// 这里是确定src是开了事务的
		if d.src != nil {
			err1 := d.src.Commit()
			if err1 != nil {
				// dst提交成功，src失败，只能记录日志
				d.l.Error("源表提交事务失败", logger.Error(err1))
			}
		}
		return nil
	case PatternDstOnly:
		return d.dst.Commit()
	default:
		return errUnknownPattern
	}
}

// Rollback 回滚事务 实现TxCommitter接口
func (d *DoubleWriteTx) Rollback() error {
	switch d.pattern {
	case PatternSrcOnly:
		return d.src.Rollback()
	case PatternSrcFirst:
		err := d.src.Rollback()
		if err != nil {
			return err
		}
		if d.dst != nil {
			err1 := d.dst.Rollback()
			if err1 != nil {
				d.l.Error("目标表回滚事务失败", logger.Error(err1))
			}
		}
		return nil
	case PatternDstFirst:
		err := d.dst.Rollback()
		if err != nil {
			return err
		}
		if d.src != nil {
			err1 := d.src.Rollback()
			if err1 != nil {
				d.l.Error("源表回滚事务失败", logger.Error(err1))
			}
		}
		return nil
	case PatternDstOnly:
		return d.dst.Rollback()
	default:
		return errUnknownPattern
	}
}

func (d *DoubleWriteTx) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	// 这个没办法改写
	// 我没办法返回一个双写的 sql.Stmt
	panic("双写模式下不支持")
}

func (d *DoubleWriteTx) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	switch d.pattern {
	case PatternSrcOnly:
		return d.src.ExecContext(ctx, query, args...)
	case PatternSrcFirst:
		res, err := d.src.ExecContext(ctx, query, args...)
		if err == nil && d.dst != nil {
			_, err1 := d.dst.ExecContext(ctx, query, args...)
			if err1 != nil {
				d.l.Error("双写写入 dst 失败",
					logger.String("sql", query),
					logger.Error(err1))
			}
		}
		return res, err
	case PatternDstFirst:
		res, err := d.dst.ExecContext(ctx, query, args...)
		// d.src!=nil 开事务开成功了
		if err == nil && d.src != nil {
			_, err1 := d.src.ExecContext(ctx, query, args...)
			if err1 != nil {
				d.l.Error("双写写入 src 失败",
					logger.String("sql", query),
					logger.Error(err1))
			}
		}
		return res, err
	case PatternDstOnly:
		return d.dst.ExecContext(ctx, query, args...)
	default:
		return nil, errUnknownPattern
	}
}

// QueryContext 查多行
func (d *DoubleWriteTx) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	switch d.pattern {
	case PatternSrcOnly, PatternSrcFirst:
		return d.src.QueryContext(ctx, query, args...)
	case PatternDstOnly, PatternDstFirst:
		return d.dst.QueryContext(ctx, query, args...)
	default:
		return nil, errUnknownPattern
	}
}

// QueryRowContext 查单行
func (d *DoubleWriteTx) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	switch d.pattern {
	case PatternSrcOnly, PatternSrcFirst:
		return d.src.QueryRowContext(ctx, query, args...)
	case PatternDstOnly, PatternDstFirst:
		return d.dst.QueryRowContext(ctx, query, args...)
	default:
		// 这样返回没有带上错误信息
		//return &sql.Row{}
		panic(errUnknownPattern)
	}
}

const (
	// PatternSrcOnly 第一阶段 读写源表
	PatternSrcOnly = "src_only"
	// PatternSrcFirst 第二阶段 读写源表，写目标表
	PatternSrcFirst = "src_first"
	// PatternDstFirst 第三阶段 读写目标表，写源表
	PatternDstFirst = "dst_first"
	// PatternDstOnly 第四阶段 读写目标表
	PatternDstOnly = "dst_only"
)
