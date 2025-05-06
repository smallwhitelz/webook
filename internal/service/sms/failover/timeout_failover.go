package failover

import (
	"context"
	"sync/atomic"
	"webook/internal/service/sms"
)

// TimeoutFailoverSMSService 当某个服务商短信服务连续超时达到某个阈值
// 进行服务商切换
type TimeoutFailoverSMSService struct {
	svcs []sms.Service
	// 当前正在使用节点
	idx int32
	// 连续几个超时
	cnt int32
	// 切换的阈值，只读的
	threshold int32
}

func NewTimeoutFailoverSMSService(svcs []sms.Service, threshold int32) *TimeoutFailoverSMSService {
	return &TimeoutFailoverSMSService{
		svcs:      svcs,
		threshold: threshold,
	}
}

// Send 这里的原子操作已经保证了数据竞争的安全
// 但是多步操作依然会有并发安全问题，但是这个我们是可以接受的
// 如果想要彻底解决并发安全，就要加锁，但是加锁了在高并发的时候效率会大幅度下降
func (t *TimeoutFailoverSMSService) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	idx := atomic.LoadInt32(&t.idx)
	cnt := atomic.LoadInt32(&t.cnt)
	// 超过阈值，执行切换
	if cnt >= t.threshold {
		newIdx := (idx + 1) % int32(len(t.svcs))
		if atomic.CompareAndSwapInt32(&t.idx, idx, newIdx) {
			// 重制这个cnt计数
			atomic.StoreInt32(&t.cnt, 0)
		}
		idx = newIdx
	}
	svc := t.svcs[idx]
	err := svc.Send(ctx, tplId, args, numbers...)
	switch err {
	case nil:
		// 连续超时，所以不超时的时候重制为0
		atomic.StoreInt32(&t.cnt, 0)
		return nil
	case context.DeadlineExceeded:
		atomic.AddInt32(&t.cnt, 1)
	default:
		// 遇到了错误，但是不是超时错误，你要考虑怎么搞
		// 我可以增加cnt，也可以不增加
		// 如果强调一定是超时，可以不增加
		//如果是EOF之类的错误，你还可以考虑直接切换
	}
	return err
}
