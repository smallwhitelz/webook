package cache

import (
	"context"
	"errors"
	"fmt"
	lru "github.com/hashicorp/golang-lru"
	"sync"
	"time"
)

/**
  技术选型考虑的点
  1. 功能性：功能是否能够完全覆盖你的需求。
  2. 社区和支持度：社区是否活跃，文档是否齐全
     以及搜索引擎能不能搜到你需要的各种信息，有没有帮你踩过坑
  3. 非功能性：易用性（用户友好度，学习曲线要平滑）
     扩展性（如果开源软件的某些功能需要定制，框架是否支持定制，以及定制的难度高不高）
     性能（追求性能的公司，往往有能力自研）
*/

// LocalCodeCache 本地缓存实现
type LocalCodeCache struct {
	cache *lru.Cache
	// 普通锁，或者说写锁
	lock sync.Mutex
	// 读写锁
	expiration time.Duration
}

func (l *LocalCodeCache) Set(ctx context.Context, biz, phone, code string) error {
	l.lock.Lock()
	defer l.lock.Unlock()
	// 这里可以考虑用读写锁来优化，但是效果不是很好
	// 因为你可以预期，大部分时候是要走到写锁里面的

	// 选用本地缓存，很不幸的是，没有获得过期时间的接口，所以要自己维持一个过期时间的字段
	key := l.key(biz, phone)
	now := time.Now()
	val, ok := l.cache.Get(key)
	if !ok {
		// 说明没有验证码
		l.cache.Add(key, codeItem{
			code:   code,
			cnt:    3,
			expire: now.Add(l.expiration),
		})
		return nil
	}
	itm, ok := val.(codeItem)
	if !ok {
		// 理论来说是不可能走到这的
		return errors.New("系统错误")
	}
	if itm.expire.Sub(now) > time.Minute*9 {
		// 不到一分钟
		return ErrCodeSendTooMany
	}
	// 重发
	l.cache.Add(key, codeItem{
		code:   code,
		cnt:    3,
		expire: now.Add(l.expiration),
	})
	return nil
}

func (l *LocalCodeCache) Verify(ctx context.Context, biz, phone, inputCode string) (bool, error) {
	l.lock.Lock()
	// 尽量都用这种写法，因为go中panic掉之后，是不会释放锁的
	defer l.lock.Unlock()
	key := l.key(biz, phone)
	val, ok := l.cache.Get(key)
	if !ok {
		// 没有发送验证码
		return false, ErrKeyNotExist
	}
	itm, ok := val.(codeItem)
	if !ok {
		// 理论来说是不可能走到这的
		return false, errors.New("系统错误")
	}
	if itm.cnt < 0 {
		return false, ErrCodeVerifyTooMany
	}
	itm.cnt--
	return itm.code == inputCode, nil
}

func (l *LocalCodeCache) key(biz string, phone string) string {
	return fmt.Sprintf("phone_code:%s:%s", biz, phone)
}

func NewLocalCodeCache(c *lru.Cache, expiration time.Duration) *LocalCodeCache {
	return &LocalCodeCache{
		cache:      c,
		expiration: expiration,
	}
}

type codeItem struct {
	code string
	// 可验证次数
	cnt int
	// 过期时间
	expire time.Time
}
