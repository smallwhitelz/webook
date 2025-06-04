package main

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
	"log"
	"net/http"
)

func main() {
	InitViperV1()
	app := InitApp()
	initPrometheus()
	for _, c := range app.consumers {
		err := c.Start()
		if err != nil {
			panic(err)
		}
	}
	err := app.server.Serve()
	if err != nil {
		panic(err)
	}
	// 没有封装grpc工具包的写法
	// 缺陷是来一个grpc服务就要写出这样模版化的代码
	//server := grpc.NewServer()
	//intrv1.RegisterInteractiveServiceServer(server, app.server)
	//l, err := net.Listen("tcp", ":8090")
	//if err != nil {
	//	panic(err)
	//}
	//server.Serve(l)
}

func initPrometheus() {
	go func() {
		// 专门给 prometheus 用的端口
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":8081", nil)
	}()
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
