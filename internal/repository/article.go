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
			c.l.Error("读者接口，放入文章缓存失败", logger.Int64("aid", id), logger.Error(er))
		}
	}()
	return res, nil
}

// GetById 获取单个文章详情接口
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
			c.l.Error("设置文章详情缓存失败", logger.Int64("aid", id), logger.Error(er))
		}
	}()
	return res, nil
}

// GetByAuthor 创作者查询列表接口
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
			c.l.Error("查询缓存第一页失败", logger.Int64("uid", uid), logger.Error(err))
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
				c.l.Error("缓存第一页失败", logger.Int64("uid", uid), logger.Error(err))
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

// SyncStatus 同步文章状态接口，这里是将文章设置为不可见
func (c *CachedArticleRepository) SyncStatus(ctx context.Context, uid int64, id int64, status domain.ArticleStatus) error {
	err := c.dao.SyncStatus(ctx, uid, id, status.ToUint8())
	if err == nil {
		er := c.cache.DelFirstPage(ctx, uid)
		if er != nil {
			// 也要记录日志
			c.l.Error("删除缓存第一页失败",
				logger.Int64("uid", uid),
				logger.Error(er))
		}
	}
	return err
}

// Sync 在dao层操作制作库和线上库，用于发表文章
func (c *CachedArticleRepository) Sync(ctx context.Context, art domain.Article) (int64, error) {
	id, err := c.dao.Sync(ctx, c.toEntity(art))
	if err == nil {
		er := c.cache.DelFirstPage(ctx, art.Author.Id)
		if er != nil {
			// 也要记录日志
			c.l.Error("删除缓存第一页失败",
				logger.Int64("aid", art.Id),
				logger.Int64("uid", art.Author.Id),
				logger.Error(er))
		}
	}
	// 在这里尝试设置缓存
	// 我们认为，当一条帖子发表，他会立刻有人访问，所以在这里也可以设置缓存
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		user, er := c.userRepo.FindById(ctx, art.Author.Id)

		if er != nil {
			// 记录日志
			c.l.Error("发表帖子查询作者失败", logger.Int64("aid", art.Id), logger.Error(er))
			return
		}
		art.Author = domain.Author{
			Id:   user.Id,
			Name: user.Nickname,
		}
		// 你可以灵活设置过期时间
		// 1. 比如可以给大V设置超长过期时间，因为粉丝多，短时间发一篇文章，粉丝在几天内都会读完
		// 2. 如果是一个普通写手，文章很久才被人偶然访问一次，就可以设置短一点的过期时间
		er = c.cache.SetPub(ctx, art)
		if er != nil {
			// 记录日志
			c.l.Error("发表帖子设置缓存失败", logger.Int64("aid", art.Id), logger.Error(er))
		}
	}()
	return id, err
}

// Create 新增文章
func (c *CachedArticleRepository) Create(ctx context.Context, art domain.Article) (int64, error) {
	id, err := c.dao.Insert(ctx, c.toEntity(art))
	if err == nil {
		er := c.cache.DelFirstPage(ctx, art.Author.Id)
		if er != nil {
			// 也要记录日志
			c.l.Error("新增新数据成功后，删除缓存第一页数据失败",
				logger.Int64("aid", id),
				logger.Int64("uid", art.Author.Id),
				logger.Error(err))
		}
	}
	return id, err
}

// 修改文章
func (c *CachedArticleRepository) Update(ctx context.Context, art domain.Article) error {
	err := c.dao.UpdateById(ctx, c.toEntity(art))
	if err == nil {
		er := c.cache.DelFirstPage(ctx, art.Author.Id)
		if er != nil {
			// 也要记录日志
			c.l.Error("更新数据成功后，删除缓存第一页数据失败",
				logger.Int64("aid", art.Id),
				logger.Int64("uid", art.Author.Id),
				logger.Error(err))
		}
	}
	return err
}

func NewCachedArticleRepository(dao dao.ArticleDao, userRepo UserRepository, cache cache.ArticleCache, l logger.LoggerV1) ArticleRepository {
	return &CachedArticleRepository{
		dao:      dao,
		cache:    cache,
		userRepo: userRepo,
		l:        l,
	}
}

// SyncV1 非事务实现，不同数据库
// repo层同步制作库和线上库数据，操作两个dao
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

// SyncV2 开启事务，同一个数据库不同表的实现
// 缺陷：强制Repository引入db，相当于Repository强制依赖了DAO依赖的东西
// 没有坚持面向接口编程原则
// 跨层依赖
func (c *CachedArticleRepository) SyncV2(ctx context.Context, art domain.Article) (int64, error) {
	// 开启事务
	tx := c.db.WithContext(ctx).Begin()
	// 检测事务是否开启成功
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
	// 提交事务
	tx.Commit()
	return id, err
}

// NewCachedArticleRepositoryV2 这里是在repo层分发
// 操作的是dao层
func NewCachedArticleRepositoryV2(readerDAO dao.ArticleReaderDAO,
	authorDAO dao.ArticleAuthorDAO) *CachedArticleRepository {
	return &CachedArticleRepository{
		readerDAO: readerDAO,
		authorDAO: authorDAO,
	}
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

// preCache 预加载，我们猜测，客户加载完列表一般情况下会访问第一条数据，所以我们默认缓存第一条
// 这在高并发下场景下非常有用
func (c *CachedArticleRepository) preCache(ctx context.Context, arts []domain.Article) {
	const size = 1024 * 1024
	// 不缓存大对象，没有意义
	// 但是也不是说所有的大对象都不缓存，一定还是根据业务的相关性进行抉择
	if len(arts) > 0 && len(arts[0].Content) < size {
		err := c.cache.Set(ctx, arts[0])
		if err != nil {
			// 记录缓存
			c.l.Error("缓存列表第一条数据失败",
				logger.Int64("aid", arts[0].Id),
				logger.Error(err))
		}
	}
}
