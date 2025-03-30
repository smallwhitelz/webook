package cache

import (
	"context"
	"github.com/redis/go-redis/v9"
)

var FolloweesNotFound = redis.Nil

type FeedEventCache interface {
	SetFollowees(ctx context.Context, follower int64, followees []int64) error
	GetFollowees(ctx context.Context, follower int64) ([]int64, error)
}
