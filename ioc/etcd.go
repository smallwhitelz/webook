package ioc

import (
	"github.com/spf13/viper"
	etcdv3 "go.etcd.io/etcd/client/v3"
)

func InitEtcd() *etcdv3.Client {
	type Config struct {
		Addrs []string
	}
	var cfg Config
	err := viper.UnmarshalKey("etcd", &cfg)
	if err != nil {
		panic(err)
	}
	cli, err := etcdv3.New(etcdv3.Config{
		Endpoints: cfg.Addrs,
		Username:  "root",
		Password:  "1234",
	})
	return cli
}
