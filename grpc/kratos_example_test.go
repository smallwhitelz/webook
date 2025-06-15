package grpc

import (
	"context"
	etcd "github.com/go-kratos/kratos/contrib/registry/etcd/v2"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"testing"
	"time"
)

// KratosTestSuite kratos的服务注册与发现
type KratosTestSuite struct {
	suite.Suite
	etcdClient *etcdv3.Client
}

func (s *KratosTestSuite) SetupSuite() {
	cli, err := etcdv3.New(etcdv3.Config{
		Endpoints: []string{"localhost:12379"},
	})
	require.NoError(s.T(), err)
	s.etcdClient = cli
}

func (s *KratosTestSuite) TestClient() {

	r := etcd.New(s.etcdClient)
	cc, err := grpc.DialInsecure(context.Background(),
		grpc.WithEndpoint("discovery:///user"),
		grpc.WithDiscovery(r),
	)
	require.NoError(s.T(), err)
	defer cc.Close()

	client := NewUserServiceClient(cc)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	resp, err := client.GetByID(ctx, &GetByIDRequest{
		Id: 123,
	})
	require.NoError(s.T(), err)
	s.T().Log(resp.User)
}

// TestServer 启动服务器
func (s *KratosTestSuite) TestServer() {
	grpcSrv := grpc.NewServer(
		grpc.Address(":8090"),
		grpc.Middleware(recovery.Recovery()),
	)
	RegisterUserServiceServer(grpcSrv, &Server{})
	// etcd 注册中心
	r := etcd.New(s.etcdClient)
	app := kratos.New(
		kratos.Name("user"),
		kratos.Server(
			grpcSrv,
		),
		kratos.Registrar(r),
	)
	app.Run()
}

func TestKratos(t *testing.T) {
	suite.Run(t, new(KratosTestSuite))
}
