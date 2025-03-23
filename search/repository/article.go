package repository

import (
	"context"
	"github.com/ecodeclub/ekit/slice"
	"webook/search/domain"
	"webook/search/repository/dao"
)

type articleRepository struct {
	dao  dao.ArticleDAO
	tags dao.TagDAO
}

func (a *articleRepository) InputArticle(ctx context.Context, article domain.Article) error {
	return a.dao.InputArticle(ctx, dao.Article{
		Id:      article.Id,
		Title:   article.Title,
		Status:  article.Status,
		Content: article.Content,
		Tags:    article.Tags,
	})
}

func (a *articleRepository) SearchArticle(ctx context.Context, uid int64, keywords []string) ([]domain.Article, error) {
	artIDs, err := a.tags.Search(ctx, uid, "article", keywords)
	if err != nil {
		return nil, err
	}
	arts, err := a.dao.Search(ctx, artIDs, keywords)
	if err != nil {
		return nil, err
	}
	return slice.Map(arts, func(idx int, src dao.Article) domain.Article {
		return domain.Article{
			Id:      src.Id,
			Title:   src.Title,
			Status:  src.Status,
			Content: src.Content,
			Tags:    src.Tags,
		}
	}), nil
}

func NewArticleRepository(dao dao.ArticleDAO) ArticleRepository {
	return &articleRepository{dao: dao}
}
