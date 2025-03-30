package repository

import (
	"context"
	"encoding/json"
	"webook/feed/domain"
	"webook/feed/repository/cache"
	"webook/feed/repository/dao"
)

type FeedEventRepo interface {
	// CreatePushEvents 批量推事件
	CreatePushEvents(ctx context.Context, evts []domain.FeedEvent) error
	// CreatePullEvents 创建拉事件
	CreatePullEvents(ctx context.Context, evt domain.FeedEvent) error

	FindPushEventsWithTyp(ctx context.Context, name string, uid int64,
		timestamp int64, limit int64) ([]domain.FeedEvent, error)
}

type feedEventRepo struct {
	pullDao   dao.FeedPullEventDAO
	pushDao   dao.FeedPushEventDAO
	feedCache cache.FeedEventCache
}

func (f *feedEventRepo) CreatePushEvents(ctx context.Context, evts []domain.FeedEvent) error {
	pushEvents := make([]dao.FeedPushEvent, 0, len(evts))
	for _, evt := range evts {
		pushEvents = append(pushEvents, f.convertToPushEventDao(evt))
	}
	return f.pushDao.CreatePushEvents(ctx, pushEvents)
}

func (f *feedEventRepo) CreatePullEvents(ctx context.Context, evt domain.FeedEvent) error {
	return f.pullDao.CreatePullEvents(ctx, f.convertToPullEventDao(evt))
}

func (f *feedEventRepo) FindPushEventsWithTyp(ctx context.Context, name string, uid int64, timestamp int64, limit int64) ([]domain.FeedEvent, error) {
	//TODO implement me
	panic("implement me")
}

func (f *feedEventRepo) convertToPullEventDao(evt domain.FeedEvent) dao.FeedPullEvent {
	val, _ := json.Marshal(evt.Ext)
	return dao.FeedPullEvent{
		Id:      evt.ID,
		UID:     evt.Uid,
		Type:    evt.Type,
		Content: string(val),
		Ctime:   evt.Ctime.Unix(),
	}
}

func (f *feedEventRepo) convertToPushEventDao(evt domain.FeedEvent) dao.FeedPushEvent {
	val, _ := json.Marshal(evt.Ext)
	return dao.FeedPushEvent{
		Id:      evt.ID,
		UID:     evt.Uid,
		Type:    evt.Type,
		Content: string(val),
		Ctime:   evt.Ctime.Unix(),
	}
}

func NewFeedEventRepo(pullDao dao.FeedPullEventDAO, pushDao dao.FeedPushEventDAO, feedCache cache.FeedEventCache) FeedEventRepo {
	return &feedEventRepo{
		pullDao:   pullDao,
		pushDao:   pushDao,
		feedCache: feedCache,
	}
}
