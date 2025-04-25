package article

import (
	"context"
	"github.com/IBM/sarama"
	"time"
	"webook/internal/domain"
	"webook/internal/repository"
	"webook/internal/repository/dao"
	"webook/pkg/canalx"
	"webook/pkg/logger"
	"webook/pkg/saramax"
)

type MySQLBinlogConsumer struct {
	client sarama.Client
	l      logger.V1
	repo   *repository.CachedArticleRepository
}

func (r *MySQLBinlogConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("pub_articles_cache",
		r.client)
	if err != nil {
		return err
	}
	go func() {
		err := cg.Consume(context.Background(),
			[]string{"webook_binlog"},
			saramax.NewHandler[canalx.Message[dao.PublishedArticle]](r.l, r.Consume))
		if err != nil {
			r.l.Error("退出了消费循环异常", logger.Error(err))
		}
	}()
	return err
}

func (r *MySQLBinlogConsumer) Consume(msg *sarama.ConsumerMessage, val canalx.Message[dao.PublishedArticle]) error {
	if val.Table != "published_articles" {
		// 我不关心的
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	for _, data := range val.Data {
		var err error
		switch data.Status {
		case domain.ArticleStatusPublished.ToUint8():
			// 发表
			err = r.repo.Cache().SetPub(ctx, r.repo.ToDomain(dao.Article(data)))
		case domain.ArticleStatusPrivate.ToUint8():
			err = r.repo.Cache().DelPub(ctx, data.Id)
		}
		if err != nil {
			// 可以继续，也可以中断
			return err
		}
	}
	return nil
}
