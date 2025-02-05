package service

import (
	"context"
	"time"
	"webook/internal/domain"
	"webook/internal/repository"
	"webook/pkg/logger"
)

type CronJobService interface {
	Preempt(ctx context.Context) (domain.Job, error)
	ResetNextTime(ctx context.Context, j domain.Job) error
}

type cronJobService struct {
	repo            repository.CronJobRepository
	l               logger.V1
	refreshInterval time.Duration
}

func NewCronJobService(repo repository.CronJobRepository, l logger.V1) CronJobService {
	return &cronJobService{repo: repo, l: l, refreshInterval: time.Minute}
}

func (c *cronJobService) ResetNextTime(ctx context.Context, j domain.Job) error {
	nextTime := j.NextTime()
	return c.repo.UpdateNextTime(ctx, j.Id, nextTime)
}

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
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
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
