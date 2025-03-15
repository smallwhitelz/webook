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
}
