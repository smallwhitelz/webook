package main

import (
	"bytes"
	"context"
	"github.com/fsnotify/fsnotify"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
	"go.uber.org/zap"
	"log"
	"net/http"
	"time"
	"webook/ioc"
)

func main() {
	//InitViperV1()
	InitViperWatch()
	InitLogger()
	tpCancel := ioc.InitOTEL()
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		// 关闭服务器清理资源也是耗时的，这里超时直接shutdown掉trace
		tpCancel(ctx)
	}()
	app := InitWebServer()
	initPrometheus()
	for _, c := range app.consumers {
		err := c.Start()
		if err != nil {
			panic(err)
		}
	}
	//app.cron.Start()
	//defer func() {
	//	// 等待定时任务退出
	//	<-app.cron.Stop().Done()
	//}()
	server := app.server
	server.GET("/hello", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "hello,访问成功！")
	})
	//server.Run(":8081")
	server.Run(":8080")

}

func initPrometheus() {
	go func() {
		// 专门给 prometheus 用的端口
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":8081", nil)
	}()
}

func InitLogger() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	zap.ReplaceGlobals(logger)
}

func InitViper() {
	// 读取的文件名称
	viper.SetConfigName("dev")
	// 读取的文件类型
	viper.SetConfigType("yaml")
	// 当前工作目录的 config 子目录
	viper.AddConfigPath("config")
	// 读取配置
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	log.Println(viper.Get("test.key"))
}

// InitViperV1 利用viper读取启动参数Program arguments
func InitViperV1() {
	cfile := pflag.String("config", "config/dev.yaml", "配置文件路径")
	// 这一步后 cfile才有值
	pflag.Parse()
	viper.SetConfigType("yaml")
	viper.SetConfigFile(*cfile)
	// 读取配置
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	log.Println(viper.Get("test.key"))
}

// InitViperV2 比如在测试或者本地调试的时候，我们懒得写配置文件，就可以在Go中手写这个配置，然后传给viper
// ps：都手写了，可以直接写死在ioc里
func InitViperV2() {
	cfg := `
test:
  key: 123

redis:
  addr: "43.154.97.245:6379"

db:
  dsn: "root:root@tcp(43.154.97.245:13316)/webook"
`
	viper.SetConfigType("yaml")
	err := viper.ReadConfig(bytes.NewReader([]byte(cfg)))
	if err != nil {
		panic(err)
	}
}

// InitViperWatch 监听配置变更，场景比如为功能A设置一个开关，最开始开启A，一旦A有问题，直接关掉
func InitViperWatch() {
	cfile := pflag.String("config", "config/dev.yaml", "配置文件路径")
	// 这一步后 cfile才有值
	pflag.Parse()
	viper.SetConfigType("yaml")
	viper.SetConfigFile(*cfile)
	// 有严格的顺序要求，一定在set，add等方法之后调用
	viper.WatchConfig()
	// 读取配置
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	log.Println(viper.Get("test.key"))
}

// initViperRemote 远程配置中心
func initViperRemote() {
	// 如果etcd有密码，那这里会报错，社区不支持读取etcd密码
	err := viper.AddRemoteProvider("etcd3", "http://43.154.97.245:12379", "/webook")
	if err != nil {
		panic(err)
	}
	viper.SetConfigType("yaml")
	viper.OnConfigChange(func(in fsnotify.Event) {
		log.Println("远程配置中心发生变更")
	})
	err = viper.ReadRemoteConfig()
	if err != nil {
		panic(err)
	}
	// 这一步如果在ReadRemoteConfig前面，会出现并发读写的问题，就会直接panic
	go func() {
		for {
			// 监听远程配置中心变更
			err = viper.WatchRemoteConfig()
			if err != nil {
				panic(err)
			}
			log.Println("watch", viper.GetString("test.key"))
			time.Sleep(time.Second * 3)
		}
	}()

}

//func InitUser(db *gorm.DB, redisClient redis.Cmdable, codeSvc service.CodeService, server *gin.Engine) {
//	ud := dao.NewUserDao(db)
//	uc := cache.NewUserCache(redisClient)
//	ur := repository.NewCachedUserRepository(ud, uc)
//	us := service.NewUserService(ur)
//	hdl := web.NewUserHandler(us, codeSvc)
//	hdl.RegisterRoutes(server)
//}
//
//func initCodeSvc(redisClient redis.Cmdable) *service.codeService {
//	cc := cache.NewCodeCache(redisClient)
//	crepo := repository.NewCodeRepository(cc)
//	return service.NewCodeService(crepo, initMemorySms())
//}

//func initMemorySms() sms.Service {
//	return localsms.NewService()
//}
//
//func initWebServer() *gin.Engine {
//	server := gin.Default()
//
//	useJWT(server)
//	return server
//}
//
//func useJWT(server *gin.Engine) {
//	login := &middleware.LoginJWTMiddlewareBuilder{}
//	server.Use(login.CheckLogin())
//}
//
//func useSession(server *gin.Engine) {
//	login := &middleware.LoginMiddlewareBuilder{}
//	// 存储数据，也就是你的userId存哪里
//	// 直接存在cookie
//	store := cookie.NewStore([]byte("secret"))
//	// 基于内存的实现
//	//store := memstore.NewStore([]byte("XCptI5cGK8etly19icar00Rk9klXGUai"), []byte("mxyp9dp08X0V9waVqk7hbGs4lERuxwc1"))
//	//store, err := redis.NewStore(16, "tcp", "43.154.97.245:6379", "",
//	//	[]byte("XCptI5cGK8etly19icar00Rk9klXGUai"), []byte("mxyp9dp08X0V9waVqk7hbGs4lERuxwc1"))
//	//if err != nil {
//	//	panic(err)
//	//}
//	server.Use(sessions.Sessions("ssid", store), login.CheckLogin())
//}
