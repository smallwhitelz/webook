package grpc

import (
	"context"
	"errors"
	commentv1 "webook/api/proto/gen/comment/v1"
)

type RateLimitComment struct {
	CommentServiceServer
}

// GetCommentList 针对热门资源限流和非热门资源限流的一种方式
func (c *RateLimitComment) GetCommentList(ctx context.Context, request *commentv1.CommentListRequest) (*commentv1.CommentListResponse, error) {
	// 一般是通过热榜功能，提前计算放到redis里面，问一下redis就知道是不是热门资源
	isHotBiz := c.isHost(request.Biz, request.Bizid)
	if isHotBiz {
		// 限流阈值400/s
	} else {
		// 限流阈值100/s
	}
	return c.CommentServiceServer.GetCommentList(ctx, request)
}

func (c *RateLimitComment) GetCommentListV1(ctx context.Context, request *commentv1.CommentListRequest) (*commentv1.CommentListResponse, error) {
	// 一般是通过热榜功能，提前计算放到redis里面，问一下redis就知道是不是热门资源
	isHotBiz := c.isHost(request.Biz, request.Bizid)
	if !isHotBiz && ctx.Value("downgrade") == "true" {
		// 非热门资源触发降级
		return &commentv1.CommentListResponse{}, errors.New("非热门资源被降级")
	}
	return c.CommentServiceServer.GetCommentList(ctx, request)
}

func (c *RateLimitComment) CreateComment(ctx context.Context, request *commentv1.CreateCommentRequest) (*commentv1.CreateCommentResponse, error) {
	if ctx.Value("limited") == "true" || ctx.Value("downgrade") == "true" {
		// 转Kafka
	}
	err := c.svc.CreateComment(ctx, convertToDomain(request.GetComment()))
	return &commentv1.CreateCommentResponse{}, err
}

func (c *RateLimitComment) isHost(biz string, bizid int64) bool {
	return true
}
