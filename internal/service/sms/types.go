package sms

import "context"

// Service 发送短信的抽象
// 屏蔽不同供应商之间的区别
// 这里发送短信的所有第三方服务治理都是遵循：开闭原则，非侵入式，所以引用了装饰器模式
//
//go:generate mockgen -source=./types.go -package=smsmocks -destination=./mocks/sms.mock.go Service
type Service interface {
	Send(ctx context.Context, tplId string, args []string, numbers ...string) error
}
