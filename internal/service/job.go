package service

import (
	"context"
	"time"
	"webook/internal/domain"
	"webook/internal/repository"
	"webook/pkg/logger"
)

// CronJobService mysql的分布式任务调度实现
type CronJobService interface {
	Preempt(ctx context.Context) (domain.Job, error)
	ResetNextTime(ctx context.Context, j domain.Job) error
	// 这里也可以暴露Job整个的增删改查方法，让用户可以通过http接口对Job进行操作
}

type cronJobService struct {
	repo            repository.CronJobRepository
	l               logger.LoggerV1
	refreshInterval time.Duration
}

func NewCronJobService(repo repository.CronJobRepository, l logger.LoggerV1) CronJobService {
	return &cronJobService{repo: repo, l: l, refreshInterval: time.Minute}
}

func (c *cronJobService) ResetNextTime(ctx context.Context, j domain.Job) error {
	nextTime := j.NextTime()
	return c.repo.UpdateNextTime(ctx, j.Id, nextTime)
}

// Preempt 抢占
func (c *cronJobService) Preempt(ctx context.Context) (domain.Job, error) {
	j, err := c.repo.Preempt(ctx)
	if err != nil {
		return domain.Job{}, err
	}
	// 续约
	ticker := time.NewTicker(c.refreshInterval)
	go func() {
		for range ticker.C {
			c.refresh(j.Id)
		}
	}()
	j.CancelFunc = func() {
		ticker.Stop()
		// 单独起一个ctx，上面的抢占后执行很可能已经过去很久，所以不能用父ctx
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		// 释放锁
		err := c.repo.Release(ctx, j.Id)
		if err != nil {
			c.l.Error("释放 job 失败",
				logger.Int64("jid", j.Id),
				logger.Error(err))
		}
	}
	return j, err
}

func (c *cronJobService) refresh(id int64) {
	// 本质是更新一下时间
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err := c.repo.UpdateUtime(ctx, id)
	if err != nil {
		c.l.Error("续约失败", logger.Error(err), logger.Int64("jid", id))
	}
}
