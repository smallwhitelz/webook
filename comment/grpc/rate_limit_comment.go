package grpc

import (
	"context"
	commentv1 "webook/api/proto/gen/comment/v1"
)

type RateLimitComment struct {
	CommentServiceServer
}

func (c *RateLimitComment) CreateComment(ctx context.Context, request *commentv1.CreateCommentRequest) (*commentv1.CreateCommentResponse, error) {
	if ctx.Value("limited") == "true" || ctx.Value("downgrade") == "true" {
		// 转Kafka
	}
	err := c.svc.CreateComment(ctx, convertToDomain(request.GetComment()))
	return &commentv1.CreateCommentResponse{}, err
}
