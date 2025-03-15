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

	// GetCommentList Comment的id为0 获取一级评论
	// 按照 ID 倒序排序
	GetCommentList(ctx context.Context, biz string, bizId int64, minId int64, limit int64) ([]domain.Comment, error)
	GetMoreReplies(ctx context.Context, rid int64, maxID int64, limit int64) ([]domain.Comment, error)
}

type commentService struct {
	repo repository.CommentRepository
}

func (c *commentService) GetMoreReplies(ctx context.Context, rid int64, maxID int64, limit int64) ([]domain.Comment, error) {
	return c.repo.GetMoreReplies(ctx, rid, maxID, limit)
}

func NewCommentService(repo repository.CommentRepository) CommentService {
	return &commentService{repo: repo}
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
