package grpc

import (
	"context"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"testing"
	"time"
)

type InterceptorTestSuite struct {
	suite.Suite
}

func (s *InterceptorTestSuite) TestClient() {
	t := s.T()
	cc, err := grpc.Dial("localhost:8090", grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	client := NewUserServiceClient(cc)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	time.Sleep(time.Millisecond * 100)
	resp, err := client.GetByID(ctx, &GeyByIDRequest{Id: 123})
	require.NoError(t, err)
	t.Log(resp)
}

func (s *InterceptorTestSuite) TestServer() {
	t := s.T()
	server := grpc.NewServer(grpc.ChainUnaryInterceptor(
		NewLogInterceptor(t)))
	RegisterUserServiceServer(server, &Server{
		Name: "interceptor_test",
	})

	// 限流单个业务接口
	//RegisterUserServiceServer(server, &LimiterUserServer{
	//	UserServiceServer: &Server{
	//		Name: "interceptor_test",
	//	},
	//})
	l, err := net.Listen("tcp", ":8090")
	require.NoError(t, err)
	server.Serve(l)
}
func NewLogInterceptor(t *testing.T) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		t.Log("请求处理前", req, info)
		resp, err = handler(ctx, req)
		t.Log("请求处理后", resp, err)
		return
	}
}

func TestInterceptorTestSuite(t *testing.T) {
	suite.Run(t, new(InterceptorTestSuite))
}
