package repository

import (
	"context"
	"github.com/ecodeclub/ekit/slice"
	"time"
	"webook/pkg/logger"
	"webook/tag/domain"
	"webook/tag/repository/cache"
	"webook/tag/repository/dao"
)

type TagRepository interface {
	CreateTag(ctx context.Context, tag domain.Tag) (int64, error)
	GetTags(ctx context.Context, uid int64) ([]domain.Tag, error)
	BindTagToBiz(ctx context.Context, uid int64, biz string, bizId int64, tags []int64) error
	GetBizTags(ctx context.Context, uid int64, biz string, bizId int64) ([]domain.Tag, error)
	GetTagsById(ctx context.Context, tagIds []int64) ([]domain.Tag, error)
}

type CachedTagRepository struct {
	dao   dao.TagDAO
	cache cache.TagCache
	l     logger.V1
}

func (c *CachedTagRepository) GetTagsById(ctx context.Context, tagIds []int64) ([]domain.Tag, error) {
	tags, err := c.dao.GetTagsById(ctx, tagIds)
	if err != nil {
		return nil, err
	}
	return slice.Map(tags, func(idx int, src dao.Tag) domain.Tag {
		return c.toDomain(src)
	}), nil
}

func (c *CachedTagRepository) GetBizTags(ctx context.Context, uid int64, biz string, bizId int64) ([]domain.Tag, error) {
	tags, err := c.dao.GetTagsByBiz(ctx, uid, biz, bizId)
	if err != nil {
		return nil, err
	}
	return slice.Map(tags, func(idx int, src dao.Tag) domain.Tag {
		return c.toDomain(src)
	}), nil
}

func (c *CachedTagRepository) BindTagToBiz(ctx context.Context, uid int64, biz string, bizId int64, tags []int64) error {
	return c.dao.CreateTagBiz(ctx, slice.Map(tags, func(idx int, src int64) dao.TagBiz {
		return dao.TagBiz{
			BizId: bizId,
			Biz:   biz,
			Uid:   uid,
			Tid:   src,
		}
	}))
}

// PreloadUserTags 在toB的场景下，可以提前缓存预加载
func (c *CachedTagRepository) PreloadUserTags(ctx context.Context) error {
	// 我们要存的是 uid => 我的所有标签
	// 这边分批次预加载
	// 数据取出来，调用Append
	offset := 0
	batch := 100
	for {
		dbCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		tags, err := c.dao.GetTags(dbCtx, offset, batch)
		cancel()
		if err != nil {
			return err
		}
		for _, tag := range tags {
			rctx, cancel := context.WithTimeout(context.Background(), time.Second)
			err := c.cache.Append(rctx, tag.Uid, c.toDomain(tag))
			cancel()
			if err != nil {
				continue
			}
		}
		if len(tags) < batch {
			return nil
		}
		offset = offset + batch
	}
}

func (c *CachedTagRepository) GetTags(ctx context.Context, uid int64) ([]domain.Tag, error) {
	res, err := c.cache.GetTags(ctx, uid)
	if err == nil {
		return res, nil
	}
	tags, err := c.dao.GetTagsByUid(ctx, uid)
	if err != nil {
		return nil, err
	}
	res = slice.Map(tags, func(idx int, src dao.Tag) domain.Tag {
		return c.toDomain(src)
	})
	err = c.cache.Append(ctx, uid, res...)
	if err != nil {
		// 记录日志

	}
	return res, nil
}

func (c *CachedTagRepository) CreateTag(ctx context.Context, tag domain.Tag) (int64, error) {
	id, err := c.dao.CreateTag(ctx, c.toEntity(tag))
	if err != nil {
		return 0, err
	}
	err = c.cache.Append(ctx, tag.Uid, tag)
	if err != nil {
		// 记录日志
		c.l.Error("新建标签更新缓存失败", logger.Error(err))
	}
	return id, nil
}

func (c *CachedTagRepository) toEntity(tag domain.Tag) dao.Tag {
	return dao.Tag{
		Name: tag.Name,
		Uid:  tag.Uid,
	}
}

func (c *CachedTagRepository) toDomain(tag dao.Tag) domain.Tag {
	return domain.Tag{
		Id:   tag.Id,
		Name: tag.Name,
		Uid:  tag.Uid,
	}
}

func NewCachedTagRepository(dao dao.TagDAO, cache cache.TagCache, l logger.V1) TagRepository {
	return &CachedTagRepository{dao: dao, cache: cache, l: l}
}
