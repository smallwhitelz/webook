package stress_test

import (
	"context"
	"encoding/json"
	"fmt"
	feedv1 "gitee.com/geekbang/basic-go/webook/api/proto/gen/feed/v1"
	followv1 "gitee.com/geekbang/basic-go/webook/api/proto/gen/follow/v1"
	followMocks "gitee.com/geekbang/basic-go/webook/api/proto/gen/follow/v1/mocks"
	"gitee.com/geekbang/basic-go/webook/feed/ioc"
	"gitee.com/geekbang/basic-go/webook/feed/repository"
	"gitee.com/geekbang/basic-go/webook/feed/repository/cache"
	"gitee.com/geekbang/basic-go/webook/feed/repository/dao"
	"gitee.com/geekbang/basic-go/webook/feed/service"
	"gitee.com/geekbang/basic-go/webook/feed/test"
	"gitee.com/geekbang/basic-go/webook/feed/test/stress_test/web"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"math/rand"
	"testing"
	"time"
)

// 生成拉事件
func generatePullEvent(mockFollowClient *followMocks.MockFollowServiceClient, id int64) test.ArticleEvent {
	mockFollowClient.EXPECT().GetFollowStatic(gomock.Any(), &followv1.GetFollowStaticRequest{
		Followee: id,
	}).Return(&followv1.GetFollowStaticResponse{
		FollowStatic: &followv1.FollowStatic{
			Followers: 1000,
		},
	}, nil)
	return test.ArticleEvent{
		Uid:   fmt.Sprintf("%d", id),
		Aid:   fmt.Sprintf("%d", time.Now().UnixNano()),
		Title: fmt.Sprintf("%d发布了文章", id),
	}
}

// 生成推事件
func generatePushEvent(mockFollowClient *followMocks.MockFollowServiceClient, id, i int64) test.ArticleEvent {
	// 生成几个推事件都包含id i
	mockFollowClient.EXPECT().GetFollowStatic(gomock.Any(), &followv1.GetFollowStaticRequest{
		Followee: id + i,
	}).Return(&followv1.GetFollowStaticResponse{
		FollowStatic: &followv1.FollowStatic{
			Followers: 2,
		},
	}, nil)
	mockFollowClient.EXPECT().GetFollower(gomock.Any(), &followv1.GetFollowerRequest{
		Followee: id + i,
	}).Return(&followv1.GetFollowerResponse{
		FollowRelations: []*followv1.FollowRelation{
			{
				Id:       time.Now().UnixNano(),
				Follower: id,
				Followee: id + i,
			},
			{
				Id:       time.Now().UnixNano(),
				Follower: id + i + 1,
				Followee: id + i,
			},
		},
	}, nil)
	return test.ArticleEvent{
		Uid:   fmt.Sprintf("%d", id+i),
		Aid:   fmt.Sprintf("%d", time.Now().UnixNano()),
		Title: fmt.Sprintf("%d发布了文章", id+i),
	}
}

// 生成数据
func Test_ADDFeed(t *testing.T) {
	server, followClient, _ := test.InitGrpcServer(t)
	//生成拉事件的压力测试的数据
	for i := 2; i < 100000; i++ {
		event := generatePullEvent(followClient, int64(i))
		ext, _ := json.Marshal(event)
		_, err := server.CreateFeedEvent(context.Background(), &feedv1.CreateFeedEventRequest{
			FeedEvent: &feedv1.FeedEvent{
				Type:    service.ArticleEventName,
				Content: string(ext),
			},
		})
		require.NoError(t, err)
	}

	//生成推事件的压力测试数据
	for i := 0; i < 100000; i++ {
		event := generatePushEvent(followClient, int64(300001), int64(i))
		ext, _ := json.Marshal(event)
		_, err := server.CreateFeedEvent(context.Background(), &feedv1.CreateFeedEventRequest{
			FeedEvent: &feedv1.FeedEvent{
				Type:    service.ArticleEventName,
				Content: string(ext),
			},
		})
		require.NoError(t, err)
	}

	//生成推拉事件的压力测试数据
	//拉事件服用上面拉事件的测试数据
	for i := 0; i < 100000; i++ {
		event := generatePushEvent(followClient, int64(400001), int64(i))
		ext, _ := json.Marshal(event)
		_, err := server.CreateFeedEvent(context.Background(), &feedv1.CreateFeedEventRequest{
			FeedEvent: &feedv1.FeedEvent{
				Type:    service.ArticleEventName,
				Content: string(ext),
			},
		})
		require.NoError(t, err)
	}
}

