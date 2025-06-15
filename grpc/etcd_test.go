package grpc

import (
	"context"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"testing"
	"time"
)

// EtcdTestSuite 使用Etcd作为注册中心实现服务发现
type EtcdTestSuite struct {
	suite.Suite
	cli *etcdv3.Client
}

func (s *EtcdTestSuite) SetupSuite() {
	// 初始化etcd客户端
	cli, err := etcdv3.New(etcdv3.Config{
		Endpoints: []string{"43.154.97.245:12379"},
		Username:  "root",
		Password:  "1234",
	})
	require.NoError(s.T(), err)
	s.cli = cli
}

func (s *EtcdTestSuite) TestClient() {
	t := s.T()
	etcdResolver, err := resolver.NewBuilder(s.cli)
	require.NoError(t, err)
	// 一定要注意是三斜杠，否则报错根本看不出来！！！！！！！！！
	cc, err := grpc.NewClient("etcd:///service/user",
		grpc.WithResolvers(etcdResolver),
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

func (s *EtcdTestSuite) TestServer() {
	t := s.T()
	em, err := endpoints.NewManager(s.cli, "service/user")
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	addr := "127.0.0.1:8090"
	key := "service/user/" + addr
	l, err := net.Listen("tcp", ":8090")
	require.NoError(t, err)

	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	// 租期 5秒
	var ttl int64 = 5
	// 进行租约
	leaseResp, err := s.cli.Grant(ctx, ttl)
	require.NoError(t, err)
	// 注册一个实例
	// 为什么在Serve前注册？
	// 因为Serve是一个阻塞的方法
	err = em.AddEndpoint(ctx, key, endpoints.Endpoint{
		// 定位信息，客户端如何连你
		Addr: addr,
		// 告诉etcd，加的这个节点是带有租约的
	}, etcdv3.WithLease(leaseResp.ID))
	require.NoError(t, err)

	kaCtx, kaCancel := context.WithCancel(context.Background())
	go func() {
		// 进行续约
		ch, err2 := s.cli.KeepAlive(kaCtx, leaseResp.ID)
		require.NoError(t, err2) // 如果续约失败，在这里就可以看出来
		for kaResp := range ch {
			t.Log(kaResp.String())
		}
	}()

	go func() {
		// 模拟注册信息变动
		ticker := time.NewTicker(time.Second)
		for now := range ticker.C {
			ctx1, cancel1 := context.WithTimeout(context.Background(), time.Second)
			// 这样写可以更新数据变动
			err1 := em.Update(ctx1, []*endpoints.UpdateWithOpts{
				{
					Update: endpoints.Update{
						Op:  endpoints.Add,
						Key: key,
						Endpoint: endpoints.Endpoint{
							Addr:     addr,
							Metadata: now.String(),
						},
					},
					// 切记，加了租约后，后面的改动不要忘掉租约，否则修改后就没有了租约
					Opts: []etcdv3.OpOption{etcdv3.WithLease(leaseResp.ID)},
				},
			})
			cancel1()
			if err1 != nil {
				t.Log(err1)
			}
			// 这样写也可以，因为底层调用的是同一个方法
			// 看其底层，AddEndpoint的语义更加接近 Insert or Update
			//err2 := em.AddEndpoint(ctx1, key, endpoints.Endpoint{
			//	Addr:     addr,
			//	Metadata: now.String(),
			//},etcdv3.WithLease(leaseResp.ID))
			//cancel1()
			//if err2 != nil {
			//	t.Log(err2)
			//}
		}
	}()
	server := grpc.NewServer()
	RegisterUserServiceServer(server, &Server{})
	server.Serve(l)

	// 退出服务，固然要退出续约
	kaCancel()
	// 退出服务删除掉注册的Endpoint
	err = em.DeleteEndpoint(ctx, key)
	if err != nil {
		t.Log(err)
	}
	// grpc优雅的退出
	server.GracefulStop()

	// 退出后，etcd的客户端也要优雅的关掉
	// 这里一定是要先删除endpoint，再关掉cli，顺序不能乱
	s.cli.Close()

	// etcdctl --endpoints=127.0.0.1:2379 --user=root --password=1234  get service/user --prefix 该命令可以看到存在etcd中的内容
}

func TestEtcd(t *testing.T) {
	suite.Run(t, new(EtcdTestSuite))
}
