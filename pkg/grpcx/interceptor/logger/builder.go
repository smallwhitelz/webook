package logger

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"runtime"
	"time"
	"webook/pkg/grpcx/interceptor"
	"webook/pkg/logger"
)

type InterceptorBuilder struct {
	l logger.V1
	interceptor.Builder
}

func (b *InterceptorBuilder) BuildServerUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		start := time.Now()
		event := "normal"
		defer func() {
			// 最终输出日志
			cost := time.Since(start)

			// 发生了panic
			if rec := recover(); rec != nil {
				// 类型断言
				switch re := rec.(type) {
				case error:
					err = re
				default:
					err = fmt.Errorf("%v", rec)
				}
				event = "recover"
				stack := make([]byte, 4096)
				stack = stack[:runtime.Stack(stack, true)]
				err = status.New(codes.Internal, "panic, err "+err.Error()).Err()
			}
			fields := []logger.Field{
				// unary stream 是 grpc的两种调用形态
				// unary 理解为一次调用
				// stream是连绵不绝的调用
				logger.String("type", "unary"),
				logger.Int64("cost", cost.Milliseconds()),
				logger.String("event", event),
				logger.String("method", info.FullMethod),
				// 客户端信息
				logger.String("peer", b.PeerName(ctx)),
				logger.String("peer_id", b.PeerIP(ctx)),
			}
			st, _ := status.FromError(err)
			if st != nil {
				// 错误码
				fields = append(fields, logger.String("code", st.Code().String()))
				fields = append(fields, logger.String("code_msg", st.Message()))
			}

			b.l.Info("RPC调用", fields...)
		}()
		resp, err = handler(ctx, req)
		return
	}
}
