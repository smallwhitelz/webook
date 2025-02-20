package grpc

import (
	"context"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"webook/pkg/limiter"
)

type LimiterUserServer struct {
	limiter limiter.Limiter
	UserServiceServer
}

// GetByID 利用装饰器限流单个业务接口
func (s *LimiterUserServer) GetByID(ctx context.Context, req *GeyByIDRequest) (*GeyByIDResponse, error) {
	key := fmt.Sprintf("limiter:user:get_by_id:%d", req.Id)
	limited, err := s.limiter.Limit(ctx, key)
	if err != nil {
		// 保守做法
		return nil, status.Errorf(codes.ResourceExhausted, "限流")
	}
	if limited {
		return nil, status.Errorf(codes.ResourceExhausted, "限流")
	}
	return s.UserServiceServer.GetByID(ctx, req)
}
