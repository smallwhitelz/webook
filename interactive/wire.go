//go:build wireinject

package main

import (
	"github.com/google/wire"
	"webook/interactive/events"
	"webook/interactive/grpc"
	"webook/interactive/ioc"
	repository2 "webook/interactive/repository"
	cache2 "webook/interactive/repository/cache"
	dao2 "webook/interactive/repository/dao"
	service2 "webook/interactive/service"
)

var thirdPartySet = wire.NewSet(
	ioc.InitSrcDB, ioc.InitDstDB, ioc.InitBizDB, ioc.InitDoubleWritePool,
	ioc.InitLogger, ioc.InitRedis,
	ioc.InitSaramaClient, ioc.InitSaramaSyncProducer)

var interactiveSvcSet = wire.NewSet(dao2.NewGORMInteractiveDAO,
	cache2.NewInteractiveRedisCache,
	repository2.NewCachedInteractiveRepository,
	service2.NewInteractiveService,
)

func InitApp() *App {
	wire.Build(
		thirdPartySet,
		interactiveSvcSet,
		ioc.InitConsumers,
		ioc.NewGrpcxServer,
		ioc.InitGinxServer,
		ioc.InitInteractiveProducer,
		ioc.InitFixerConsumer,
		events.NewInteractiveReadEventConsumer,
		grpc.NewInteractiveServiceServer,
		wire.Struct(new(App), "*"))
	return new(App)
}
