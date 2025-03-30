package dao

import (
	"context"
	"gorm.io/gorm"
)

type FeedPushEventDAO interface {
	CreatePushEvents(ctx context.Context, events []FeedPushEvent) error
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

func (f *feedPushEventDAO) CreatePushEvents(ctx context.Context, events []FeedPushEvent) error {
	return f.db.WithContext(ctx).Create(&events).Error
}

func NewFeedPushEventDAO(db *gorm.DB) FeedPushEventDAO {
	return &feedPushEventDAO{db: db}
}
