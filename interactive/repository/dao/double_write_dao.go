package dao

import (
	"context"
	"errors"
	"github.com/ecodeclub/ekit/syncx/atomicx"
	"gorm.io/gorm"
	"webook/pkg/logger"
)

var errUnknownPattern = errors.New("未知的双写模式")

type DoubleWriteDAO struct {
	src     InteractiveDAO
	dst     InteractiveDAO
	pattern *atomicx.Value[string]
	l       logger.V1
}

func NewDoubleWriteDAO(src *gorm.DB, dst *gorm.DB, l logger.V1) *DoubleWriteDAO {
	return &DoubleWriteDAO{
		src:     NewGORMInteractiveDAO(src),
		dst:     NewGORMInteractiveDAO(dst),
		l:       l,
		pattern: atomicx.NewValueOf(PatternSrcOnly),
	}
}

func (d *DoubleWriteDAO) UpdatePattern(pattern string) {
	d.pattern.Store(pattern)
}

func (d *DoubleWriteDAO) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	pattern := d.pattern.Load()
	switch pattern {
	case PatternSrcOnly:
		return d.src.IncrReadCnt(ctx, biz, bizId)
	case PatternSrcFirst:
		err := d.src.IncrReadCnt(ctx, biz, bizId)
		if err != nil {
			return err
		}
		err = d.dst.IncrReadCnt(ctx, biz, bizId)
		if err != nil {
			// 要不要return
			// 正常来说，我们认为双写阶段，src 成功了就算业务上成功了
			d.l.Error("双写写入dst失败",
				logger.String("biz", biz),
				logger.Int64("biz_id", bizId),
				logger.Error(err))
		}
		return nil
	case PatternDstFirst:
		err := d.dst.IncrReadCnt(ctx, biz, bizId)
		if err == nil {
			err1 := d.src.IncrReadCnt(ctx, biz, bizId)
			if err1 != nil {
				d.l.Error("双写写入src失败",
					logger.String("biz", biz),
					logger.Int64("biz_id", bizId),
					logger.Error(err1))
			}
		}
		return err
	case PatternDstOnly:
		return d.dst.IncrReadCnt(ctx, biz, bizId)
	default:
		return errUnknownPattern
	}
}

func (d *DoubleWriteDAO) BatchIncrReadCnt(ctx context.Context, bizs []string, bizIds []int64) error {
	//TODO implement me
	panic("implement me")
}

func (d *DoubleWriteDAO) InsertLikeInfo(ctx context.Context, biz string, bizId int64, uid int64) error {
	//TODO implement me
	panic("implement me")
}

func (d *DoubleWriteDAO) DeleteLikeInfo(ctx context.Context, biz string, bizId int64, uid int64) error {
	//TODO implement me
	panic("implement me")
}

func (d *DoubleWriteDAO) InsertCollectionBiz(ctx context.Context, cb UserCollectionBiz) error {
	//TODO implement me
	panic("implement me")
}

func (d *DoubleWriteDAO) GetLikeInfo(ctx context.Context, biz string, bizId int64, uid int64) (UserLikeBiz, error) {
	//TODO implement me
	panic("implement me")
}

func (d *DoubleWriteDAO) GetCollectInfo(ctx context.Context, biz string, bizId int64, uid int64) (UserCollectionBiz, error) {
	//TODO implement me
	panic("implement me")
}

func (d *DoubleWriteDAO) Get(ctx context.Context, biz string, bizId int64) (Interactive, error) {
	pattern := d.pattern.Load()
	switch pattern {
	case PatternSrcFirst, PatternSrcOnly:
		return d.src.Get(ctx, biz, bizId)
	case PatternDstFirst, PatternDstOnly:
		return d.dst.Get(ctx, biz, bizId)
	default:
		return Interactive{}, errUnknownPattern
	}
}

// GetV1 在双写的时候进行校验并且修复数据，但是这样耦合性较高
func (d *DoubleWriteDAO) GetV1(ctx context.Context, biz string, bizId int64) (Interactive, error) {
	pattern := d.pattern.Load()
	switch pattern {
	case PatternSrcFirst, PatternSrcOnly:
		intr, err := d.src.Get(ctx, biz, bizId)
		if err == nil {
			intrDst, err1 := d.dst.Get(ctx, biz, bizId)
			if err1 != nil {
				if intr != intrDst {
					// 记录日志，并且通知修复程序
				}
			}
		}
		return intr, err
	case PatternDstFirst, PatternDstOnly:
		return d.dst.Get(ctx, biz, bizId)
	default:
		return Interactive{}, errUnknownPattern
	}
}

func (d *DoubleWriteDAO) GetByIds(ctx context.Context, biz string, ids []int64) ([]Interactive, error) {
	//TODO implement me
	panic("implement me")
}

const (
	PatternSrcOnly  = "src_only"
	PatternSrcFirst = "src_first"
	PatternDstFirst = "dst_first"
	PatternDstOnly  = "dst_only"
)
