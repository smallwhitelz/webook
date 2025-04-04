package test

import (
	feedv1 "gitee.com/geekbang/basic-go/webook/api/proto/gen/feed/v1"
	followMocks "gitee.com/geekbang/basic-go/webook/api/proto/gen/follow/v1/mocks"
	"gitee.com/geekbang/basic-go/webook/feed/grpc"
	"gitee.com/geekbang/basic-go/webook/feed/ioc"
	"gitee.com/geekbang/basic-go/webook/feed/repository"
	"gitee.com/geekbang/basic-go/webook/feed/repository/cache"
	"gitee.com/geekbang/basic-go/webook/feed/repository/dao"
	"gitee.com/geekbang/basic-go/webook/feed/service"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"
	"testing"
)

func InitGrpcServer(t *testing.T) (feedv1.FeedSvcServer, *followMocks.MockFollowServiceClient, *gorm.DB) {
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
	feedEventGrpcSvc := grpc.NewFeedEventGrpcSvc(feedService)
	return feedEventGrpcSvc, followClient, db
}
