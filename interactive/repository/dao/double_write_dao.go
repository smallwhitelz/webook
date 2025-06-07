package dao

import (
	"context"
	"errors"
	"github.com/ecodeclub/ekit/syncx/atomicx"
	"gorm.io/gorm"
	"webook/pkg/logger"
)

var errUnknownPattern = errors.New("未知的双写模式")

// DoubleWriteDAO src为准的时候，src成功即成功，dst失败记录日志即可，以dst为准的时候同理
// 缺陷：1. 要修改的代码很多，容易遗漏 2.每一个DAO都要改过去，过于繁杂
type DoubleWriteDAO struct {
	src InteractiveDAO
	dst InteractiveDAO
	// 根据模式分发，有读有写用原子操作
	pattern *atomicx.Value[string]
	l       logger.LoggerV1
}

func NewDoubleWriteDAO(src *gorm.DB, dst *gorm.DB, l logger.LoggerV1) *DoubleWriteDAO {
	return &DoubleWriteDAO{
		src:     NewGORMInteractiveDAO(src),
		dst:     NewGORMInteractiveDAO(dst),
		l:       l,
		pattern: atomicx.NewValueOf(PatternSrcOnly),
	}
}

// UpdatePattern 在运行的时候切换pattern
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
			// 要不要return？
			// 正常来说，我们认为双写阶段，src成功了就算业务上成功了
			d.l.Error("双写写入 dst 失败",
				logger.String("biz", biz),
				logger.Int64("bizId", bizId),
				logger.Error(err))
		}
		return nil
	case PatternDstFirst:
		err := d.dst.IncrReadCnt(ctx, biz, bizId)
		if err == nil {
			err1 := d.src.IncrReadCnt(ctx, biz, bizId)
			if err1 != nil {
				d.l.Error("双写写入 src 失败",
					logger.String("biz", biz),
					logger.Int64("bizId", bizId),
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
	case PatternSrcOnly, PatternSrcFirst:
		return d.src.Get(ctx, biz, bizId)
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
	// PatternSrcOnly 第一阶段 读写源表
	PatternSrcOnly = "src_only"
	// PatternSrcFirst 第二阶段 读写源表，写目标表
	PatternSrcFirst = "src_first"
	// PatternDstFirst 第三阶段 读写目标表，写源表
	PatternDstFirst = "dst_first"
	// PatternDstOnly 第四阶段 读写目标表
	PatternDstOnly = "dst_only"
)
