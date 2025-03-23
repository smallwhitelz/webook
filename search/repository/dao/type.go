package dao

import "context"

type UserDAO interface {
	InputUser(ctx context.Context, user User) error
	Search(ctx context.Context, keywords []string) ([]User, error)
}

type ArticleDAO interface {
	InputArticle(ctx context.Context, article Article) error
	// Search  artIds命中了索引的 article id
	Search(ctx context.Context, artIds []int64, keywords []string) ([]Article, error)
}

type AnyDAO interface {
	Input(ctx context.Context, idxName string, docID string, data string) error
}
type TagDAO interface {
	Search(ctx context.Context, uid int64, biz string, keywords []string) ([]int64, error)
}
