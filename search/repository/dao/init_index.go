package dao

import (
	"context"
	_ "embed"
	"github.com/olivere/elastic/v7"
	"golang.org/x/sync/errgroup"
	"time"
)

var (
	//go:embed user_index.json
	userIndex string
	//go:embed article_index.json
	articleIndex string
	//go:embed tags_index.json
	tagIndex string
)

func InitES(client *elastic.Client) error {
	const timeout = time.Second * 10
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	var eg errgroup.Group
	eg.Go(func() error {
		return tryCreateIndex(ctx, client, UserIndexName, userIndex)
	})
	eg.Go(func() error {
		return tryCreateIndex(ctx, client, ArticleIndexName, articleIndex)
	})
	eg.Go(func() error {
		return tryCreateIndex(ctx, client, TagIndexName, tagIndex)
	})

	return eg.Wait()
}

func tryCreateIndex(ctx context.Context,
	client *elastic.Client,
	idxName, idxCfg string,
) error {
	// 索引可能已经建好了
	ok, err := client.IndexExists(idxName).Do(ctx)
	if err != nil {
		return err
	}
	if ok {
		return nil
	}
	_, err = client.CreateIndex(idxName).Body(idxCfg).Do(ctx)
	return err
}
