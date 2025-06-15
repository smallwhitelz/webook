package ioc

import (
	"github.com/spf13/viper"
	etcdv3 "go.etcd.io/etcd/client/v3"
)

func InitEtcd() *etcdv3.Client {
	type Config struct {
		EtcdAddr string `yaml:"etcdAddr"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	}
	var cfg Config
	err := viper.UnmarshalKey("etcd", &cfg)
	if err != nil {
		panic(err)
	}
	client, err := etcdv3.New(etcdv3.Config{
		Endpoints: []string{cfg.EtcdAddr},
		Username:  cfg.Username,
		Password:  cfg.Password,
	})
	if err != nil {
		panic(err)
	}
	return client
}
