package ioc

import (
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"webook/pkg/logger"
)

func InitLogger() logger.V1 {
	l, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	return logger.NewZapLogger(l)
}

// InitLoggerV1 两种写法都可以，这种可以控制生产环境NewProductionConfig和开发环境NewDevelopmentConfig
func InitLoggerV1() logger.V1 {
	cfg := zap.NewDevelopmentConfig()
	err := viper.UnmarshalKey("log", &cfg)
	if err != nil {
		panic(err)
	}
	l, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	return logger.NewZapLogger(l)
}
