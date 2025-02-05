package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
	"webook/internal/domain"
)

var ErrKeyNotExist = redis.Nil

type UserCache interface {
	Get(ctx context.Context, uid int64) (domain.User, error)
	Set(ctx context.Context, du domain.User) error
}

type RedisUserCache struct {
	cmd        redis.Cmdable
	expiration time.Duration
}

func (c *RedisUserCache) Get(ctx context.Context, uid int64) (domain.User, error) {
	key := c.key(uid)
	// 我假定这个地方用用json来
	data, err := c.cmd.Get(ctx, key).Result()
	if err != nil {
		return domain.User{}, err
	}
	var u domain.User
	err = json.Unmarshal([]byte(data), &u)
	return u, err
}

func (c *RedisUserCache) key(uid int64) string {
	return fmt.Sprintf("user:info:%d", uid)
}

func (c *RedisUserCache) Set(ctx context.Context, du domain.User) error {
	key := c.key(du.Id)
	// 我假定这个地方用用json来
	data, err := json.Marshal(du)
	if err != nil {
		return err
	}
	return c.cmd.Set(ctx, key, data, c.expiration).Err()
}

// 不建议这样使用，Client是一个结构体，实现了Cmdable接口，
// 如果是单体应用，可以直接使用Client，如果是微服务，使用集群就要使用ClusterClient
//type UserCacheV1 struct {
//	client redis.Client
//}

func NewUserCache(cmd redis.Cmdable) UserCache {
	return &RedisUserCache{
		cmd:        cmd,
		expiration: time.Minute * 15,
	}
}

// 一定不要自己去初始化你需要的东西，让外面传进来
//func NewUserCacheV1(addr string) *RedisUserCache {
//	cmd := redis.NewClient(&redis.Options{Addr: addr})
//	return &RedisUserCache{
//		cmd:        cmd,
//		expiration: time.Minute * 15,
//	}
//}
