package client

import (
	"context"
	"github.com/ecodeclub/ekit/syncx/atomicx"
	grpc "google.golang.org/grpc"
	"math/rand"
	intrv1 "webook/api/proto/gen/intr/v1"
)

type InteractiveClient struct {
	remote intrv1.InteractiveServiceClient
	local  intrv1.InteractiveServiceClient

	// 阈值，用原子操作，因为可能会出现读写同时的情况
	threshold *atomicx.Value[int32]
}

func (i *InteractiveClient) IncrReadCnt(ctx context.Context, in *intrv1.IncrReadCntRequest, opts ...grpc.CallOption) (*intrv1.IncrReadCntResponse, error) {
	return i.SelectClient().IncrReadCnt(ctx, in, opts...)
}

func (i *InteractiveClient) Like(ctx context.Context, in *intrv1.LikeRequest, opts ...grpc.CallOption) (*intrv1.LikeResponse, error) {
	return i.SelectClient().Like(ctx, in, opts...)
}

func (i *InteractiveClient) CancelLike(ctx context.Context, in *intrv1.CancelLikeRequest, opts ...grpc.CallOption) (*intrv1.CancelLikeResponse, error) {
	return i.SelectClient().CancelLike(ctx, in, opts...)
}

func (i *InteractiveClient) Collect(ctx context.Context, in *intrv1.CollectRequest, opts ...grpc.CallOption) (*intrv1.CollectResponse, error) {
	return i.SelectClient().Collect(ctx, in, opts...)
}

func (i *InteractiveClient) Get(ctx context.Context, in *intrv1.GetRequest, opts ...grpc.CallOption) (*intrv1.GetResponse, error) {
	return i.SelectClient().Get(ctx, in, opts...)
}

func (i *InteractiveClient) GetByIds(ctx context.Context, in *intrv1.GetByIdsRequest, opts ...grpc.CallOption) (*intrv1.GetByIdsResponse, error) {
	return i.SelectClient().GetByIds(ctx, in, opts...)
}

// SelectClient 控制选择远程调用还是本地调用
func (i *InteractiveClient) SelectClient() intrv1.InteractiveServiceClient {
	//[0,100)的随机数
	num := rand.Int31n(100)
	// 如果随机数小于阈值，就用远程调用
	if num < i.threshold.Load() {
		return i.remote
	}
	return i.local
}

// UpdateThreshold 暴露一个方法进行修改阈值
func (i *InteractiveClient) UpdateThreshold(val int32) {
	i.threshold.Store(val)
}

func NewInteractiveClient(remote intrv1.InteractiveServiceClient, local intrv1.InteractiveServiceClient) *InteractiveClient {
	return &InteractiveClient{remote: remote, local: local, threshold: atomicx.NewValue[int32]()}
}
