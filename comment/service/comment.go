package service

import (
	"context"
	"webook/comment/domain"
	"webook/comment/repository"
)

type CommentService interface {

	// CreateComment 创建评论
	CreateComment(ctx context.Context, comment domain.Comment) error

	// DeleteComment 删除评论，删除本评论何其子评论
	DeleteComment(ctx context.Context, id int64) error

	GetCommentList(ctx context.Context, biz string, bizId int64, minId int64, limit int64) ([]domain.Comment, error)
}

type commentService struct {
	repo repository.CommentRepository
}

func (c *commentService) GetCommentList(ctx context.Context, biz string, bizId int64, minId int64, limit int64) ([]domain.Comment, error) {
	list, err := c.repo.FindByBiz(ctx, biz, bizId, minId, limit)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (c *commentService) CreateComment(ctx context.Context, comment domain.Comment) error {
	return c.repo.CreateComment(ctx, comment)
}

func (c *commentService) DeleteComment(ctx context.Context, id int64) error {
	return c.repo.DeleteComment(ctx, domain.Comment{
		Id: id,
	})
}
