package ioc

import (
	"github.com/IBM/sarama"
	"github.com/gin-gonic/gin"
	prometheus2 "github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
	"webook/interactive/repository/dao"
	"webook/pkg/ginx"
	"webook/pkg/gormx/connpool"
	"webook/pkg/logger"
	"webook/pkg/migrator/events"
	"webook/pkg/migrator/events/fixer"
	"webook/pkg/migrator/scheduler"
)

func InitGinxServer(l logger.LoggerV1, src SrcDB, dst DstDB, p *connpool.DoubleWritePool, producer events.Producer) *ginx.Server {
	newScheduler := scheduler.NewScheduler[dao.Interactive](l, src, dst, p, producer)
	engine := gin.Default()
	group := engine.Group("/migrator")
	ginx.InitCount(prometheus2.CounterOpts{
		Namespace: "geektime_daming",
		Subsystem: "webook_intr_admin",
		Name:      "biz_code",
		Help:      "统计业务错误码",
	})
	newScheduler.RegisterRoutes(group)
	return &ginx.Server{
		Engine: engine,
		Addr:   viper.GetString("migrator.http.addr"),
	}
}

func InitInteractiveProducer(p sarama.SyncProducer) events.Producer {
	return events.NewSaramaProducer("inconsistent_interactive", p)
}

func InitFixerConsumer(client sarama.Client,
	l logger.LoggerV1,
	src SrcDB,
	dst DstDB) *fixer.Consumer[dao.Interactive] {
	res, err := fixer.NewConsumer[dao.Interactive](client, l, src, dst, "inconsistent_interactive")
	if err != nil {
		panic(err)
	}
	return res
}
