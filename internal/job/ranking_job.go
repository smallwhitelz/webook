package job

import (
	"context"
	rlock "github.com/gotomicro/redis-lock"
	"sync"
	"time"
	"webook/internal/service"
	"webook/pkg/logger"
)

// RankingJob 利用分布式锁和开源cronjob框架去定时计算热榜数据
type RankingJob struct {
	svc service.RankingService
	l   logger.LoggerV1
	// ctx过期时间
	timeout time.Duration
	client  *rlock.Client
	key     string
	// 保护锁，因为有多个goroutine执行，也会有并发问题，所以用本地锁去保护分布式锁这里的逻辑
	localLock *sync.Mutex
	lock      *rlock.Lock
}

func NewRankingJob(svc service.RankingService, l logger.LoggerV1, client *rlock.Client, timeout time.Duration) *RankingJob {
	return &RankingJob{
		svc:       svc,
		l:         l,
		client:    client,
		key:       "job:ranking",
		localLock: &sync.Mutex{},
		timeout:   timeout,
	}
}

func (r *RankingJob) Name() string {
	return "ranking"
}

func (r *RankingJob) Run() error {
	r.localLock.Lock()
	lock := r.lock
	if lock == nil {
		// 抢分布式锁
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*4)
		defer cancel()
		lock, err := r.client.Lock(ctx, r.key, r.timeout, &rlock.FixIntervalRetry{
			Interval: time.Millisecond * 100,
			Max:      3,
			// 重试的超时
		}, time.Second)
		if err != nil {
			r.localLock.Unlock()
			r.l.Warn("获取分布式锁失败", logger.Error(err))
			return nil
		}
		r.lock = lock
		r.localLock.Unlock()
		// 分布式锁是有过期时间的，所以我们要进行续约
		go func() {
			// 并不是非得一半就续约
			er := lock.AutoRefresh(r.timeout/2, r.timeout)
			if er != nil {
				// 续约失败
				// 你也没办法中断当下正在调度的热榜计算（如果有）
				r.localLock.Lock()
				r.lock = nil
				r.localLock.Unlock()
			}
		}()
	}
	// 这边就是你拿到了锁
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()
	return r.svc.TopN(ctx)
}

// Close 主动释放锁
// 这个功能有点鸡肋，关机后，不会续约，分布式锁的超时时间一到，自然会释放锁
func (r *RankingJob) Close() error {
	r.localLock.Lock()
	lock := r.lock
	r.localLock.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return lock.Unlock(ctx)
}

// RunV1 这个分布式锁的写法的缺陷是同一时刻只有一个goroutine在计算热榜
// 会出现实例0计算完热榜释放锁紧接着实例1立刻拿到锁又去计算热榜
func (r *RankingJob) RunV1() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*4)
	defer cancel()
	lock, err := r.client.Lock(ctx, r.key,
		// 超过这个时间分布式锁也会释放
		r.timeout,
		// 重试机制
		&rlock.FixIntervalRetry{
			// 每隔100ms重试一次
			Interval: time.Millisecond * 100,
			// 最多重试3次
			Max: 3,
			// 每次重试的超时时间是1s
		}, time.Second)
	if err != nil {
		return err
	}
	defer func() {
		// 解锁
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		er := lock.Unlock(ctx)
		if er != nil {
			r.l.Error("ranking job 释放分布式锁失败", logger.Error(er))
		}
	}()
	ctx, cancel = context.WithTimeout(context.Background(), r.timeout)
	defer cancel()
	return r.svc.TopN(ctx)
}

// RunV2 一开始不考虑任何情况下实现的Job方法，直接掉热榜计算方法就可以
func (r *RankingJob) RunV2() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*4)
	defer cancel()
	return r.svc.TopN(ctx)
}
