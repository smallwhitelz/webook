//go:build wireinject

package startup

import (
	"github.com/google/wire"
	"webook/interactive/grpc"
	repository2 "webook/interactive/repository"
	cache2 "webook/interactive/repository/cache"
	dao2 "webook/interactive/repository/dao"
	service2 "webook/interactive/service"
)

var thirdPartySet = wire.NewSet( // 第三方依赖
	InitRedis, InitDB,
	//InitSaramaClient,
	//InitSyncProducer,
	InitLogger,
)

var interactiveSvcSet = wire.NewSet(dao2.NewGORMInteractiveDAO,
	cache2.NewInteractiveRedisCache,
	repository2.NewCachedInteractiveRepository,
	service2.NewInteractiveService,
)

func InitInteractiveService() *grpc.InteractiveServiceServer {
	wire.Build(thirdPartySet, interactiveSvcSet, grpc.NewInteractiveServiceServer)
	return new(grpc.InteractiveServiceServer)
}

//func InitInteractiveService() service2.InteractiveService {
//	wire.Build(thirdPartySet, interactiveSvcSet)
//	return service2.NewInteractiveService(nil)
//}
