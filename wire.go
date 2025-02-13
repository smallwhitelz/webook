//go:build wireinject

package main

import (
	"github.com/google/wire"
	repository2 "webook/interactive/repository"
	cache2 "webook/interactive/repository/cache"
	dao2 "webook/interactive/repository/dao"
	service2 "webook/interactive/service"
	"webook/internal/events/article"
	"webook/internal/repository"
	"webook/internal/repository/cache"
	"webook/internal/repository/dao"
	"webook/internal/service"
	"webook/internal/web"
	ijwt "webook/internal/web/jwt"
	"webook/ioc"
)

var interactiveSvcSet = wire.NewSet(dao2.NewGORMInteractiveDAO,
	cache2.NewInteractiveRedisCache,
	repository2.NewCachedInteractiveRepository,
	service2.NewInteractiveService)

var rankingSvcSet = wire.NewSet(
	cache.NewRankingRedisCache,
	repository.NewCachedRankingRepository,
	service.NewBatchRankingService,
)

func InitWebServer() *App {
	wire.Build(
		// 第三方依赖
		ioc.InitDB, ioc.InitRedis,
		ioc.InitLogger,
		ioc.InitSaramaClient,
		ioc.InitSyncProducer,
		ioc.InitRlockClient,
		ioc.InitEtcd,
		// dao部分
		dao.NewUserDao,
		dao.NewArticleGORMDAO,

		//interactiveSvcSet,
		//ioc.InitIntrClient,
		ioc.InitIntrClientV1,
		rankingSvcSet,
		ioc.InitRankingJob,
		ioc.InitJobs,
		article.NewSaramaSyncProducer,
		//events.NewInteractiveReadEventConsumer,
		ioc.InitConsumers,

		// cache部分
		cache.NewRedisCodeCache,
		cache.NewUserCache,
		cache.NewArticleRedisCache,
		// repository部分
		repository.NewCachedUserRepository,
		repository.NewCodeRepository,
		repository.NewCachedArticleRepository,

		// service部分
		ioc.InitSMSService,
		ioc.InitWechatService,
		service.NewUserService,
		service.NewCodeService,
		service.NewArticleService,

		// handler部分
		web.NewUserHandler,
		web.NewArticleHandler,
		ijwt.NewRedisJWTHandler,
		web.NewOAuth2WechatHandler,
		ioc.InitGinMiddlewares,
		ioc.InitWebServer,

		wire.Struct(new(App), "*"),
	)
	return new(App)
}
