package job

import (
	"context"
	"github.com/google/uuid"
	rlock "github.com/gotomicro/redis-lock"
	"github.com/hashicorp/go-multierror"
	"github.com/redis/go-redis/v9"
	"sync"
	"sync/atomic"
	"time"
	"webook/internal/service"
	"webook/pkg/logger"
)

type RankingJobV1 struct {
	svc     service.RankingService
	l       logger.V1
	timeout time.Duration
	client  *rlock.Client
	key     string

	localLock *sync.Mutex
	lock      *rlock.Lock

	// 作业提示
	// 随机生成一个，就代表当前负载。你可以每隔一分钟生成一个
	load *atomic.Int32

	// 作业使用的
	nodeID         string
	redisClient    redis.Cmdable
	rankingLoadKey string
	closeSignal    chan struct{}
	loadTicker     *time.Ticker
}

func NewRankingJobV1(
	svc service.RankingService,
	l logger.V1,
	client *rlock.Client,
	timeout time.Duration,
	redisClient redis.Cmdable,
	loadInterval time.Duration,
) *RankingJobV1 {
	res := &RankingJobV1{svc: svc,
		key:       "job:ranking",
		l:         l,
		client:    client,
		localLock: &sync.Mutex{},
		timeout:   timeout,

		// 作业用的
		// 为自己生成一个 UUID
		nodeID:         uuid.New().String(),
		redisClient:    redisClient,
		rankingLoadKey: "ranking_job_nodes_load",
		load:           &atomic.Int32{},
		closeSignal:    make(chan struct{}),
		loadTicker:     time.NewTicker(loadInterval),
	}
	// 开启
	res.loadCycle()
	return res
}

func (r *RankingJobV1) Name() string {
	return "ranking"
}

// go fun() { r.Run()}

// RunV1 2024.1.23 答疑 在这个地方，返回下一次的执行时间，或者说，任务调度框架，
// 会在我这里返回之后，才开始计时
//func (r *RankingJobV1) RunV1() error {
//	return nil
//}

func (r *RankingJobV1) Run() error {
	r.localLock.Lock()
	lock := r.lock
	r.localLock.Unlock()
	if lock == nil {
		// 我能不能在这里，看一眼我是不是负载最低的，如果是，我就尝试获取分布式锁
		// 如果我的负载低于 70% 的节点

		// 抢分布式锁
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*4)
		defer cancel()
		lock, err := r.client.Lock(ctx, r.key, r.timeout,
			&rlock.FixIntervalRetry{
				Interval: time.Millisecond * 100,
				Max:      3,
				// 重试的超时
			}, time.Second)
		if err != nil {
			//r.l.Warn("获取分布式锁失败", logger.Error(err))
			return nil
		}
		r.l.Debug(r.nodeID + "获得了分布式锁 ")
		r.lock = lock
		go func() {
			// 并不是非得一半就续约
			// 如果是自己手写的自动续约，那么可以在续约的时候检查一下负载
			er := lock.AutoRefresh(r.timeout/2, r.timeout)
			if er != nil {
				// 续约失败了
				// 你也没办法中断当下正在调度的热榜计算（如果有）
				r.localLock.Lock()
				r.lock = nil
				//lock.Unlock()
				r.localLock.Unlock()
			}
		}()
	}
	// 这边就是你拿到了锁
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()
	// 如果 topN 是分步骤的。比如说分成了三个步骤：
	return r.svc.TopN(ctx)
}

func (r *RankingJobV1) loadCycle() {
	go func() {
		for range r.loadTicker.C {
			// 上报负载
			r.reportLoad()
			r.releaseLockIfNeed()
		}
	}()
}

func (r *RankingJobV1) releaseLockIfNeed() {
	// 检测自己是不是负载最低，如果不是，那么就直接释放分布式锁。
	r.localLock.Lock()
	lock := r.lock
	r.localLock.Unlock()
	if lock != nil {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()
		// 最低负载的
		// 这里是有一个优化的
		// 你可以说，
		// 1. 如果我的负载高于平均值，就释放分布式锁
		// 2. 如果我的负载高于中位数，就释放分布式锁
		// 3. 如果我的负载高于 70%，就释放分布式锁
		res, err := r.redisClient.ZPopMin(ctx, r.rankingLoadKey).Result()
		if err != nil {
			// 记录日志
			return
		}
		head := res[0]
		if head.Member.(string) != r.nodeID {
			// 不是自己，释放锁
			r.l.Debug(r.nodeID+" 不是负载最低的节点，释放分布式锁",
				logger.Field{Key: "head", Val: head})
			r.localLock.Lock()
			r.lock = nil
			r.localLock.Unlock()
			lock.Unlock(ctx)
		}
	}
}

// 上报负载
func (r *RankingJobV1) reportLoad() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	load := r.load.Load()
	r.l.Debug(r.nodeID+" 上报负载: ", logger.Int32("load", load))
	r.redisClient.ZAdd(ctx, r.rankingLoadKey,
		redis.Z{Member: r.nodeID, Score: float64(load)})
	cancel()
	return
}

func (r *RankingJobV1) Close() error {
	r.localLock.Lock()
	lock := r.lock
	r.localLock.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var err *multierror.Error
	if lock != nil {
		err = multierror.Append(err, lock.Unlock(ctx))
	}
	if r.loadTicker != nil {
		r.loadTicker.Stop()
	}
	// 删除自己的负载
	err = multierror.Append(err, r.redisClient.ZRem(ctx, r.rankingLoadKey, redis.Z{Member: r.nodeID}).Err())
	return err
}
