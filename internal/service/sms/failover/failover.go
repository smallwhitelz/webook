package failover

import (
	"context"
	"errors"
	"log"
	"sync/atomic"
	"webook/internal/service/sms"
)

// FailOverSMSService 依赖第三方服务商，当某一个服务商出现问题了，可以进行故障切换，保证我们的短信服务正常
type FailOverSMSService struct {
	svcs []sms.Service

	// v1的字段
	// 当前服务商下标
	idx uint64
}

func NewFailOverSMSService(svcs []sms.Service) *FailOverSMSService {
	return &FailOverSMSService{
		svcs: svcs,
	}
}

// Send 每次轮训都会从第一个服务商开始，大量的请求都是第一个，负载不均衡
// 如果 svcs有十几个，那么轮训会很慢
func (f *FailOverSMSService) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	for _, svc := range f.svcs {
		err := svc.Send(ctx, tplId, args, numbers...)
		if err == nil {
			return nil
		}
		log.Println(err)
	}
	return errors.New("轮询了所有服务商，但是发送都失败了")
}

// SendV1 起始下标轮询
// 并且出错也轮询
func (f *FailOverSMSService) SendV1(ctx context.Context, tplId string, args []string, numbers ...string) error {
	idx := atomic.AddUint64(&f.idx, 1)
	length := uint64(len(f.svcs))
	for i := idx; i < idx+length; i++ {
		// 取余数计算下标
		svc := f.svcs[i%length]
		err := svc.Send(ctx, tplId, args, numbers...)
		switch err {
		case nil:
			return nil
		case context.Canceled, context.DeadlineExceeded:
			// 前者是被取消，后者是超时
			return err
		}
		log.Println(err)
	}
	return errors.New("轮询了所有服务商，但是发送都失败了")
}
