package grpc

import (
	"context"
	"webook/api/proto/gen/comment/v1"
	"webook/comment/domain"
	"webook/comment/service"
)

type CommentServiceServer struct {
	commentv1.UnimplementedCommentServiceServer
	svc service.CommentService
}

func (c *CommentServiceServer) GetCommentList(ctx context.Context, request *commentv1.CommentListRequest) (*commentv1.CommentListResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (c *CommentServiceServer) DeleteComment(ctx context.Context, request *commentv1.DeleteCommentRequest) (*commentv1.DeleteCommentResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (c *CommentServiceServer) CreateComment(ctx context.Context, request *commentv1.CreateCommentRequest) (*commentv1.CreateCommentResponse, error) {
	err := c.svc.CreateComment(ctx, convertToDomain(request.GetComment()))
	return &commentv1.CreateCommentResponse{}, err
}

func (c *CommentServiceServer) GetMoreReplies(ctx context.Context, request *commentv1.GetMoreRepliesRequest) (*commentv1.GetMoreRepliesResponse, error) {
	//TODO implement me
	panic("implement me")
}

func convertToDomain(comment *commentv1.Comment) domain.Comment {
	domainComment := domain.Comment{
		Id:      comment.Id,
		Biz:     comment.Biz,
		BizID:   comment.Bizid,
		Content: comment.Content,
		Commentator: domain.User{
			ID: comment.Uid,
		},
	}
	if comment.GetParentComment() != nil {
		domainComment.ParentComment = &domain.Comment{
			Id: comment.GetParentComment().GetId(),
		}
	}
	if comment.GetRootComment() != nil {
		domainComment.RootComment = &domain.Comment{
			Id: comment.GetRootComment().GetId(),
		}
	}
	return domainComment
}
