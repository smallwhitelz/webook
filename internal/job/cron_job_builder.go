package job

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/robfig/cron/v3"
	"strconv"
	"time"
	"webook/pkg/logger"
)

type CronJobBuilder struct {
	l      logger.LoggerV1
	vector *prometheus.SummaryVec
}

func NewCronJobBuilder(l logger.LoggerV1, opt prometheus.SummaryOpts) *CronJobBuilder {
	vector := prometheus.NewSummaryVec(opt, []string{"job", "success"})
	return &CronJobBuilder{
		l:      l,
		vector: vector,
	}
}

func (c *CronJobBuilder) Build(job Job) cron.Job {
	name := job.Name()
	return cronJobAdapterFun(func() {
		// 还可以接入tracing
		start := time.Now()
		c.l.Debug("开始运行",
			logger.String("name", name))
		err := job.Run()
		if err != nil {
			c.l.Error("执行失败",
				logger.Error(err),
				logger.String("name", name))
		}
		c.l.Debug("结束运行",
			logger.String("name", name))
		duration := time.Since(start).Milliseconds()
		c.vector.WithLabelValues(name, strconv.FormatBool(err == nil)).Observe(float64(duration))
	})
}

type cronJobAdapterFun func()

func (c cronJobAdapterFun) Run() {
	c()
}
