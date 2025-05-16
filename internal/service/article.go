package service

import (
	"context"
	"errors"
	"time"
	"webook/internal/domain"
	"webook/internal/events/article"
	"webook/internal/repository"
	"webook/pkg/logger"
)

//go:generate mockgen -source=./article.go -package=svcmocks -destination=./mocks/article.mock.go ArticleService
type ArticleService interface {
	Save(ctx context.Context, art domain.Article) (int64, error)
	Publish(ctx context.Context, art domain.Article) (int64, error)
	Withdraw(ctx context.Context, uid int64, id int64) error
	GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error)
	GetById(ctx context.Context, id int64) (domain.Article, error)
	GetPubById(ctx context.Context, id, uid int64) (domain.Article, error)
	ListPub(ctx context.Context, start time.Time, offset, limit int) ([]domain.Article, error)
}

type articleService struct {
	repo     repository.ArticleRepository
	producer article.Producer

	// V1写法专用
	readerRepo repository.ArticleReaderRepository
	authorRepo repository.ArticleAuthorRepository
	l          logger.LoggerV1
}

func (a *articleService) Save(ctx context.Context, art domain.Article) (int64, error) {
	art.Status = domain.ArticleStatusUnpublished
	if art.Id > 0 {
		err := a.repo.Update(ctx, art)
		return art.Id, err
	} else {
		return a.repo.Create(ctx, art)
	}
}

func (a *articleService) Publish(ctx context.Context, art domain.Article) (int64, error) {
	art.Status = domain.ArticleStatusPublished
	return a.repo.Sync(ctx, art)
}

func (a *articleService) Withdraw(ctx context.Context, uid int64, id int64) error {
	return a.repo.SyncStatus(ctx, uid, id, domain.ArticleStatusPrivate)
}

func (a *articleService) ListPub(ctx context.Context, start time.Time, offset, limit int) ([]domain.Article, error) {
	return a.repo.ListPub(ctx, start, offset, limit)
}

func (a *articleService) GetPubById(ctx context.Context, id, uid int64) (domain.Article, error) {
	res, err := a.repo.GetPubById(ctx, id)
	go func() {
		if err == nil {
			// 在这里发一个消息
			er := a.producer.ProduceReadEvent(article.ReadEvent{
				Aid: id,
				Uid: uid,
			})
			if er != nil {
				a.l.Error("发送 ReadEvent 失败",
					logger.Int64("aid", id),
					logger.Int64("uid", uid),
					logger.Error(er))
			}
		}
	}()

	return res, err
}

func (a *articleService) GetById(ctx context.Context, id int64) (domain.Article, error) {
	return a.repo.GetById(ctx, id)
}

func (a *articleService) GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error) {
	return a.repo.GetByAuthor(ctx, uid, offset, limit)
}

func NewArticleService(repo repository.ArticleRepository, producer article.Producer) ArticleService {
	return &articleService{
		repo:     repo,
		producer: producer,
	}
}

// NewArticleServiceV1 在service层同步数据，
// 采用两个repo，这里相当于在service层分发 操作repo
func NewArticleServiceV1(readerRepo repository.ArticleReaderRepository,
	authorRepo repository.ArticleAuthorRepository, l logger.LoggerV1) *articleService {
	return &articleService{
		readerRepo: readerRepo,
		authorRepo: authorRepo,
		l:          l,
	}
}

// PublishV1 这里可以开启数据库事务吗？
// 不可以，对于service来说，你不知道repo是用什么实现的
// 就算是mysql，事务这个东西只对同一个库不同表有效，这里万一分库分表也不适用
// 所以service这一层不建议开事务
func (a *articleService) PublishV1(ctx context.Context, art domain.Article) (int64, error) {
	// 先操作制作库
	// 这里操作线上库
	var (
		id  = art.Id
		err error
	)
	if art.Id > 0 {
		err = a.authorRepo.Update(ctx, art)
	} else {
		id, err = a.authorRepo.Create(ctx, art)
	}
	if err != nil {
		return 0, err
	}
	art.Id = id

	// 一般不急着引入重试，可以上线后看看这里失败率
	for i := 0; i < 3; i++ {
		// 我可能线上库已经有数据了
		// 也可能没有
		err = a.readerRepo.Save(ctx, art)
		if err != nil {
			// 多接入一些 tracing 的工具
			a.l.Error("保存到制作库成功，但是到线上库失败",
				logger.Int64("aid", art.Id),
				logger.Error(err))
		} else {
			return id, err
		}

	}
	a.l.Error("保存到制作库成功，但是到线上库失败，重试耗尽",
		logger.Int64("aid", art.Id),
		logger.Error(err))
	return id, errors.New("保存到线上库失败，重试次数耗尽")
}
