package dao

import (
	"context"
	"gorm.io/gorm"
)

type FeedPullEventDAO interface {
	CreatePullEvents(ctx context.Context, evt FeedPullEvent) error
	FindPullEventsListWithTyp(ctx context.Context, typ string, uids []int64, timestamp int64, limit int64) ([]FeedPullEvent, error)
	FindPullEventList(ctx context.Context, uids []int64, timestamp int64, limit int64) ([]FeedPullEvent, error)
}

// FeedPullEvent 对应的是发件箱
type FeedPullEvent struct {
	Id int64 `gorm:"primaryKey,autoIncrement"`
	// 发件人
	UID  int64 `gorm:"index;column:uid"`
	Type string
	// 这边放的就是关键的扩展字段，不同的事件类型，有不用的解析方式
	Content string
	Ctime   int64
	// 正常来说，这个表里的数据不会被更新
	//Utime int64
}

type feedPullEventDAO struct {
	db *gorm.DB
}

func (f *feedPullEventDAO) FindPullEventList(ctx context.Context, uids []int64, timestamp int64, limit int64) ([]FeedPullEvent, error) {
	var events []FeedPullEvent
	err := f.db.WithContext(ctx).
		Where("uid in ?", uids).
		Where("ctime < ?", timestamp).
		Order("ctime desc").
		Limit(int(limit)).
		Find(&events).Error
	return events, err
}

func (f *feedPullEventDAO) FindPullEventsListWithTyp(ctx context.Context, typ string, uids []int64, timestamp int64, limit int64) ([]FeedPullEvent, error) {
	var events []FeedPullEvent
	err := f.db.WithContext(ctx).
		Where("uid IN ?", uids).
		Where("type = ?", typ).
		Where("ctime < ?", timestamp).
		Order("ctime desc").
		Limit(int(limit)).Find(&events).Error
	return events, err
}

func (f *feedPullEventDAO) CreatePullEvents(ctx context.Context, evt FeedPullEvent) error {
	return f.db.WithContext(ctx).Create(&evt).Error
}

func NewFeedPullEventDAO(db *gorm.DB) FeedPullEventDAO {
	return &feedPullEventDAO{db: db}
}
