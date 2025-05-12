package repository

import (
	"context"
	"webook/internal/domain"
)

//go:generate mockgen -source=./article_reader.go -package=repomocks -destination=./mocks/article_reader.mock.go ArticleReaderRepository
type ArticleReaderRepository interface {
	// Save 有则更新，无则插入，也就是insert or update语义
	Save(ctx context.Context, art domain.Article) error
}
