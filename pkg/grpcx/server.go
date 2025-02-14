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
	Username string
	Password string
	client   *etcdv3.Client
	kaCancel func()
	L        logger.V1
}

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
	s.register()
	return s.Server.Serve(l)
}

func (s *Server) register() error {
	client, err := etcdv3.New(etcdv3.Config{
		Endpoints: []string{s.EtcdAddr},
		Username:  s.Username,
		Password:  s.Password,
	})
	if err != nil {
		return err
	}
	s.client = client
	em, err := endpoints.NewManager(client, "service/"+s.Name)
	addr := "127.0.0.1:" + strconv.Itoa(s.Port)
	key := "service/" + s.Name + "/" + addr
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	// 租期
	var ttl int64 = 5
	leaseResp, err := client.Grant(ctx, ttl)
	if err != nil {
		return err
	}
	err = em.AddEndpoint(ctx, key, endpoints.Endpoint{
		// 定位信息，客户端如何连你
		Addr: addr,
	}, etcdv3.WithLease(leaseResp.ID))
	if err != nil {
		return err
	}
	kaCtx, kaCancel := context.WithCancel(context.Background())
	s.kaCancel = kaCancel
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
		// 依赖注入的写法话就不要关
		return s.client.Close()
	}
	s.GracefulStop()
	return nil
}
