package wego

import (
	"webook/pkg/ginx"
	"webook/pkg/grpcx"
	"webook/pkg/samarax"
)

// App 当你在 wire 里面使用这个结构体的时候，要注意不是所有的服务都需要全部字段，
// 那么在 wire 的时候就不要使用 * 了
type App struct {
	GRPCServer *grpcx.Server
	WebServer  *ginx.Server
	Consumers  []samarax.Consumer
}
