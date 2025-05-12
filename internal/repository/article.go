package repository

import (
	"context"
	"github.com/ecodeclub/ekit/slice"
	"gorm.io/gorm"
	"time"
	"webook/internal/domain"
	"webook/internal/repository/cache"
	"webook/internal/repository/dao"
	"webook/pkg/logger"
)

type ArticleRepository interface {
	Create(ctx context.Context, art domain.Article) (int64, error)
	Update(ctx context.Context, art domain.Article) error
	Sync(ctx context.Context, art domain.Article) (int64, error)
	SyncStatus(ctx context.Context, uid int64, id int64, status domain.ArticleStatus) error
	GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error)
	GetById(ctx context.Context, id int64) (domain.Article, error)
	GetPubById(ctx context.Context, id int64) (domain.Article, error)
	ListPub(ctx context.Context, start time.Time, offset int, limit int) ([]domain.Article, error)
}

type CachedArticleRepository struct {
	dao   dao.ArticleDao
	cache cache.ArticleCache
	// 如果你直接访问UserDAO，你就绕开了repository
	// repository 一般都有缓存机制
	userRepo  UserRepository
	readerDAO dao.ArticleReaderDAO
	authorDAO dao.ArticleAuthorDAO

	db *gorm.DB
	l  logger.LoggerV1
}

func (c *CachedArticleRepository) ListPub(ctx context.Context, start time.Time, offset int, limit int) ([]domain.Article, error) {
	arts, err := c.dao.ListPub(ctx, start, offset, limit)
	if err != nil {
		return nil, err
	}
	return slice.Map[dao.PublishedArticle, domain.Article](arts, func(idx int, src dao.PublishedArticle) domain.Article {
		return c.toDomain(dao.Article(src))
	}), nil
}

func (c *CachedArticleRepository) GetPubById(ctx context.Context, id int64) (domain.Article, error) {
	res, err := c.cache.GetPub(ctx, id)
	if err == nil {
		return res, err
	}
	art, err := c.dao.GetPubById(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}
	// 我现在要去查询 User 信息，拿到创作者信息
	res = c.toDomain(dao.Article(art))
	author, err := c.userRepo.FindById(ctx, art.AuthorId)
	if err != nil {
		return domain.Article{}, err
		// 这种要额外记录日志，因为吞掉了错误信息
		//return res,err
	}
	res.Author.Name = author.Nickname
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		er := c.cache.SetPub(ctx, res)
		if er != nil {
			// 记录日志
		}
	}()
	return res, nil
}

func (c *CachedArticleRepository) GetById(ctx context.Context, id int64) (domain.Article, error) {
	res, err := c.cache.Get(ctx, id)
	if err == nil {
		return res, nil
	}
	art, err := c.dao.GetById(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}
	res = c.toDomain(art)
	go func() {
		er := c.cache.Set(ctx, res)
		if er != nil {
			// 记录日志
		}
	}()
	return res, nil
}

func (c *CachedArticleRepository) GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error) {
	// 首先第一步，判定要不要查缓存
	// 事实上，limit <= 100 都可以查询缓存
	if offset == 0 && limit == 100 {
		res, err := c.cache.GetFirstPage(ctx, uid)
		if err == nil {
			return res, err
		} else {
			// 要考虑记录日志
			// 缓存未命中，你是可以忽略的
		}
	}
	arts, err := c.dao.GetByAuthor(ctx, uid, offset, limit)
	if err != nil {
		return nil, err
	}
	res := slice.Map[dao.Article, domain.Article](arts, func(idx int, src dao.Article) domain.Article {
		return c.toDomain(src)
	})
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if offset == 0 && limit == 100 {
			// 缓存回写失败，不一定是大问题，但有可能是大问题
			err = c.cache.SetFirstPage(ctx, uid, res)
			if err != nil {
				// 记录日志
				// 我需要监控这里
			}
		}
	}()
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		c.preCache(ctx, res)
	}()
	return res, nil
}

func (c *CachedArticleRepository) SyncStatus(ctx context.Context, uid int64, id int64, status domain.ArticleStatus) error {
	err := c.dao.SyncStatus(ctx, uid, id, status.ToUint8())
	if err == nil {
		er := c.cache.DelFirstPage(ctx, uid)
		if er != nil {
			// 也要记录日志
		}
	}
	return err
}

