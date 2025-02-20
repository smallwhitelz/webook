package repository

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"
	"webook/internal/domain"
	"webook/internal/repository/cache"
	"webook/internal/repository/dao"
)

var (
	ErrDuplicateUser = dao.ErrDuplicateEmail
	ErrUserNotFound  = dao.ErrRecordNotFound
)

type UserRepository interface {
	Create(ctx context.Context, user domain.User) error
	FindByEmail(ctx context.Context, email string) (domain.User, error)
	FindById(ctx context.Context, uid int64) (domain.User, error)
	FindByPhone(ctx context.Context, phone string) (domain.User, error)
	UpdateNonSensitiveInfo(ctx context.Context, user domain.User) error
	FindByWechat(ctx context.Context, openId string) (domain.User, error)
}

type CachedUserRepository struct {
	dao   dao.UserDAO
	cache cache.UserCache
}

func (repo *CachedUserRepository) FindByWechat(ctx context.Context, openId string) (domain.User, error) {
	ue, err := repo.dao.FindByWechat(ctx, openId)
	if err != nil {
		return domain.User{}, err
	}
	return repo.toDomain(ue), nil
}

//type DBConfig struct {
//	DSN string
//}
//
//type CacheConfig struct {
//	Addr string
//}

// NewUserRepositoryV1 非依赖注入写法，强耦合（跨层的），严重缺乏扩展性
//func NewUserRepositoryV1(dbCfg DBConfig, cCfg CacheConfig) (*CachedUserRepository, error) {
//	db, err := gorm.Open(mysql.Open(dbCfg.DSN))
//	if err != nil {
//		return nil, err
//	}
//	ud := dao.NewUserDao(db)
//	uc := cache.NewUserCache(redis.NewClient(&redis.Options{Addr: cCfg.Addr}))
//	return &CachedUserRepository{
//		dao:   ud,
//		cache: uc,
//	}, nil
//}

func NewCachedUserRepository(dao dao.UserDAO, c cache.UserCache) UserRepository {
	return &CachedUserRepository{
		dao:   dao,
		cache: c,
	}
}

func (repo *CachedUserRepository) Create(ctx context.Context, user domain.User) error {
	return repo.dao.Insert(ctx, repo.toEntity(user))
}

func (repo *CachedUserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	u, err := repo.dao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	return repo.toDomain(u), nil
}

func (repo *CachedUserRepository) toDomain(u dao.User) domain.User {
	return domain.User{
		Id:          u.Id,
		Email:       u.Email.String,
		Phone:       u.Phone.String,
		Password:    u.Password,
		Nickname:    u.Nickname,
		Birthday:    time.UnixMilli(u.Birthday),
		Description: u.Description,
		Ctime:       time.UnixMilli(u.Ctime),
		WechatInfo: domain.WechatInfo{
			OpenId:  u.WechatOpenId.String,
			UnionId: u.WechatUnionId.String,
		},
	}
}

func (repo *CachedUserRepository) toEntity(u domain.User) dao.User {
	return dao.User{
		Id: u.Id,
		Email: sql.NullString{
			String: u.Email,
			Valid:  u.Email != "",
		},
		Phone: sql.NullString{
			String: u.Phone,
			Valid:  u.Phone != "",
		},
		Password: u.Password,
		WechatUnionId: sql.NullString{
			String: u.WechatInfo.UnionId,
			Valid:  u.WechatInfo.UnionId != "",
		},
		WechatOpenId: sql.NullString{
			String: u.WechatInfo.OpenId,
			Valid:  u.WechatInfo.OpenId != "",
		},
		Nickname:    u.Nickname,
		Birthday:    u.Birthday.UnixMilli(),
		Description: u.Description,
	}
}

func (repo *CachedUserRepository) FindById(ctx context.Context, uid int64) (domain.User, error) {
	du, err := repo.cache.Get(ctx, uid)
	// 只要 err为 nil，就返回
	if err == nil {
		return du, nil
	}

	// 检测限流/熔断/降级标记位
	if ctx.Value("downgrade") == "true" {
		return du, errors.New("触发降级，不在查询数据库")
	}
	// err 不为 nil，就要查询数据库
	// err 有两种可能
	// 1. key 不存在，说明redis是正常的
	// 2. 访问redis有问题。可能是网络有问题，也有可能是 redis 本身就崩溃了
	u, err := repo.dao.FindById(ctx, uid)
	if err != nil {
		return domain.User{}, err
	}
	du = repo.toDomain(u)
	go func() {
		err = repo.cache.Set(ctx, du)
		if err != nil {
			log.Println(err)
		}
	}()
	err = repo.cache.Set(ctx, du)
	if err != nil {
		// 网络崩了，也可能是 redis 崩了
		log.Println(err)
	}
	return du, nil
}

func (repo *CachedUserRepository) FindByIdV1(ctx context.Context, uid int64) (domain.User, error) {
	du, err := repo.cache.Get(ctx, uid)
	// 只要 err为 nil，就返回
	switch err {
	case nil:
		return du, nil
	case cache.ErrKeyNotExist:
		u, err := repo.dao.FindById(ctx, uid)
		if err != nil {
			return domain.User{}, err
		}
		du = repo.toDomain(u)
		go func() {
			err = repo.cache.Set(ctx, du)
			if err != nil {
				log.Println(err)
			}
		}()
		err = repo.cache.Set(ctx, du)
		if err != nil {
			// 网络崩了，也可能是 redis 崩了
			log.Println(err)
		}
		return du, nil
	default:
		// 接近降级的写法
		return domain.User{}, err
	}
}

func (repo *CachedUserRepository) UpdateNonSensitiveInfo(ctx context.Context, user domain.User) error {
	err := repo.dao.UpdateById(ctx, repo.toEntity(user))
	if err != nil {
		return err
	}
	// 更新成功后，找到这个用户
	u, err := repo.dao.FindById(ctx, user.Id)
	if err != nil {
		return err
	}
	// 更新缓存
	return repo.cache.Set(ctx, repo.toDomain(u))
}

func (repo *CachedUserRepository) FindByPhone(ctx context.Context, phone string) (domain.User, error) {
	u, err := repo.dao.FindByPhone(ctx, phone)
	if err != nil {
		return domain.User{}, err
	}
	return repo.toDomain(u), nil
}
