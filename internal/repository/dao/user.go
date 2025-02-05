package dao

import (
	"context"
	"database/sql"
	"errors"
	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"time"
)

var (
	ErrDuplicateEmail = errors.New("邮箱已存在，请勿重复注册")
	ErrRecordNotFound = gorm.ErrRecordNotFound
)

type UserDAO interface {
	Insert(ctx context.Context, user User) error
	FindByEmail(ctx context.Context, email string) (User, error)
	FindById(ctx context.Context, uid int64) (User, error)
	UpdateById(ctx context.Context, entity User) error
	FindByPhone(ctx context.Context, phone string) (User, error)
	FindByWechat(ctx context.Context, openId string) (User, error)
}

type GORMUserDAO struct {
	db *gorm.DB
}

func (dao *GORMUserDAO) FindByWechat(ctx context.Context, openId string) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("wechat_open_id = ?", openId).First(&u).Error
	return u, err
}

func NewUserDao(db *gorm.DB) UserDAO {
	return &GORMUserDAO{db: db}
}

func (dao *GORMUserDAO) Insert(ctx context.Context, user User) error {
	now := time.Now().UnixMilli()
	user.Ctime = now
	user.Utime = now
	err := dao.db.WithContext(ctx).Create(&user).Error
	if me, ok := err.(*mysql.MySQLError); ok {
		const duplicateErr uint16 = 1062
		if me.Number == duplicateErr {
			// 用户冲突：唯一索引冲突：邮箱冲突
			return ErrDuplicateEmail
		}
	}
	return err
}

func (dao *GORMUserDAO) FindByEmail(ctx context.Context, email string) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("email = ?", email).First(&u).Error
	return u, err
}

func (dao *GORMUserDAO) FindById(ctx context.Context, uid int64) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("id = ?", uid).First(&u).Error
	return u, err
}

func (dao *GORMUserDAO) UpdateById(ctx context.Context, entity User) error {
	return dao.db.WithContext(ctx).Model(&entity).Where("id = ?", entity.Id).Updates(map[string]any{
		"nickname":    entity.Nickname,
		"birthday":    entity.Birthday,
		"description": entity.Description,
		"utime":       time.Now().UnixMilli(),
	}).Error
}

func (dao *GORMUserDAO) FindByPhone(ctx context.Context, phone string) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("phone = ?", phone).First(&u).Error
	return u, err
}

type User struct {
	Id int64 `gorm:"primaryKey,autoIncrement"`
	// 代表这是一个可以为NULL的列
	//Email *string
	Email    sql.NullString `gorm:"unique"`
	Password string

	Nickname string `gorm:"type=varchar(128)"`
	// YYYY-MM-DD
	Birthday    int64
	Description string `gorm:"type=varchar(4096)"`

	// 代表这是一个可以为NULL的列
	Phone sql.NullString `gorm:"unique"`

	// 1 如果查询要求同时使用openid和unionid，就要创建联合唯一索引
	// 2 如果查询只用openid，那么就openid上创建唯一索引或者 <openid,unionid>联合索引
	// 3 如果查询只用unionid，那么就在unionid上创建唯一索引或者 <unionid,openid>联合索引
	WechatOpenId  sql.NullString `gorm:"unique"`
	WechatUnionId sql.NullString

	// 时区 UTC 0 的毫秒数
	// 创建时间
	Ctime int64
	// 更新时间
	Utime int64
}
