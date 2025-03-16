package dao

import "context"

// FollowRelation 存储用户的关注数据
type FollowRelation struct {
	ID int64 `gorm:"column:id;autoIncrement;primaryKey;"`

	// 要在这两个列上，创建一个联合唯一索引
	// 如果你认为查询一个人关注了多少人，是主要查询场景
	// <follower, followee>
	// 如果你认为查询一个人有哪些粉丝，是主要查询场景
	// <followee, follower>
	// 我查我关注了哪些人？ WHERE follower = 123(我的 uid)
	Follower int64 `gorm:"uniqueIndex:follower_followee"`
	Followee int64 `gorm:"uniqueIndex:follower_followee"`

	// 软删除策略
	Status uint8

	// 如果你的关注有类型的，有优先级，有一些备注数据的
	// Type string
	// Priority string
	// Gid 分组ID

	Ctime int64
	Utime int64
}

const (
	FollowRelationStatusUnknown uint8 = iota
	FollowRelationStatusActive
	FollowRelationStatusInactive
)

type FollowRelationDAO interface {
	// CreateFollowRelation 创建关注关系
	CreateFollowRelation(ctx context.Context, f FollowRelation) error
	// UpdateStatus 更新状态
	UpdateStatus(ctx context.Context, follower int64, followee int64, status uint8) error
	FollowRelationList(ctx context.Context, follower int64, offset int64, limit int64) ([]FollowRelation, error)
	FollowRelationDetail(ctx context.Context, follower int64, followee int64) (FollowRelation, error)
	CntFollow(ctx context.Context, uid int64) (int64, error)
	CntFollowee(ctx context.Context, uid int64) (int64, error)
}

// UserRelation 另外一种设计方案，但是不要这么做
type UserRelation struct {
	ID     int64 `gorm:"primaryKey,autoIncrement,column:id"`
	Uid1   int64 `gorm:"column:uid1;type:int(11);not null;uniqueIndex:user_contact_index"`
	Uid2   int64 `gorm:"column:uid2;type:int(11);not null;uniqueIndex:user_contact_index"`
	Block  bool  // 拉黑
	Mute   bool  // 屏蔽
	Follow bool  // 关注
}

type UserRelationV1 struct {
	ID   int64 `gorm:"primaryKey,autoIncrement,column:id"`
	Uid1 int64 `gorm:"column:uid1;type:int(11);not null;uniqueIndex:user_contact_index"`
	Uid2 int64 `gorm:"column:uid2;type:int(11);not null;uniqueIndex:user_contact_index"`
	Type string
}

type FollowStatics struct {
	ID  int64 `gorm:"primaryKey,autoIncrement,column:id"`
	Uid int64 `gorm:"unique"`
	// 有多少粉丝
	Followers int64
	// 关注了多少人
	Followees int64

	Utime int64
	Ctime int64
}
