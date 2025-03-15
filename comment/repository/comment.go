package repository

import (
	"context"
	"database/sql"
	"golang.org/x/sync/errgroup"
	"time"
	"webook/comment/domain"
	"webook/comment/repository/dao"
	"webook/pkg/logger"
)

type CommentRepository interface {
	CreateComment(ctx context.Context, comment domain.Comment) error
	// DeleteComment 删除评论，删除本评论及其子评论
	DeleteComment(ctx context.Context, comment domain.Comment) error
	FindByBiz(ctx context.Context, biz string, bizId int64, minId int64, limit int64) ([]domain.Comment, error)
	GetMoreReplies(ctx context.Context, rid int64, maxID int64, limit int64) ([]domain.Comment, error)
}

type CachedCommentRepo struct {
	dao dao.CommentDAO
	l   logger.V1
}

func (c *CachedCommentRepo) GetMoreReplies(ctx context.Context, rid int64, maxID int64, limit int64) ([]domain.Comment, error) {
	cs, err := c.dao.FindRepliesByRid(ctx, rid, maxID, limit)
	if err != nil {
		return nil, err
	}
	res := make([]domain.Comment, 0, len(cs))
	for _, cm := range cs {
		res = append(res, c.toDomain(cm))
	}
	return res, nil
}

func (c *CachedCommentRepo) FindByBiz(ctx context.Context, biz string, bizId int64, minId int64, limit int64) ([]domain.Comment, error) {
	daoComments, err := c.dao.FindByBiz(ctx, biz, bizId, minId, limit)
	if err != nil {
		return nil, err
	}
	res := make([]domain.Comment, 0, len(daoComments))
	// 这时候要去找子评论了，找三条
	var eg errgroup.Group
	// 触发降级
	downgrade := ctx.Value("downgrade") == "true"
	for _, dc := range daoComments {
		// 保证在开另一个协程的时候他操作的都是同一个dc
		newDC := dc
		cm := c.toDomain(dc)
		res = append(res, cm)
		// 什么都不去做，不去找他的子评论
		if downgrade {
			continue
		}
		eg.Go(func() error {
			subComments, err := c.dao.FindRepliesByPid(ctx, newDC.Id, 0, 3)
			if err != nil {
				return err
			}
			cm.Children = make([]domain.Comment, 0, len(subComments))
			for _, sc := range subComments {
				cm.Children = append(cm.Children, c.toDomain(sc))
			}
			return nil
		})
	}
	return res, eg.Wait()
}

func (c *CachedCommentRepo) CreateComment(ctx context.Context, comment domain.Comment) error {
	return c.dao.Insert(ctx, c.toEntity(comment))
}

func (c *CachedCommentRepo) DeleteComment(ctx context.Context, comment domain.Comment) error {
	return c.dao.Delete(ctx, dao.Comment{
		Id: comment.Id,
	})
}

func NewCachedCommentRepo(dao dao.CommentDAO, l logger.V1) CommentRepository {
	return &CachedCommentRepo{dao: dao, l: l}
}

func (c *CachedCommentRepo) toEntity(domainComment domain.Comment) dao.Comment {
	daoComment := dao.Comment{
		Id:      domainComment.Id,
		Uid:     domainComment.Commentator.ID,
		Biz:     domainComment.Biz,
		BizID:   domainComment.BizID,
		Content: domainComment.Content,
	}
	if domainComment.RootComment != nil {
		daoComment.RootID = sql.NullInt64{
			Valid: true,
			Int64: domainComment.RootComment.Id,
		}
	}
	if domainComment.ParentComment != nil {
		daoComment.PID = sql.NullInt64{
			Valid: true,
			Int64: domainComment.ParentComment.Id,
		}
	}
	daoComment.Ctime = time.Now().UnixMilli()
	daoComment.Utime = time.Now().UnixMilli()
	return daoComment
}

func (c *CachedCommentRepo) toDomain(daoComment dao.Comment) domain.Comment {
	val := domain.Comment{
		Id: daoComment.Id,
		Commentator: domain.User{
			ID: daoComment.Uid,
		},
		Biz:     daoComment.Biz,
		BizID:   daoComment.BizID,
		Content: daoComment.Content,
		CTime:   time.UnixMilli(daoComment.Ctime),
		UTime:   time.UnixMilli(daoComment.Utime),
	}
	if daoComment.PID.Valid {
		val.ParentComment = &domain.Comment{
			Id: daoComment.PID.Int64,
		}
	}
	if daoComment.RootID.Valid {
		val.RootComment = &domain.Comment{
			Id: daoComment.RootID.Int64,
		}
	}
	return val
}
