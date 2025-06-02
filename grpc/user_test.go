package grpc

import (
	"context"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"testing"
)

func TestServer(t *testing.T) {
	// 常见一个grpc的服务
	gs := grpc.NewServer()
	// 创建一个UserService实现的实例
	us := &Server{}
	//注册一下
	RegisterUserServiceServer(gs, us)
	//创建一个监听网络端口的Listener
	l, err := net.Listen("tcp", ":8090")
	require.NoError(t, err)
	// 调用grpc Server上的Serve方法
	err = gs.Serve(l)
	t.Log(err)
}

func TestClient(t *testing.T) {
	// 初始化一个连接池（准确说，是池上池）
	cc, err := grpc.NewClient(":8090",
		// 测试环境，不想用https，所以用insecure.NewCredentials
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	// 初始化客户端
	client := NewUserServiceClient(cc)
	// 发起调用
	resp, err := client.GetByID(context.Background(), &GetByIDRequest{
		Id: 123,
	})
	require.NoError(t, err)
	t.Log(resp)
}

func TestOneOf(t *testing.T) {
	u := &User{}
	email, ok := u.Contacts.(*User_Email)
	if ok {
		t.Log("我传入的是 email", email)
		return
	}
}
