package sms

import "context"

// Service 发送短信的抽象
// 屏蔽不同供应商之间的区别
type Service interface {
	Send(ctx context.Context, tplId string, args []string, numbers ...string) error
}
