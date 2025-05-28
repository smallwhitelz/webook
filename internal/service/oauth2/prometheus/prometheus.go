package prometheus

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"time"
	"webook/internal/domain"
	"webook/internal/service/oauth2/wechat"
)

// Decorator 通过装饰器监控微信扫码验证响应时间
type Decorator struct {
	wechat.Service
	sum prometheus.Summary
}

func NewDecorator(svc wechat.Service, sum prometheus.Summary) *Decorator {
	return &Decorator{
		Service: svc,
		sum:     sum,
	}
}

func (d *Decorator) VerifyCode(ctx context.Context, code string) (domain.WechatInfo, error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start).Milliseconds()
		d.sum.Observe(float64(duration))
	}()
	return d.Service.VerifyCode(ctx, code)
}
