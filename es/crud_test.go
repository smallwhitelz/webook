package es

import (
	"context"
	"encoding/json"
	elastic "github.com/elastic/go-elasticsearch/v8"
	olivere "github.com/olivere/elastic/v7"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"strings"
	"testing"
	"time"
)

type ElasticSearchTestSuite struct {
	suite.Suite
	es      *elastic.Client
	olivere *olivere.Client
}

func (s *ElasticSearchTestSuite) SetupSuite() {
	es, err := elastic.NewClient(elastic.Config{
		Addresses: []string{"http://43.154.97.245:9200"},
	})
	require.NoError(s.T(), err)
	s.es = es
	ol, err := olivere.NewClient(olivere.SetURL("http://43.154.97.245:9200"))
	require.NoError(s.T(), err)
	s.olivere = ol
}

func (s *ElasticSearchTestSuite) TestCreateIndex() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	def := `
{
    "settings": {
        "number_of_shards": 3,
        "number_of_replicas": 3
    },
    "mappings": {
        "properties": {
            "email": {
                "type": "text"
            },
            "phone": {
                "type": "keyword"
            },
            "birthday": {
                "type": "date"
            }
        }
    }
}
`

	resp, err := s.es.Indices.Create("user_idx_go",
		s.es.Indices.Create.WithContext(ctx),
		s.es.Indices.Create.WithBody(strings.NewReader(def)))
	require.NoError(s.T(), err)
	require.Equal(s.T(), 200, resp.StatusCode)

	_, err = s.olivere.CreateIndex("user_idx_go_ol").Body(def).Do(ctx)
	require.NoError(s.T(), err)
}

func (s *ElasticSearchTestSuite) TestPutDoc() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	data := `
{  
  "email": "john@example.com",  
  "phone": "1234567890",  
  "birthday": "2000-01-01"  
}
`
	_, err := s.es.Index("user_idx_go", strings.NewReader(data), s.es.Index.WithContext(ctx))
	require.NoError(s.T(), err)

	_, err = s.olivere.Index().Index("user_idx_go").
		BodyJson(User{Email: "john@example.com"}).Do(ctx)
	require.NoError(s.T(), err)
}

func (s *ElasticSearchTestSuite) TestGetDoc() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	//s.es.Get()
	query := `
{  
  "query": {  
    "range": {  
      "birthday": {
        "gte": "1990-01-01"
      }
    }  
  }  
}
`
	_, err := s.es.Search(s.es.Search.WithContext(ctx),
		s.es.Search.WithIndex("user_idx_go"),
		s.es.Search.WithBody(strings.NewReader(query)))
	require.NoError(s.T(), err)

	olQuery := olivere.NewMatchQuery("email", "john")
	resp, err := s.olivere.Search("user_idx_go").Query(olQuery).Do(ctx)
	require.NoError(s.T(), err)
	for _, hit := range resp.Hits.Hits {
		var u User
		err = json.Unmarshal(hit.Source, &u)
		require.NoError(s.T(), err)
		s.T().Log(u)
	}
}

func TestElasticSearch(t *testing.T) {
	suite.Run(t, new(ElasticSearchTestSuite))
}

type User struct {
	Email string `json:"email"`
	Phone string `json:"phone"`
}
