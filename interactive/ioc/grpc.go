package ioc

import (
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	grpc2 "webook/interactive/grpc"
	"webook/pkg/grpcx"
)

func NewGrpcxServer(intrSvc *grpc2.InteractiveServiceServer) *grpcx.Server {
	s := grpc.NewServer()
	// 这里是我们封装的反向去调
	intrSvc.Register(s)
	// 也可以这样去注册，这样就是需要一个一个去调
	//intrv1.RegisterInteractiveServiceServer(s,intrSvc)
	return &grpcx.Server{
		Server: s,
		Addr:   viper.GetString("grpc.server.addr"),
	}
}
