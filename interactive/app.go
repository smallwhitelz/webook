package main

import (
	"webook/internal/events"
	"webook/pkg/ginx"
	"webook/pkg/grpcx"
)

type App struct {
	consumers   []events.Consumer
	server      *grpcx.Server
	adminServer *ginx.Server
	// 没有封装grpc工具包的写法
	//server *grpc.InteractiveServiceServer
}
