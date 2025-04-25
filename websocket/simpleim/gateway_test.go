package simpleim

import (
	"github.com/IBM/sarama"
	"github.com/ecodeclub/ekit/syncx"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
)

type GatewayTestSuite struct {
	suite.Suite
	client sarama.Client
}

func (g *GatewayTestSuite) SetupSuite() {
	cfg := sarama.NewConfig()
	cfg.Producer.Return.Successes = true
	cfg.Producer.Return.Errors = true
	client, err := sarama.NewClient([]string{"localhost:9094"}, cfg)
	g.client = client
	require.NoError(g.T(), err)
}

func (g *GatewayTestSuite) TestGateway() {
	// 启动三个实例，分别监听端口 8081,8082 和 8083，模拟分布式环境
	go func() {
		err := g.startGateway("gateway_8081", ":8081")
		g.T().Log("8081 退出服务", err)
	}()

	go func() {
		err := g.startGateway("gateway_8082", ":8082")
		g.T().Log("8082 退出服务", err)
	}()

	err := g.startGateway("gateway_8083", ":8083")
	g.T().Log("8083 退出服务", err)
}

func (g *GatewayTestSuite) startGateway(instance, addr string) error {
	// 启动一个 gateway 的实例
	producer, err := sarama.NewSyncProducerFromClient(g.client)
	require.NoError(g.T(), err)
	gateway := &WsGateway{
		conns: &syncx.Map[int64, *Conn]{},
		svc: &Service{
			producer: producer,
		},
		client:     g.client,
		instanceID: instance,
	}
	return gateway.Start(addr)
}

func TestWsGateway(t *testing.T) {
	suite.Run(t, new(GatewayTestSuite))
}
