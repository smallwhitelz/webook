package ioc

import (
	"context"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	prometheus2 "github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
	otelgin "go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"strings"
	"time"
	"webook/internal/web"
	ijwt "webook/internal/web/jwt"
	"webook/internal/web/middleware"
	"webook/pkg/ginx"
	"webook/pkg/ginx/middleware/prometheus"
	"webook/pkg/ginx/middleware/ratelimit"
	"webook/pkg/limiter"
	"webook/pkg/logger"
)

//func InitWebServerV1(mdls []gin.HandlerFunc, hdls []web.Handler) *gin.Engine {
//	server := gin.Default()
//	server.Use(mdls...)
//	for _, hdl := range hdls {
//		hdl.RegisterRoutes(server)
//	}
//
//	return server
//}

func InitWebServer(mdls []gin.HandlerFunc, userHdl *web.UserHandler,
	artHdl *web.ArticleHandler,
	wechatHdl *web.OAuth2WechatHandler) *gin.Engine {
	//gin.SetMode(gin.ReleaseMode)
	server := gin.Default()
	//server := gin.New()
	server.Use(mdls...)
	userHdl.RegisterRoutes(server)
	wechatHdl.RegisterRoutes(server)
	artHdl.RegisterRoutes(server)
	return server
}

func InitGinMiddlewares(redisClient redis.Cmdable, hdl ijwt.Handler, l logger.LoggerV1) []gin.HandlerFunc {
	pb := &prometheus.Builder{
		Namespace: "geektime_zl",
		Subsystem: "webook",
		Name:      "gin_http",
		Help:      "这是一个统计 GIN 的http接口数据",
	}
	ginx.InitCount(prometheus2.CounterOpts{
		Namespace: "geektime_zl",
		Subsystem: "webook",
		Name:      "biz_code",
		Help:      "统计业务错误码	",
	})
	ginx.SetLogger(l)
	return []gin.HandlerFunc{
		cors.New(cors.Config{
			//AllowAllOrigins: true, 所有请求都允许
			//AllowOrigins:     []string{"http://localhost:3000"}, // 允许访问的域
			// 是否允许cookie之类的东西要不要带过来
			AllowCredentials: true,
			AllowHeaders:     []string{"Content-Type", "Authorization"},
			// 一般不要去配，允许所有方法基本不会有什么危险
			//AllowMethods: []string{"POST"},
			//允许前端访问你的后端响应中带的头部
			ExposeHeaders: []string{"x-jwt-token", "x-refresh-token"},
			AllowOriginFunc: func(origin string) bool {
				if strings.HasPrefix(origin, "http://localhost") {
					//if strings.Contains(origin, "localhost") 都可以
					return true
				}
				return strings.Contains(origin, "your_company.com")
			},
			MaxAge: 12 * time.Hour,
		}), func(ctx *gin.Context) {
			println("这是我的middleware")
		},
		pb.BuildResponseTime(),
		pb.BuildActiveRequest(),
		otelgin.Middleware("webook"),
		ratelimit.NewBuilder(limiter.NewRedisSlidingWindowLimiter(redisClient, time.Second, 1000)).Build(),
		middleware.NewLogMiddlewareBuilder(
			func(ctx context.Context, al middleware.AccessLog) {
				l.Debug("", logger.Field{Key: "req", Val: al})
			}).AllowReqBody().AllowRespBody().Build(),
		middleware.NewLoginJWTMiddlewareBuilder(hdl).CheckLogin(),
	}
}
