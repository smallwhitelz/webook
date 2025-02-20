package grpc

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"webook/pkg/limiter"
)

type InterceptorBuilder struct {
	limiter limiter.Limiter
	key     string
}

// NewInterceptorBuilder key 1. limiter:interactive-service => 整个点赞应用被限流
func NewInterceptorBuilder(limiter limiter.Limiter, key string) *InterceptorBuilder {
	return &InterceptorBuilder{limiter: limiter, key: key}
}

// BuildServerUnaryInterceptorBiz 限流业务
func (b *InterceptorBuilder) BuildServerUnaryInterceptorBiz() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		if getReq, ok := req.(*GeyByIDRequest); ok {
			key := fmt.Sprintf("limiter:user:get_by_id:%d", getReq.Id)
			limited, err := b.limiter.Limit(ctx, key)
			if err != nil {
				// 保守做法
				return nil, status.Errorf(codes.ResourceExhausted, "限流")
			}
			if limited {
				return nil, status.Errorf(codes.ResourceExhausted, "限流")
			}
		}
		return handler(ctx, req)
	}
}
