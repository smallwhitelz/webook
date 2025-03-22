package service

import (
	"context"
	"golang.org/x/sync/errgroup"
	"strings"
	"webook/search/domain"
	"webook/search/repository"
)

type SearchService interface {
	Search(ctx context.Context, uid int64, expression string) (domain.SearchResult, error)
}

type searchService struct {
	userRepo    repository.UserRepository
	articleRepo repository.ArticleRepository
}

func (s *searchService) Search(ctx context.Context, uid int64, expression string) (domain.SearchResult, error) {
	// 你要搜索用户，也要搜索文章
	// 要对expression进行处理
	// 输入预处理
	keywords := strings.Split(expression, " ")
	var eg errgroup.Group
	var res domain.SearchResult
	eg.Go(func() error {
		users, err := s.userRepo.SearchUser(ctx, keywords)
		res.Users = users
		return err
	})
	eg.Go(func() error {
		arts, err := s.articleRepo.SearchArticle(ctx, uid, keywords)
		res.Articles = arts
		return err
	})
	return res, eg.Wait()
}

func NewSearchService(userRepo repository.UserRepository, articleRepo repository.ArticleRepository) SearchService {
	return &searchService{userRepo: userRepo, articleRepo: articleRepo}
}
