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

func (g *GORMFollowRelationDAO) CntFollow(ctx context.Context, uid int64) (int64, error) {
	var res int64
	err := g.db.WithContext(ctx).Select("count(follower)").
		// 如果没有额外的索引，绝对是全表扫描
		// 可以考虑 followee上建一个额外索引
		Where("followee = ? AND status = ?", uid, FollowRelationStatusActive).Count(&res).Error
	return res, err
}

func (g *GORMFollowRelationDAO) CntFollowee(ctx context.Context, uid int64) (int64, error) {
	var res int64
	err := g.db.WithContext(ctx).
		Select("count(followee)").
		// 可以命中索引，因为我们的索引是<follower,followee>
		Where("follower = ? AND status = ?", uid, FollowRelationStatusActive).Count(&res).Error
	return res, err
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
	// 如果用了之前interactive的方式去实现关注、粉丝数量的功能，也就是如果初始化了FollowStatics这个表
	// 在这里更新 FollowStatis 的计数（也是 upsert）可以用事务
}

func NewGORMFollowRelationDAO(db *gorm.DB) FollowRelationDAO {
	return &GORMFollowRelationDAO{db: db}
}
