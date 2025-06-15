package grpc

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	consul "github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"testing"
	"time"
)

// ConsulRegistryTestSuite 使用Consul来作为注册中心
type ConsulRegistryTestSuite struct {
	suite.Suite
	client *consul.Client
}

func (s *ConsulRegistryTestSuite) SetupSuite() {
	client, err := consul.NewClient(&consul.Config{
		Address: "localhost:8500",
	})
	require.NoError(s.T(), err)
	s.client = client
}

// 启动服务端
func (s *ConsulRegistryTestSuite) TestServer() {
	l, err := net.Listen("tcp", ":8090")
	require.NoError(s.T(), err)
	err = s.consulRegister()
	require.NoError(s.T(), err)
	server := grpc.NewServer()
	RegisterUserServiceServer(server, &Server{})
	server.Serve(l)
}

func (s *ConsulRegistryTestSuite) consulRegister() error {
	//生成注册对象
	registration := &consul.AgentServiceRegistration{
		Name:    "user-service",
		ID:      uuid.New().String(),
		Port:    8090,
		Address: "127.0.0.1",
	}
	err := s.client.Agent().ServiceRegister(registration)
	return err
}

// 启动服务端
func (s *ConsulRegistryTestSuite) TestClient() {
	cd, err := NewConsuleResolverBuilder(s.client, &ConsulResolverConfig{
		CacheFile: "/tmp/cache/consul",
	})
	require.NoError(s.T(), err)
	cc, err := grpc.Dial(fmt.Sprintf("%s:///%s", "consul", "user-service"),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithResolvers(cd))
	require.NoError(s.T(), err)
	uc := NewUserServiceClient(cc)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	resp, err := uc.GetByID(ctx, &GetByIDRequest{
		Id: 123,
	})
	require.NoError(s.T(), err)
	s.T().Log(resp)
}

func TestConsul(t *testing.T) {
	suite.Run(t, new(ConsulRegistryTestSuite))
}
