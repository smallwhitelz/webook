package grpcx

import (
	"context"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"google.golang.org/grpc"
	"net"
	"strconv"
	"time"
	"webook/pkg/logger"
)

type Server struct {
	*grpc.Server
	EtcdAddr string
	Port     int
	Name     string
	L        logger.LoggerV1
	client   *etcdv3.Client
	kaCancel func()
}

// 依赖注入的形式
//func NewServer(c *etcdv3.Client) *Server {
//	return &Server{
//		client: c,
//	}
//}

func (s *Server) Serve() error {
	addr := ":" + strconv.Itoa(s.Port)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	err = s.register()
	if err != nil {
		return err
	}
	return s.Server.Serve(l)
}

func (s *Server) register() error {
	client, err := etcdv3.New(etcdv3.Config{
		Endpoints: []string{s.EtcdAddr},
		Username:  "root",
		Password:  "1234",
	})
	if err != nil {
		return err
	}
	s.client = client
	em, err := endpoints.NewManager(client, "service/"+s.Name)
	addr := "127.0.0.1:" + strconv.Itoa(s.Port)
	//addr := netx.GetOutboundIP() + ":" + strconv.Itoa(s.Port)
	key := "service/" + s.Name + "/" + addr
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	// 租期 5秒
	var ttl int64 = 5
	// 进行租约
	leaseResp, err := client.Grant(ctx, ttl)
	if err != nil {
		return err
	}
	// 注册一个实例
	// 为什么在Serve前注册？
	// 因为Serve是一个阻塞的方法
	err = em.AddEndpoint(ctx, key, endpoints.Endpoint{
		// 定位信息，客户端如何连你
		Addr: addr,
		// 告诉etcd，加的这个节点是带有租约的
	}, etcdv3.WithLease(leaseResp.ID))
	if err != nil {
		return err
	}

	kaCtx, kaCancel := context.WithCancel(context.Background())
	s.kaCancel = kaCancel
	// 进行续约
	ch, err := client.KeepAlive(kaCtx, leaseResp.ID)
	go func() {
		for kaResp := range ch {
			s.L.Debug(kaResp.String())
		}
	}()
	return err
}

func (s *Server) Close() error {
	if s.kaCancel != nil {
		s.kaCancel()
	}

	if s.client != nil {
		// 依赖注入的形式的话就不要关
		// 因为你的client可能被你用也可能被别人用了
		return s.client.Close()
	}
	s.GracefulStop()
	return nil
}
