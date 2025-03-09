package ioc

import (
	"github.com/spf13/viper"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
	grpc2 "webook/interactive/grpc"
	"webook/pkg/grpcx"
	"webook/pkg/logger"
)

func NewGrpcxServer(intrSvc *grpc2.InteractiveServiceServer, client *clientv3.Client, l logger.V1) *grpcx.Server {
	type Config struct {
		Port    int    `yaml:"port"`
		Name    string `yaml:"name"`
		EtcdTTL int64  `yaml:"etcdTTL"`
	}
	s := grpc.NewServer()
	intrSvc.Register(s)
	var cfg Config
	err := viper.UnmarshalKey("grpc.server", &cfg)
	if err != nil {
		panic(err)
	}
	return &grpcx.Server{
		Server:     s,
		Port:       cfg.Port,
		EtcdTTL:    cfg.EtcdTTL,
		EtcdClient: client,
		Name:       cfg.Name,
		L:          l,
	}
}
