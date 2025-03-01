package grpc

import (
	"context"
	"google.golang.org/grpc"
	rewardv1 "webook/api/proto/gen/reward/v1"
	"webook/reward/service"
)

type RewardServiceServer struct {
	rewardv1.UnimplementedRewardServiceServer
	svc service.RewardService
}

func NewRewardServiceServer(svc service.RewardService) *RewardServiceServer {
	return &RewardServiceServer{svc: svc}
}

func (r *RewardServiceServer) Register(server *grpc.Server) {
	rewardv1.RegisterRewardServiceServer(server, r)
}

func (r *RewardServiceServer) PreReward(ctx context.Context, req *rewardv1.PreRewardRequest) (*rewardv1.PreRewardResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (r *RewardServiceServer) GetReward(ctx context.Context, req *rewardv1.GetRewardRequest) (*rewardv1.GetRewardResponse, error) {
	//TODO implement me
	panic("implement me")
}
