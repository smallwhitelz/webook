package service

import (
	"context"
	"fmt"
	"github.com/ecodeclub/ekit/slice"
	"golang.org/x/sync/errgroup"
	"sort"
	"sync"
	followv1 "webook/api/proto/gen/follow/v1"
	"webook/feed/domain"
	"webook/feed/repository"
)

type feedService struct {
	repo repository.FeedEventRepo
	// 对应的 string 就是 type
	handlerMap   map[string]Handler
	followClient followv1.FollowServiceClient
}

func NewFeedService(repo repository.FeedEventRepo, handlerMap map[string]Handler) FeedService {
	return &feedService{
		repo:       repo,
		handlerMap: handlerMap,
	}
}

// RegisterService 这里的作用相当于给handlerMap赋值，在ioc的时候可以直接用NewFeedService，也可以不在这里初始化，
// 然后在 RegisterService 初始化
func (f *feedService) RegisterService(typ string, handler Handler) {
	f.handlerMap[typ] = handler
}

func (f *feedService) CreateFeedEvent(ctx context.Context, feed domain.FeedEvent) error {
	handler, ok := f.handlerMap[feed.Type]
	if !ok {
		// 说明type不对
		// 你还可以考虑兜底机制
		// 有一个 defaultHandler，然后调用 defaultHandler
		return fmt.Errorf("未能找到对应的 Handler %s", feed.Type)
	}
	return handler.CreateFeedEvent(ctx, feed.Ext)
}

// GetFeedEventListV1 不依赖handler的直接查询
func (f *feedService) GetFeedEventListV1(ctx context.Context, uid, timestamp, limit int64) ([]domain.FeedEvent, error) {
	// 直接查询
	var eg errgroup.Group
	var lock sync.Mutex
	events := make([]domain.FeedEvent, 0, limit*2)
	eg.Go(func() error {
		// 查询发件箱
		resp, err := f.followClient.GetFollowee(ctx, &followv1.GetFolloweeRequest{Follower: uid, Limit: 10000})
		if err != nil {
			return err
		}
		followeeIDs := slice.Map(resp.FollowRelations, func(idx int, src *followv1.FollowRelation) int64 {
			return src.Followee
		})
		evts, err := f.repo.FindPullEvents(ctx, followeeIDs, timestamp, limit)
		if err != nil {
			return err
		}
		lock.Lock()
		events = append(events, evts...)
		lock.Unlock()
		return nil
	})
	eg.Go(func() error {
		evts, err := f.repo.FindPushEvents(ctx, uid, timestamp, limit)
		if err != nil {
			return err
		}
		lock.Lock()
		events = append(events, evts...)
		lock.Unlock()
		return nil
	})
	err := eg.Wait()
	if err != nil {
		return nil, err
	}
	// 你已经查询到所有数据，现在要排序了
	sort.Slice(events, func(i, j int) bool {
		return events[i].Ctime.UnixMilli() > events[j].Ctime.UnixMilli()
	})
	// 高版本可以直接用Go内置方法min()
	minVal := slice.Min[int]([]int{int(limit), len(events)})
	return events[:minVal], nil
}

func (f *feedService) GetFeedEventList(ctx context.Context, uid, timestamp, limit int64) ([]domain.FeedEvent, error) {
	var eg errgroup.Group
	var lock sync.Mutex
	events := make([]domain.FeedEvent, 0, limit*int64(len(f.handlerMap)))
	for _, handler := range f.handlerMap {
		h := handler
		eg.Go(func() error {
			evts, err := h.FindFeedEvents(ctx, uid, timestamp, limit)
			if err != nil {
				return err
			}
			lock.Lock()
			events = append(events, evts...)
			lock.Unlock()
			return nil
		})
	}
	err := eg.Wait()
	if err != nil {
		return nil, err
	}
	sort.Slice(events, func(i, j int) bool {
		return events[i].Ctime.UnixMilli() > events[j].Ctime.UnixMilli()
	})
	return events[:min[int](int(limit), len(events))], nil
}

// GetFeedEventListV1Homework 用这个来演示作业：支持活跃用户
func (f *feedService) GetFeedEventListV1Homework(ctx context.Context, uid int64, timestamp, limit int64) ([]domain.FeedEvent, error) {
	// 直接查询
	var eg errgroup.Group
	var lock sync.Mutex
	events := make([]domain.FeedEvent, 0, limit*2)
	eg.Go(func() error {
		// 在这里判定是否是活跃用户。如果是活跃用户，我们认为它不需要查询发件箱
		// 所以问题就在于你要设计对应的判定活跃用户的算法
		// 直接返回
		if f.isActiveUser(uid) {
			return nil
		}
		// 查询发件箱
		resp, err := f.followClient.GetFollowee(ctx, &followv1.GetFolloweeRequest{Follower: uid, Limit: 10000})
		if err != nil {
			return err
		}
		followeeIDs := slice.Map(resp.FollowRelations, func(idx int, src *followv1.FollowRelation) int64 {
			return src.Followee
		})
		evts, err := f.repo.FindPullEvents(ctx, followeeIDs, timestamp, limit)
		if err != nil {
			return err
		}
		lock.Lock()
		events = append(events, evts...)
		lock.Unlock()
		return nil
	})

	eg.Go(func() error {
		evts, err := f.repo.FindPushEvents(ctx, uid, timestamp, limit)
		if err != nil {
			return err
		}
		lock.Lock()
		events = append(events, evts...)
		lock.Unlock()
		return nil
	})

	err := eg.Wait()
	if err != nil {
		return nil, err
	}
	// 你已经查询所有的数据，现在要排序
	sort.Slice(events, func(i, j int) bool {
		return events[i].Ctime.UnixMilli() > events[j].Ctime.UnixMilli()
	})
	return events[:slice.Min[int]([]int{int(limit), len(events)})], nil
}

func (f *feedService) isActiveUser(uid int64) bool {
	// 在实践中，是否是活跃用户，一般都是离线任务计算的。
	// 比如说每天计算一批，或者间隔一段时间计算一批
	// 可以考虑采用连续登录之类的方案
	// 而后在判定是否是活跃用户的时候，可以利用 redis 来实现 bit array 或者布隆过滤器
	// 并且，在布隆过滤器假阳性的情况下，你可以当成真的来处理
	// 在这种时候，你只会把一些不是活跃用户的判定为活跃用户，
	// 但是这种代价是可以接受的
	return false
}
