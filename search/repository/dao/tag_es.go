package dao

import (
	"context"
	"encoding/json"
	"github.com/olivere/elastic/v7"
)

const TagIndexName = "tags_index"

type TagESDAO struct {
	client *elastic.Client
}

func (t *TagESDAO) Search(ctx context.Context, uid int64, biz string, keywords []string) ([]int64, error) {
	query := elastic.NewBoolQuery().Must(
		// 必须是我打的标签
		elastic.NewTermsQuery("uid", uid),
		elastic.NewTermsQuery("biz", biz),
		elastic.NewTermsQueryFromStrings("tags", keywords...),
	)
	resp, err := t.client.Search(TagIndexName).Query(query).Do(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]int64, 0, len(resp.Hits.Hits))
	for _, hit := range resp.Hits.Hits {
		var bt BizTags
		err = json.Unmarshal(hit.Source, &bt)
		if err != nil {
			return nil, err
		}
		res = append(res, bt.BizId)
	}
	return res, nil
}

func NewTagESDAO(client *elastic.Client) TagDAO {
	return &TagESDAO{client: client}

}

type BizTags struct {
	Uid   int64    `json:"uid"`
	Biz   string   `json:"biz"`
	BizId int64    `json:"biz_id"`
	Tags  []string `json:"tags"`
}
