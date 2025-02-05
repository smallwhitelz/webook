package dao

import (
	"context"
	"database/sql"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	mysqlDriver "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"testing"
	"time"
)

func TestTime(t *testing.T) {
	now := time.Now().UnixMilli()
	log.Println(now)
	normal := time.Unix(0, now*int64(time.Millisecond))
	format := normal.Format("2006-01-02 15:04:05")
	log.Println(format)
}

func TestGORMUserDAO_Insert(t *testing.T) {
	testCases := []struct {
		name string
		mock func(t *testing.T) *sql.DB
		ctx  context.Context
		user User

		wantErr error
	}{
		{
			name: "插入成功",
			mock: func(t *testing.T) *sql.DB {
				db, mock, err := sqlmock.New()
				assert.NoError(t, err)
				mockRes := sqlmock.NewResult(123, 1)
				// 这边要求传的是sql的正则表达式
				mock.ExpectExec("INSERT INTO .*").WillReturnResult(mockRes)
				return db
			},
			ctx: context.Background(),
			user: User{
				Nickname: "Tom",
			},
		},
		{
			name: "邮箱冲突",
			mock: func(t *testing.T) *sql.DB {
				db, mock, err := sqlmock.New()
				assert.NoError(t, err)
				// 这边要求传的是sql的正则表达式
				mock.ExpectExec("INSERT INTO .*").WillReturnError(
					&mysqlDriver.MySQLError{
						Number: 1062,
					})
				return db
			},
			ctx: context.Background(),
			user: User{
				Nickname: "Tom",
			},
			wantErr: ErrDuplicateEmail,
		},
		{
			name: "数据库错误",
			mock: func(t *testing.T) *sql.DB {
				db, mock, err := sqlmock.New()
				assert.NoError(t, err)
				// 这边要求传的是sql的正则表达式
				mock.ExpectExec("INSERT INTO .*").WillReturnError(errors.New("db 错误"))
				return db
			},
			ctx: context.Background(),
			user: User{
				Nickname: "Tom",
			},
			wantErr: errors.New("db 错误"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sqlDB := tc.mock(t)
			db, err := gorm.Open(mysql.New(mysql.Config{
				Conn:                      sqlDB,
				SkipInitializeWithVersion: true,
			}), &gorm.Config{
				DisableAutomaticPing:   false,
				SkipDefaultTransaction: true,
			})
			assert.NoError(t, err)
			dao := NewUserDao(db)
			err = dao.Insert(tc.ctx, tc.user)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}
