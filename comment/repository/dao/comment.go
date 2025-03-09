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
