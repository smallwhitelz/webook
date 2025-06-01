package job

import (
	"context"
	"fmt"
	"golang.org/x/sync/semaphore"
	"time"
	"webook/internal/domain"
	"webook/internal/service"
	"webook/pkg/logger"
)

// Executor 执行器，任务执行器
// 这个相当于调度到某个节点上后用什么去执行这个job
type Executor interface {
	Name() string
	// Exec ctx 这个是全局控制，Executor 的实现者主要要正确处理 ctx 的超时或者取消
	Exec(ctx context.Context, j domain.Job) error
}

// LocalFuncExecutor 调用本地方法的
// 这里相当于本地local去执行测试里面一个叫做test_job的job
// 这个job做了一个Do的动作
type LocalFuncExecutor struct {
	funcs map[string]func(ctx context.Context, j domain.Job) error
}

func NewLocalFuncExecutor() *LocalFuncExecutor {
	return &LocalFuncExecutor{funcs: map[string]func(ctx context.Context, j domain.Job) error{}}
}

func (l *LocalFuncExecutor) Name() string {
	return "local"
}

func (l *LocalFuncExecutor) RegisterFunc(name string, fn func(ctx context.Context, j domain.Job) error) {
	l.funcs[name] = fn

}

func (l *LocalFuncExecutor) Exec(ctx context.Context, j domain.Job) error {
	fn, ok := l.funcs[j.Name]
	if !ok {
		return fmt.Errorf("未注册本地方法 %s", j.Name)
	}
	return fn(ctx, j)
}

// Scheduler 用来调度mysql实现的分布式调度
type Scheduler struct {
	dbTimeout time.Duration

	svc service.CronJobService

	executors map[string]Executor

	l logger.LoggerV1

	limiter *semaphore.Weighted
}

func NewScheduler(svc service.CronJobService, l logger.LoggerV1) *Scheduler {
	return &Scheduler{
		svc:       svc,
		dbTimeout: time.Second,
		// 最多同时运行100个
		limiter:   semaphore.NewWeighted(100),
		l:         l,
		executors: map[string]Executor{},
	}
}

func (s *Scheduler) RegisterExecutor(exec Executor) {
	s.executors[exec.Name()] = exec
}

func (s *Scheduler) Schedule(ctx context.Context) error {
	for {
		// 放弃调度了
		if ctx.Err() != nil {
			return ctx.Err()
		}
		// 这边一直抢占，有可能抢占几百万个，所以我们要限制住他
		// 进来直接limiter，抢占到就拿一个令牌
		err := s.limiter.Acquire(ctx, 1)
		if err != nil {
			return err
		}
		dbCtx, cancel := context.WithTimeout(ctx, s.dbTimeout)
		// 开始抢占
		j, err := s.svc.Preempt(dbCtx)
		cancel()
		if err != nil {
			// 有 ERROR，意思是没抢到
			// 最简单的做法就是直接下一轮，也可以睡一段时间
			continue
		}

		// 肯定要调度执行 j
		exec, ok := s.executors[j.Executor]
		// 没找到
		if !ok {
			// 你可以直接中断，也可以下一轮
			s.l.Error("找不到执行器",
				logger.Int64("jid", j.Id),
				logger.String("executor", j.Executor))
			continue
		}
		go func() {
			defer func() {
				//执行完就释放掉
				s.limiter.Release(1)
				// 这边要释放掉
				j.CancelFunc()
			}()
			// 开始执行任务
			err1 := exec.Exec(ctx, j)
			if err1 != nil {
				s.l.Error("执行任务失败",
					logger.Int64("jid", j.Id),
					logger.String("executor", j.Executor),
					logger.Error(err1))
				return
			}
			// job执行后 更新下一次执行的时间
			err1 = s.svc.ResetNextTime(ctx, j)
			if err1 != nil {
				s.l.Error("重置下次执行时间失败",
					logger.Int64("jid", j.Id),
					logger.Error(err1))
				return
			}
		}()
	}
}
