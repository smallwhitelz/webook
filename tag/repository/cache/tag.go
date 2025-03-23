package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"strconv"
	"time"
	"webook/tag/domain"
)

var ErrKeyNotExist = redis.Nil

type TagCache interface {
	Append(ctx context.Context, uid int64, tags ...domain.Tag) error
	GetTags(ctx context.Context, uid int64) ([]domain.Tag, error)
}

type RedisTagCache struct {
	client     redis.Cmdable
	expiration time.Duration
}

func (r *RedisTagCache) GetTags(ctx context.Context, uid int64) ([]domain.Tag, error) {
	data, err := r.client.HGetAll(ctx, r.userTagsKey(uid)).Result()
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, ErrKeyNotExist
	}
	res := make([]domain.Tag, 0, len(data))
	for _, val := range data {
		var t domain.Tag
		err := json.Unmarshal([]byte(val), &t)
		if err != nil {
			return nil, err
		}
		res = append(res, t)
	}
	return res, nil
}

func (r *RedisTagCache) Append(ctx context.Context, uid int64, tags ...domain.Tag) error {
	// 要放我的标签
	key := r.userTagsKey(uid)
	pip := r.client.Pipeline()
	for _, tag := range tags {
		val, err := json.Marshal(tag)
		if err != nil {
			return err
		}
		// uid => tid_0 -> t0
		//        tid_1 -> t1
		pip.HMSet(ctx, key, strconv.FormatInt(tag.Id, 10), val)
	}
	// 你也可以考虑永不过期
	pip.Expire(ctx, key, r.expiration)
	_, err := pip.Exec(ctx)
	return err
}

func (r *RedisTagCache) userTagsKey(uid int64) string {
	return fmt.Sprintf("tag:user_tags:%d", uid)
}

func NewRedisTagCache(client redis.Cmdable, expiration time.Duration) TagCache {
	return &RedisTagCache{client: client, expiration: expiration}
}
