package dao

import (
	"context"
	"gorm.io/gorm"
)

type FeedPushEventDAO interface {
	CreatePushEvents(ctx context.Context, events []FeedPushEvent) error
	GetPushEventsWithTyp(ctx context.Context, typ string, uid int64, timestamp int64, limit int64) ([]FeedPushEvent, error)
	GetPushEvents(ctx context.Context, uid int64, timestamp int64, limit int64) ([]FeedPushEvent, error)
}

// FeedPushEvent 对应的是收件箱
type FeedPushEvent struct {
	Id int64 `gorm:"primaryKey,autoIncrement"`
	// 收件人
	UID  int64 `gorm:"index;column:uid"`
	Type string
	// 这边放的就是关键的扩展字段，不同的事件类型，有不用的解析方式
	Content string
	Ctime   int64
	// 正常来说，这个表里的数据不会被更新
	//Utime int64
}

type feedPushEventDAO struct {
	db *gorm.DB
}

func (f *feedPushEventDAO) GetPushEvents(ctx context.Context, uid int64, timestamp int64, limit int64) ([]FeedPushEvent, error) {
	var events []FeedPushEvent
	err := f.db.WithContext(ctx).
		Where("uid = ?", uid).
		Where("ctime < ?", timestamp).
		Order("ctime desc").
		Limit(int(limit)).
		Find(&events).Error
	return events, err
}

func (f *feedPushEventDAO) GetPushEventsWithTyp(ctx context.Context, typ string, uid int64, timestamp int64, limit int64) ([]FeedPushEvent, error) {
	var events []FeedPushEvent
	err := f.db.WithContext(ctx).
		Where("uid = ?", uid).
		Where("ctime < ?", timestamp).
		Where("type = ?", typ).
		Order("ctime desc").
		Limit(int(limit)).
		Find(&events).Error
	return events, err
}

func (f *feedPushEventDAO) CreatePushEvents(ctx context.Context, events []FeedPushEvent) error {
	return f.db.WithContext(ctx).Create(&events).Error
}

func NewFeedPushEventDAO(db *gorm.DB) FeedPushEventDAO {
	return &feedPushEventDAO{db: db}
}
