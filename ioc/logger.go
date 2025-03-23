package ioc

import (
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"time"
	"webook/pkg/logger"
)

func InitLoggerV2() logger.V1 {
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

func InitLogger() logger.V1 {
	// 这里我们用一个小技巧，
	// 就是直接使用 zap 本身的配置结构体来处理
	// 配置Lumberjack以支持日志文件的滚动
	lumberjackLogger := &lumberjack.Logger{
		// 注意有没有权限
		Filename:   "/Users/ericzl/logs/user.log", // 指定日志文件路径
		MaxSize:    50,                            // 每个日志文件的最大大小，单位：MB
		MaxBackups: 3,                             // 保留旧日志文件的最大个数
		MaxAge:     28,                            // 保留旧日志文件的最大天数
		Compress:   true,                          // 是否压缩旧的日志文件
	}

	// 创建zap日志核心
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(lumberjackLogger),
		zapcore.DebugLevel, // 设置日志级别
	)

	l := zap.New(core, zap.AddCaller())
	res := logger.NewZapLogger(l)
	go func() {
		// 为了演示 ELK，我直接输出日志
		ticker := time.NewTicker(time.Millisecond * 1000)
		for t := range ticker.C {
			res.Info("模拟输出日志", logger.String("time", t.String()))
		}
	}()
	return res
}
