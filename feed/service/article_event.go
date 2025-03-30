package service

import (
	"context"
	"github.com/ecodeclub/ekit/slice"
	"time"
	followv1 "webook/api/proto/gen/follow/v1"
	"webook/feed/domain"
	"webook/feed/repository"
)

type ArticleEventHandler struct {
	repo         repository.FeedEventRepo
	followClient followv1.FollowServiceClient
}

func (a *ArticleEventHandler) CreateFeedEvent(ctx context.Context, ext domain.ExtendFields) error {
	uid, err := ext.Get("followee").AsInt64()
	if err != nil {
		return err
	}
	// 找到这个人的粉丝数量，判定是拉模型还是推模型
	resp, err := a.followClient.GetFollowStatic(ctx, &followv1.GetFollowStaticRequest{
		Followee: uid,
	})
	if err != nil {
		return err
	}
	// 大于一个阈值
	if resp.FollowStatic.Followers > threshold {
		// 拉模型
		return a.repo.CreatePullEvents(ctx, domain.FeedEvent{
			Uid: uid, Type: ArticleEventName, Ext: ext, Ctime: time.Now(),
		})
	} else {
		// 推模型，也就是写扩散
		// 先查询出来粉丝
		fresp, err := a.followClient.GetFollower(ctx, &followv1.GetFollowerRequest{
			Followee: uid,
		})
		if err != nil {
			return err
		}
		events := slice.Map(fresp.FollowRelations, func(idx int, src *followv1.FollowRelation) domain.FeedEvent {
			return domain.FeedEvent{Uid: src.Follower, Type: ArticleEventName, Ext: ext, Ctime: time.Now()}
		})
		return a.repo.CreatePushEvents(ctx, events)
	}
}

func (a *ArticleEventHandler) FindFeedEvents(ctx context.Context, uid, timestamp, limit int64) ([]domain.FeedEvent, error) {
	//TODO implement me
	panic("implement me")
}

const (
	ArticleEventName = "article_event"
	threshold        = 4
	//threshold        = 32
)

func NewArticleEventHandler(repo repository.FeedEventRepo, client followv1.FollowServiceClient) Handler {
	return &ArticleEventHandler{
		repo:         repo,
		followClient: client,
	}
}
