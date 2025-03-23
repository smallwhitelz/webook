package dao

import (
	"context"
	"encoding/json"
	"github.com/ecodeclub/ekit/slice"
	"github.com/olivere/elastic/v7"
	"strconv"
	"strings"
)

const ArticleIndexName = "article_index"

type Article struct {
	Id      int64    `json:"id"`
	Title   string   `json:"title"`
	Status  int32    `json:"status"`
	Content string   `json:"content"`
	Tags    []string `json:"tags"`
}

type ArticleElasticDAO struct {
	client *elastic.Client
}

func (a *ArticleElasticDAO) InputArticle(ctx context.Context, article Article) error {
	_, err := a.client.Index().
		Index(ArticleIndexName).
		Id(strconv.FormatInt(article.Id, 10)).
		BodyJson(article).Do(ctx)
	return err
}

func (a *ArticleElasticDAO) Search(ctx context.Context, artIds []int64, keywords []string) ([]Article, error) {
	queryString := strings.Join(keywords, " ")
	// 2=>用户可见的文章
	status := elastic.NewTermsQuery("status", 2)
	title := elastic.NewMatchQuery("title", queryString)
	content := elastic.NewMatchQuery("content", queryString)
	tag := elastic.NewTermsQuery("id", slice.Map(artIds, func(idx int, src int64) any {
		return src
	})).Boost(2)
	or := elastic.NewBoolQuery().Should(title, content, tag)
	query := elastic.NewBoolQuery().Must(status, or)
	resp, err := a.client.Search(ArticleIndexName).Query(query).Do(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]Article, 0, len(resp.Hits.Hits))
	for _, hit := range resp.Hits.Hits {
		var art Article
		err := json.Unmarshal(hit.Source, &art)
		if err != nil {
			return nil, err
		}
		res = append(res, art)
	}
	return res, nil
}

func NewArticleElasticDAO(client *elastic.Client) ArticleDAO {
	return &ArticleElasticDAO{client: client}
}
