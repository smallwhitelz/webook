package repository

import (
	"context"
	"encoding/json"
	"time"
	"webook/feed/domain"
	"webook/feed/repository/cache"
	"webook/feed/repository/dao"
)

type FeedEventRepo interface {
	// CreatePushEvents 批量推事件
	CreatePushEvents(ctx context.Context, evts []domain.FeedEvent) error
	// CreatePullEvents 创建拉事件
	CreatePullEvents(ctx context.Context, evt domain.FeedEvent) error
	// FindPullEvents 获取拉事件，也就是关注的人发件箱里面的事件
	FindPullEvents(ctx context.Context, uids []int64, timestamp, limit int64) ([]domain.FeedEvent, error)
	// FindPushEvents 获取推事件，也就是自己收件箱里面的事件
	FindPushEvents(ctx context.Context, uid, timestamp, limit int64) ([]domain.FeedEvent, error)

	// FindPushEventsWithTyp 获取某个类型的推事件，也就
	FindPushEventsWithTyp(ctx context.Context, typ string, uid int64,
		timestamp int64, limit int64) ([]domain.FeedEvent, error)
	// FindPullEventsWithTyp 获取某个类型的拉事件，
	FindPullEventsWithTyp(ctx context.Context, typ string, uids []int64,
		timestamp int64, limit int64) ([]domain.FeedEvent, error)
}

type feedEventRepo struct {
	pullDao   dao.FeedPullEventDAO
	pushDao   dao.FeedPushEventDAO
	feedCache cache.FeedEventCache
}

func (f *feedEventRepo) FindPullEventsWithTyp(ctx context.Context, typ string, uids []int64, timestamp int64, limit int64) ([]domain.FeedEvent, error) {
	events, err := f.pullDao.FindPullEventsListWithTyp(ctx, typ, uids, timestamp, limit)
	if err != nil {
		return nil, err
	}
	res := make([]domain.FeedEvent, 0, len(events))
	for _, evt := range events {
		res = append(res, f.convertToPullEventDomain(evt))
	}
	return res, nil
}

func (f *feedEventRepo) FindPullEvents(ctx context.Context, uids []int64, timestamp, limit int64) ([]domain.FeedEvent, error) {
	events, err := f.pullDao.FindPullEventList(ctx, uids, timestamp, limit)
	if err != nil {
		return nil, err
	}
	ans := make([]domain.FeedEvent, 0, len(events))
	for _, e := range events {
		ans = append(ans, f.convertToPullEventDomain(e))
	}
	return ans, nil
}

func (f *feedEventRepo) FindPushEvents(ctx context.Context, uid, timestamp, limit int64) ([]domain.FeedEvent, error) {
	events, err := f.pushDao.GetPushEvents(ctx, uid, timestamp, limit)
	if err != nil {
		return nil, err
	}
	ans := make([]domain.FeedEvent, 0, len(events))
	for _, e := range events {
		ans = append(ans, f.convertToPushEventDomain(e))
	}
	return ans, nil
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

func (f *feedEventRepo) FindPushEventsWithTyp(ctx context.Context, typ string, uid int64, timestamp int64, limit int64) ([]domain.FeedEvent, error) {
	events, err := f.pushDao.GetPushEventsWithTyp(ctx, typ, uid, timestamp, limit)
	if err != nil {
		return nil, err
	}
	ans := make([]domain.FeedEvent, 0, len(events))
	for _, e := range events {
		ans = append(ans, f.convertToPushEventDomain(e))
	}
	return ans, nil
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

func (f *feedEventRepo) convertToPullEventDomain(event dao.FeedPullEvent) domain.FeedEvent {
	var ext map[string]string
	_ = json.Unmarshal([]byte(event.Content), &ext)
	return domain.FeedEvent{
		ID:    event.Id,
		Uid:   event.UID,
		Type:  event.Type,
		Ctime: time.Unix(event.Ctime, 0),
		Ext:   ext,
	}
}

func (f *feedEventRepo) convertToPushEventDomain(event dao.FeedPushEvent) domain.FeedEvent {
	var ext map[string]string
	_ = json.Unmarshal([]byte(event.Content), &ext)
	return domain.FeedEvent{
		ID:    event.Id,
		Uid:   event.UID,
		Type:  event.Type,
		Ctime: time.Unix(event.Ctime, 0),
		Ext:   ext,
	}
}

func NewFeedEventRepo(pullDao dao.FeedPullEventDAO, pushDao dao.FeedPushEventDAO, feedCache cache.FeedEventCache) FeedEventRepo {
	return &feedEventRepo{
		pullDao:   pullDao,
		pushDao:   pushDao,
		feedCache: feedCache,
	}
}
