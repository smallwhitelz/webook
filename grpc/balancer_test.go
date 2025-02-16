package grpc

import (
	"context"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	_ "google.golang.org/grpc/balancer/weightedroundrobin"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"testing"
	"time"
	_ "webook/pkg/grpcx/balancer/wrr"
)

type BalancerTestSuite struct {
	suite.Suite
	cli *etcdv3.Client
}

func (s *BalancerTestSuite) SetupSuite() {
	//cli, err := etcdv3.NewFromURL("localhost:12379")
	cli, err := etcdv3.New(etcdv3.Config{
		Endpoints: []string{"43.154.97.245:12379"},
		Username:  "root",
		Password:  "1234",
	})
	require.NoError(s.T(), err)
	s.cli = cli
}

func (s *BalancerTestSuite) TestClientCustomWRR() {
	t := s.T()
	etcdResolver, err := resolver.NewBuilder(s.cli)
	require.NoError(s.T(), err)
	cc, err := grpc.Dial("etcd:///service/user",
		grpc.WithResolvers(etcdResolver),
		grpc.WithDefaultServiceConfig(`
{
   "loadBalancingConfig": [
      {
         "custom_weighted_round_robin": {}
      }
   ]
}

`),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	client := NewUserServiceClient(cc)
	for i := 0; i < 10; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		resp, err := client.GetByID(ctx, &GeyByIDRequest{Id: 123})
		cancel()
		require.NoError(t, err)
		t.Log(resp)
	}
}

func (s *BalancerTestSuite) TestFailoverClient() {
	t := s.T()
	etcdResolver, err := resolver.NewBuilder(s.cli)
	require.NoError(s.T(), err)
	cc, err := grpc.Dial("etcd:///service/user",
		grpc.WithResolvers(etcdResolver),
		grpc.WithDefaultServiceConfig(`
{
  "loadBalancingConfig": [{"round_robin": {}}],
  "methodConfig":  [
    {
      "name": [{"service":  "UserService"}],
      "retryPolicy": {
        "maxAttempts": 4,
        "initialBackoff": "0.01s",
        "maxBackoff": "0.1s",
        "backoffMultiplier": 2.0,
        "retryableStatusCodes": ["UNAVAILABLE"]
      }
    }
  ]
}

`),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	client := NewUserServiceClient(cc)
	for i := 0; i < 10; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		resp, err := client.GetByID(ctx, &GeyByIDRequest{Id: 123})
		cancel()
		require.NoError(t, err)
		t.Log(resp)
	}
}

func (s *BalancerTestSuite) TestClientWRR() {
	t := s.T()
	etcdResolver, err := resolver.NewBuilder(s.cli)
	require.NoError(s.T(), err)
	cc, err := grpc.Dial("etcd:///service/user",
		grpc.WithResolvers(etcdResolver),
		grpc.WithDefaultServiceConfig(`
{
   "loadBalancingConfig": [
      {
         "weighted_round_robin": {}
      }
   ]
}

`),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	client := NewUserServiceClient(cc)
	for i := 0; i < 10; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		resp, err := client.GetByID(ctx, &GeyByIDRequest{Id: 123})
		cancel()
		require.NoError(t, err)
		t.Log(resp)
	}
}

func (s *BalancerTestSuite) TestClient() {
	t := s.T()
	etcdResolver, err := resolver.NewBuilder(s.cli)
	require.NoError(s.T(), err)
	cc, err := grpc.Dial("etcd:///service/user",
		grpc.WithResolvers(etcdResolver),
		grpc.WithDefaultServiceConfig(`
{
   "loadBalancingConfig": [
      {
         "round_robin": {}
      }
   ]
}

`),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	client := NewUserServiceClient(cc)
	for i := 0; i < 10; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		resp, err := client.GetByID(ctx, &GeyByIDRequest{Id: 123})
		cancel()
		require.NoError(t, err)
		t.Log(resp)
	}
}

func (s *BalancerTestSuite) TestServer() {
	go func() {
		s.startServer(":8090", 10, &Server{
			Name: ":8090",
		})
	}()
	go func() {
		s.startServer(":8091", 20, &Server{
			Name: ":8091",
		})
	}()
	s.startServer(":8092", 30, &FailedServer{
		Name: ":8092",
	})
}

func (s *BalancerTestSuite) startServer(addr string, weight int, svc UserServiceServer) {
	t := s.T()
	em, err := endpoints.NewManager(s.cli, "service/user")
	require.NoError(s.T(), err)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	addr = "127.0.0.1" + addr
	key := "service/user/" + addr
	l, err := net.Listen("tcp", addr)
	require.NoError(s.T(), err)
	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	// 租期
	var ttl int64 = 5
	leaseResp, err := s.cli.Grant(ctx, ttl)
	require.NoError(t, err)
	err = em.AddEndpoint(ctx, key, endpoints.Endpoint{
		// 定位信息，客户端如何连你
		Addr: addr,
		Metadata: map[string]any{
			"weight": weight,
		},
	}, etcdv3.WithLease(leaseResp.ID))
	require.NoError(t, err)

	kaCtx, kaCancel := context.WithCancel(context.Background())
	go func() {
		_, err2 := s.cli.KeepAlive(kaCtx, leaseResp.ID)
		require.NoError(t, err2)
		//for kaResp := range ch {
		//	t.Log(kaResp)
		//}
	}()
	server := grpc.NewServer()
	RegisterUserServiceServer(server, svc)
	server.Serve(l)
	kaCancel()
	err = em.DeleteEndpoint(ctx, key)
	if err != nil {
		t.Log(err)
	}
	// grpc优雅的退出
	server.GracefulStop()
	//s.cli.Close()
}

func TestBalancer(t *testing.T) {
	suite.Run(t, new(BalancerTestSuite))
}
