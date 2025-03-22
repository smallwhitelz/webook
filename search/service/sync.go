package service

import (
	"context"
	"webook/search/domain"
	"webook/search/repository"
)

type SyncService interface {
	InputArticle(ctx context.Context, article domain.Article) error
	InputUser(ctx context.Context, user domain.User) error
	InputAny(ctx context.Context, idxName, docID, data string) error
}

type syncService struct {
	userRepo    repository.UserRepository
	articleRepo repository.ArticleRepository
}

func (s *syncService) InputArticle(ctx context.Context, article domain.Article) error {
	return s.articleRepo.InputArticle(ctx, article)
}

func (s *syncService) InputUser(ctx context.Context, user domain.User) error {
	return s.userRepo.InputUser(ctx, user)
}

func (s *syncService) InputAny(ctx context.Context, idxName, docID, data string) error {
	//TODO implement me
	panic("implement me")
}

func NewSyncService(userRepo repository.UserRepository, articleRepo repository.ArticleRepository) SyncService {
	return &syncService{userRepo: userRepo, articleRepo: articleRepo}
}
