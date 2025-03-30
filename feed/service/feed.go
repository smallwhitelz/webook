package service

import (
	"context"
	"fmt"
	followv1 "webook/api/proto/gen/follow/v1"
	"webook/feed/domain"
	"webook/feed/repository"
)

type feedService struct {
	repo repository.FeedEventRepo
	// 对应的 string 就是 type
	handlerMap   map[string]Handler
	followClient followv1.FollowServiceClient
}

func (f *feedService) GetFeedEventList(ctx context.Context, uid, timestamp, limit int64) ([]domain.FeedEvent, error) {
	//TODO implement me
	panic("implement me")
}

func NewFeedService(repo repository.FeedEventRepo, handlerMap map[string]Handler) FeedService {
	return &feedService{
		repo:       repo,
		handlerMap: handlerMap,
	}
}

func (f *feedService) RegisterService(typ string, handler Handler) {
	f.handlerMap[typ] = handler
}

func (f *feedService) CreateFeedEvent(ctx context.Context, feed domain.FeedEvent) error {
	handler, ok := f.handlerMap[feed.Type]
	if !ok {
		// 说明type不对
		// 你还可以考虑兜底机制
		// 有一个 defaultHandler，然后调用 defaultHandler
		return fmt.Errorf("未能找到对应的 Handler %s", feed.Type)
	}
	return handler.CreateFeedEvent(ctx, feed.Ext)
}
