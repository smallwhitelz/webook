package repository

import (
	"context"
	"webook/search/repository/dao"
)

type anyRepository struct {
	dao dao.AnyDAO
}

func (a *anyRepository) Input(ctx context.Context, idxName string, docID string, data string) error {
	return a.dao.Input(ctx, idxName, docID, data)
}

func NewAnyRepository(dao dao.AnyDAO) AnyRepository {
	return &anyRepository{dao: dao}
}
