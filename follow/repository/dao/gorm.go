package dao

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type GORMFollowRelationDAO struct {
	db *gorm.DB
}

func (g *GORMFollowRelationDAO) FollowRelationDetail(ctx context.Context, follower int64, followee int64) (FollowRelation, error) {
	var res FollowRelation
	err := g.db.WithContext(ctx).
		Where("follower = ? AND followee = ? AND status = ?", follower, followee, FollowRelationStatusActive).
		First(&res).Error
	return res, err
}

func (g *GORMFollowRelationDAO) FollowRelationList(ctx context.Context, follower int64, offset int64, limit int64) ([]FollowRelation, error) {
	var res []FollowRelation
	err := g.db.WithContext(ctx).
		Where("follower = ? AND status = ?", follower, FollowRelationStatusActive).
		Offset(int(offset)).Limit(int(limit)).Find(&res).Error
	return res, err
}

func (g *GORMFollowRelationDAO) UpdateStatus(ctx context.Context,
	follower int64, followee int64, status uint8) error {
	// 如果当前status就是inactive？
	// 没必要检测数据在不在，状态对不对，正常用户在没有关注的时候谁会去取消关注？黑客更不用管，并且这里是我自己传入的状态
	// 不管外界如何操作，我的状态是默认的改变方式
	return g.db.WithContext(ctx).Where("follower = ? AND followee = ?", follower, followee).
		Updates(map[string]any{
			"status": status,
			"utime":  time.Now().UnixMilli(),
		}).Error
}

func (g *GORMFollowRelationDAO) CreateFollowRelation(ctx context.Context, f FollowRelation) error {
	// 这里保持insert or update 语义
	now := time.Now().UnixMilli()
	f.Ctime = now
	f.Utime = now
	f.Status = FollowRelationStatusActive
	return g.db.WithContext(ctx).Clauses(clause.OnConflict{
		DoUpdates: clause.Assignments(map[string]any{
			// 这代表的是关注了-取消了-再关注了
			"status": FollowRelationStatusActive,
			"utime":  now,
		}),
	}).Create(&f).Error
}

func NewGORMFollowRelationDAO(db *gorm.DB) FollowRelationDAO {
	return &GORMFollowRelationDAO{db: db}
}
