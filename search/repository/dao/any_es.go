package dao

import (
	"context"
	"github.com/olivere/elastic/v7"
)

type AnyESDAO struct {
	client *elastic.Client
}

func (a *AnyESDAO) Input(ctx context.Context, idxName string, docID string, data string) error {
	_, err := a.client.Index().Index(idxName).Id(docID).BodyString(data).Do(ctx)
	return err
}

func NewAnyESDAO(client *elastic.Client) AnyDAO {
	return &AnyESDAO{client: client}

}
