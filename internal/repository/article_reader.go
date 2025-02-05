package repository

import (
	"context"
	"webook/internal/domain"
)

type ArticleReaderRepository interface {
	// Save 有则更新，无则插入，也就是insert or update语义
	Save(ctx context.Context, art domain.Article) error
}
