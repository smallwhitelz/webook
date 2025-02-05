package cache

import (
	"context"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
	"webook/internal/repository/cache/redismocks"
)

func TestRedisCodeCache_Set(t *testing.T) {
	keyFunc := func(biz, phone string) string {
		return fmt.Sprintf("phone_code:%s:%s", biz, phone)
	}
	testCases := []struct {
		name  string
		mock  func(ctrl *gomock.Controller) redis.Cmdable
		ctx   context.Context
		biz   string
		phone string
		code  string

		wantErr error
	}{
		{
			name: "设置成功",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				res := redismocks.NewMockCmdable(ctrl)
				cmd := redis.NewCmd(context.Background())
				cmd.SetErr(nil)
				cmd.SetVal(int64(0))
				res.EXPECT().Eval(gomock.Any(), luaSetCode,
					[]string{keyFunc("test", "15212345678")}, []any{
						"123456",
					}).
					Return(cmd)
				return res
			},
			ctx:     context.Background(),
			biz:     "test",
			phone:   "15212345678",
			code:    "123456",
			wantErr: nil,
		},
		{
			name: "发送频繁",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				res := redismocks.NewMockCmdable(ctrl)
				cmd := redis.NewCmd(context.Background())
				cmd.SetErr(nil)
				cmd.SetVal(int64(-1))
				res.EXPECT().Eval(gomock.Any(), luaSetCode,
					[]string{keyFunc("test", "15212345678")}, []any{
						"123456",
					}).
					Return(cmd)
				return res
			},
			ctx:     context.Background(),
			biz:     "test",
			phone:   "15212345678",
			code:    "123456",
			wantErr: ErrCodeSendTooMany,
		},
		{
			name: "验证码不存在过期时间",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				res := redismocks.NewMockCmdable(ctrl)
				cmd := redis.NewCmd(context.Background())
				cmd.SetErr(nil)
				cmd.SetVal(int64(-2))
				res.EXPECT().Eval(gomock.Any(), luaSetCode,
					[]string{keyFunc("test", "15212345678")}, []any{
						"123456",
					}).
					Return(cmd)
				return res
			},
			ctx:     context.Background(),
			biz:     "test",
			phone:   "15212345678",
			code:    "123456",
			wantErr: errors.New("验证码存在，但是没有过期时间"),
		},
		{
			name: "redis错误",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				res := redismocks.NewMockCmdable(ctrl)
				cmd := redis.NewCmd(context.Background())
				cmd.SetErr(errors.New("redis 错误"))
				res.EXPECT().Eval(gomock.Any(), luaSetCode,
					[]string{keyFunc("test", "15212345678")}, []any{
						"123456",
					}).
					Return(cmd)
				return res
			},
			ctx:     context.Background(),
			biz:     "test",
			phone:   "15212345678",
			code:    "123456",
			wantErr: errors.New("redis 错误"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			c := NewRedisCodeCache(tc.mock(ctrl))
			err := c.Set(context.Background(), tc.biz, tc.phone, tc.code)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}