// 启动测试web
// 记得要把工作目录定位到这里
// 懒得再写一个测试的 IOC 了
func Test_Feed(t *testing.T) {
	viper.SetConfigFile("config.yaml")
	viper.WatchConfig()
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	loggerV1 := ioc.InitLogger()
	db := ioc.InitDB(loggerV1)
	feedPullEventDAO := dao.NewFeedPullEventDAO(db)
	feedPushEventDAO := dao.NewFeedPushEventDAO(db)
	cmdable := ioc.InitRedis()
	feedEventCache := cache.NewFeedEventCache(cmdable)
	feedEventRepo := repository.NewFeedEventRepo(feedPullEventDAO, feedPushEventDAO, feedEventCache)
	mockCtrl := gomock.NewController(t)
	followClient := followMocks.NewMockFollowServiceClient(mockCtrl)
	v := ioc.RegisterHandler(feedEventRepo, followClient)
	feedService := service.NewFeedService(feedEventRepo, v)
	engine := gin.Default()
	handler := web.NewFeedHandler(feedService)
	handler.RegisterRoutes(engine)
	// 设置mock数据
	//	设置关注列表的测试数据
	followClient.EXPECT().GetFollowee(gomock.Any(), gomock.Any()).Return(&followv1.GetFolloweeResponse{
		FollowRelations: getFollowRelation(1),
	}, nil).AnyTimes()
	// 设置粉丝列表的测试数据
	// 扩散百人
	followClient.EXPECT().GetFollowStatic(gomock.Any(), &followv1.GetFollowStaticRequest{
		Followee: 4,
	}).Return(&followv1.GetFollowStaticResponse{
		FollowStatic: &followv1.FollowStatic{
			Followers: 800,
		},
	}, nil).AnyTimes()
	followClient.EXPECT().GetFollower(gomock.Any(), &followv1.GetFollowerRequest{
		Followee: 4,
	}).Return(&followv1.GetFollowerResponse{
		FollowRelations: getFollowerRelation(4, 800),
	}, nil).AnyTimes()
	// 扩散千人
	followClient.EXPECT().GetFollowStatic(gomock.Any(), &followv1.GetFollowStaticRequest{
		Followee: 5,
	}).Return(&followv1.GetFollowStaticResponse{
		FollowStatic: &followv1.FollowStatic{
			Followers: 5000,
		},
	}, nil).AnyTimes()
	followClient.EXPECT().GetFollower(gomock.Any(), &followv1.GetFollowerRequest{
		Followee: 5,
	}).Return(&followv1.GetFollowerResponse{
		FollowRelations: getFollowerRelation(5, 5000),
	}, nil).AnyTimes()
	// 扩散万人
	followClient.EXPECT().GetFollowStatic(gomock.Any(), &followv1.GetFollowStaticRequest{
		Followee: 6,
	}).Return(&followv1.GetFollowStaticResponse{
		FollowStatic: &followv1.FollowStatic{
			Followers: 50000,
		},
	}, nil).AnyTimes()
	followClient.EXPECT().GetFollower(gomock.Any(), &followv1.GetFollowerRequest{
		Followee: 6,
	}).Return(&followv1.GetFollowerResponse{
		FollowRelations: getFollowerRelation(6, 10000),
	}, nil).AnyTimes()
	engine.Run("127.0.0.1:8088")
}

func getFollowRelation(id int64) []*followv1.FollowRelation {
	relations := make([]*followv1.FollowRelation, 0, 100001)
	random := rand.Intn(200) + 300
	for i := random - 200; i < random; i++ {
		relations = append(relations, &followv1.FollowRelation{
			Follower: id,
			Followee: int64(i),
		})
	}
	return relations
}

func getFollowerRelation(id int64, number int) []*followv1.FollowRelation {
	relations := make([]*followv1.FollowRelation, 0, 100001)
	for i := 1; i < number+1; i++ {
		relations = append(relations, &followv1.FollowRelation{
			Followee: id,
			Follower: int64(i),
		})
	}
	return relations
}