func (c *CachedArticleRepository) Sync(ctx context.Context, art domain.Article) (int64, error) {
	id, err := c.dao.Sync(ctx, c.toEntity(art))
	if err == nil {
		er := c.cache.DelFirstPage(ctx, art.Author.Id)
		if er != nil {
			// 也要记录日志
		}
	}
	// 在这里尝试设置缓存
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		user, er := c.userRepo.FindById(ctx, art.Author.Id)

		if er != nil {
			// 记录日志
			return
		}
		art.Author = domain.Author{
			Id:   user.Id,
			Name: user.Nickname,
		}
		// 你可以灵活设置过期时间
		er = c.cache.SetPub(ctx, art)
		if er != nil {
			// 记录日志
		}
	}()
	return id, err
}

// SyncV1 非事务实现，不同数据库
func (c *CachedArticleRepository) SyncV1(ctx context.Context, art domain.Article) (int64, error) {
	artn := c.toEntity(art)
	var (
		id  = art.Id
		err error
	)
	if id > 0 {
		err = c.authorDAO.Update(ctx, artn)
	} else {
		id, err = c.authorDAO.Create(ctx, artn)
	}
	if err != nil {
		return 0, err
	}
	artn.Id = id
	err = c.readerDAO.Upsert(ctx, artn)
	return id, err
}

// SyncV2 开启事务
// 缺陷：强制Repository引入db，相当于Repository强制依赖了DAO依赖的东西
// 没有坚持面向接口编程原则
// 跨层依赖
func (c *CachedArticleRepository) SyncV2(ctx context.Context, art domain.Article) (int64, error) {
	tx := c.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return 0, tx.Error
	}
	// 防止后面业务panic
	defer tx.Rollback()
	authorDAO := dao.NewArticleGORMAuthorDAO(tx)
	readerDAO := dao.NewArticleGORMReaderDAO(tx)
	artn := c.toEntity(art)
	var (
		id  = art.Id
		err error
	)
	if id > 0 {
		err = authorDAO.Update(ctx, artn)
	} else {
		id, err = authorDAO.Create(ctx, artn)
	}
	if err != nil {
		return 0, err
	}
	artn.Id = id
	err = readerDAO.UpsertV2(ctx, dao.PublishedArticle(artn))
	if err != nil {
		return 0, err
	}
	tx.Commit()
	return id, err
}

func NewCachedArticleRepository(dao dao.ArticleDao, userRepo UserRepository, cache cache.ArticleCache, l logger.LoggerV1) ArticleRepository {
	return &CachedArticleRepository{
		dao:      dao,
		cache:    cache,
		userRepo: userRepo,
		l:        l,
	}
}

func NewCachedArticleRepositoryV2(readerDAO dao.ArticleReaderDAO,
	authorDAO dao.ArticleAuthorDAO) *CachedArticleRepository {
	return &CachedArticleRepository{
		readerDAO: readerDAO,
		authorDAO: authorDAO,
	}
}

func (c *CachedArticleRepository) Create(ctx context.Context, art domain.Article) (int64, error) {
	id, err := c.dao.Insert(ctx, c.toEntity(art))
	if err == nil {
		er := c.cache.DelFirstPage(ctx, art.Author.Id)
		if er != nil {
			// 也要记录日志
			c.l.Error("插入新数据成功后，删除缓存第一页数据失败",
				logger.Int64("uid", art.Author.Id),
				logger.Error(err))
		}
	}
	return id, err
}

func (c *CachedArticleRepository) Update(ctx context.Context, art domain.Article) error {
	err := c.dao.UpdateById(ctx, c.toEntity(art))
	if err == nil {
		er := c.cache.DelFirstPage(ctx, art.Author.Id)
		if er != nil {
			// 也要记录日志
			c.l.Error("更新数据成功后，删除缓存第一页数据失败",
				logger.Int64("uid", art.Author.Id),
				logger.Error(err))
		}
	}
	return err
}

func (c *CachedArticleRepository) toEntity(art domain.Article) dao.Article {
	return dao.Article{
		Id:       art.Id,
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
		Status:   art.Status.ToUint8(),
	}
}

func (c *CachedArticleRepository) toDomain(art dao.Article) domain.Article {
	return domain.Article{
		Id:      art.Id,
		Title:   art.Title,
		Content: art.Content,
		Author: domain.Author{
			Id: art.AuthorId,
		},
		Ctime:  time.UnixMilli(art.Ctime),
		Utime:  time.UnixMilli(art.Utime),
		Status: domain.ArticleStatus(art.Status),
	}
}

func (c *CachedArticleRepository) preCache(ctx context.Context, arts []domain.Article) {
	const size = 1024 * 1024
	if len(arts) > 0 && len(arts[0].Content) < size {
		err := c.cache.Set(ctx, arts[0])
		if err != nil {
			// 记录缓存
		}
	}
}
