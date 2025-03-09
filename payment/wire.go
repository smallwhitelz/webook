//go:build wireinject

package main

import (
	"github.com/google/wire"
	"webook/payment/grpc"
	"webook/payment/ioc"
	"webook/payment/repository"
	"webook/payment/repository/dao"
	"webook/payment/web"
	"webook/pkg/wego"
)

func InitApp() *wego.App {
	wire.Build(
		ioc.InitEtcdClient,
		//ioc.InitKafka,
		//ioc.InitProducer,
		ioc.InitWechatClient,
		dao.NewPaymentGORMDAO,
		ioc.InitDB,
		repository.NewPaymentRepository,
		grpc.NewWechatServiceServer,
		ioc.InitWechatNativeService,
		ioc.InitWechatConfig,
		ioc.InitWechatNotifyHandler,
		ioc.InitGRPCServer,
		web.NewWechatHandler,
		ioc.InitGinServer,
		ioc.InitLogger,
		wire.Struct(new(wego.App), "WebServer", "GRPCServer"))
	return new(wego.App)
}
