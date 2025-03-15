package dao

import (
	"context"
	"database/sql"
	"gorm.io/gorm"
)

type CommentDAO interface {
	Insert(ctx context.Context, u Comment) error
	// Delete 删除本节点和其对应的子节点
	Delete(ctx context.Context, comment Comment) error
	FindByBiz(ctx context.Context, biz string, bizId int64, minId int64, limit int64) ([]Comment, error)
	FindRepliesByPid(ctx context.Context, pid int64, offset, limit int) ([]Comment, error)
	FindRepliesByRid(ctx context.Context, rid int64, maxID int64, limit int64) ([]Comment, error)
}

type Comment struct {
	Id int64 `gorm:"primaryKey,autoIncrement"`
	// 发表评论的人
	// 也就是说，如果有需要查询某个人发表过的所有评论，那么需要在这里建立一个索引
	Uid int64
	// 被评价的东西
	Biz     string `gorm:"index:biz_type_id"`
	BizID   int64  `gorm:"index:biz_type_id"`
	Content string

	// 我的根评论是那个
	// 也就是说，如果这个字段是NULL 他就是根评论
	RootID sql.NullInt64 `gorm:"column:root_id,index"`

	// 这个是NULL 也是根评论
	PID sql.NullInt64 `gorm:"column:pid,index"`

	// 外键 用于级联删除
	ParentComment *Comment `gorm:"ForeignKey:PID;AssociationForeignKey:ID;constraint:OnDelete:CASCADE"`

	Ctime int64
	// 事实上，大部分平台是不允许修改评论的
	Utime int64
}

func (*Comment) TableName() string {
	return "comments"
}

type GORMCommentDAO struct {
	db *gorm.DB
}

func (c *GORMCommentDAO) FindRepliesByRid(ctx context.Context, rid int64, maxID int64, limit int64) ([]Comment, error) {
	var res []Comment
	err := c.db.WithContext(ctx).
		Where("root_id = ? AND id > ?", rid, maxID).
		Order("id ASC").
		Limit(int(limit)).Find(&res).Error
	return res, err
}

// FindRepliesByPid 查找评论的直接评论
func (c *GORMCommentDAO) FindRepliesByPid(ctx context.Context, pid int64, offset, limit int) ([]Comment, error) {
	var res []Comment
	err := c.db.WithContext(ctx).Where("pid = ?", pid).
		Order("id DESC").
		Offset(offset).Limit(limit).Find(&res).Error
	return res, err
}

func (c *GORMCommentDAO) FindByBiz(ctx context.Context, biz string, bizId int64, minId int64, limit int64) ([]Comment, error) {
	var res []Comment
	err := c.db.WithContext(ctx).Where("biz = ? AND biz_id = ? AND id < ? AND pid IS NULL", biz, bizId, minId).
		Limit(int(limit)).Find(&res).Error
	return res, err
}

func (c *GORMCommentDAO) Insert(ctx context.Context, u Comment) error {
	return c.db.
		WithContext(ctx).
		Create(&u).
		Error
}

func (c *GORMCommentDAO) Delete(ctx context.Context, comment Comment) error {
	return c.db.WithContext(ctx).Delete(&Comment{
		Id: comment.Id,
	}).Error
}

func NewGORMCommentDAO(db *gorm.DB) CommentDAO {
	return &GORMCommentDAO{db: db}
}
