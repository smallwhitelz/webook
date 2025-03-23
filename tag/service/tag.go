package service

import (
	"context"
	"github.com/ecodeclub/ekit/slice"
	"time"
	"webook/pkg/logger"
	"webook/tag/domain"
	"webook/tag/events"
	"webook/tag/repository"
)

type TagService interface {
	// CreateTag 用户创建标签
	CreateTag(ctx context.Context, uid int64, name string) (int64, error)
	// GetTags 查找用户自己创建的标签
	GetTags(ctx context.Context, uid int64) ([]domain.Tag, error)
	// AttachTags 用户重新打标签，采用覆盖式
	AttachTags(ctx context.Context, uid int64, biz string, bizId int64, tags []int64) error
	// GetBizTags 获取资源上的标签
	GetBizTags(ctx context.Context, uid int64, biz string, bizId int64) ([]domain.Tag, error)
}

type tagService struct {
	repo     repository.TagRepository
	l        logger.V1
	producer events.Producer
}

func (t *tagService) GetBizTags(ctx context.Context, uid int64, biz string, bizId int64) ([]domain.Tag, error) {
	return t.repo.GetBizTags(ctx, uid, biz, bizId)
}

func (t *tagService) AttachTags(ctx context.Context, uid int64, biz string, bizId int64, tagIds []int64) error {
	err := t.repo.BindTagToBiz(ctx, uid, biz, bizId, tagIds)
	if err != nil {
		return err
	}
	// 异步发送
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		tags, err := t.repo.GetTagsById(ctx, tagIds)
		cancel()
		if err != nil {
			return
		}
		ctx, cancel = context.WithTimeout(context.Background(), time.Second)
		err = t.producer.ProduceSyncEvent(ctx, events.BizTags{
			Biz:   biz,
			BizId: bizId,
			Uid:   uid,
			Tags: slice.Map(tags, func(idx int, src domain.Tag) string {
				return src.Name
			}),
		})
		cancel()
		if err != nil {
			// 记录日志
			t.l.Error("发送消息失败", logger.Error(err))
		}
	}()
	return err
}

func (t *tagService) GetTags(ctx context.Context, uid int64) ([]domain.Tag, error) {
	return t.repo.GetTags(ctx, uid)
}

func (t *tagService) CreateTag(ctx context.Context, uid int64, name string) (int64, error) {
	return t.repo.CreateTag(ctx, domain.Tag{
		Name: name,
		Uid:  uid,
	})
}

func NewTagService(repo repository.TagRepository, l logger.V1, producer events.Producer) TagService {
	return &tagService{repo: repo, l: l, producer: producer}
}
