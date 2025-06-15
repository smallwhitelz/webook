package grpc

import (
	"context"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"encoding/json"
	consul "github.com/hashicorp/consul/api"
	"google.golang.org/grpc/resolver"
	"os"
)

const retryInterval = 1 * time.Second

// 源自 https://gitee.com/xiao_hange/go-admin-pkg/blob/master/pkg/grpcx/resolver/consul/consul_resolver.go
type ConsulResolverConfig struct {
	CacheFile string
}

// 需要实现 Resolver Builder 接口
type consulResolverBuilder struct {
	client *consul.Client
	ccfg   *ConsulResolverConfig
}

// Build 创建一个新的 Consul Resolver。
func (b *consulResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	serviceName := target.URL.Path
	if serviceName == "" {
		return nil, status.Errorf(codes.InvalidArgument, "resolver: missing service name in target")
	}
	serviceName = strings.TrimPrefix(serviceName, "/")
	r := &consulResolver{
		target:      target,
		client:      b.client,
		serviceName: serviceName,
		cc:          cc,
		ctx:         context.Background(),
		cancel:      func() {},
		instances:   make(map[string]struct{}),
		cacheFile:   fmt.Sprintf("%s/%s_instances.json", b.ccfg.CacheFile, serviceName),
	}
	r.ctx, r.cancel = context.WithCancel(r.ctx)
	go r.watch()
	return r, nil
}

// Scheme 返回 Consul Resolver 的方案。
func (b *consulResolverBuilder) Scheme() string {
	return "consul"
}

// NewConsuleResolverBuilder 为 Consul 创建一个新的解析器构建器。
func NewConsuleResolverBuilder(client *consul.Client, ccfg *ConsulResolverConfig) (resolver.Builder, error) {
	return &consulResolverBuilder{
		client: client,
		ccfg:   ccfg,
	}, nil
}

type consulResolver struct {
	client      *consul.Client
	serviceName string
	cc          resolver.ClientConn
	target      resolver.Target
	ctx         context.Context
	cancel      context.CancelFunc
	instances   map[string]struct{}
	mu          sync.Mutex
	cacheFile   string
}

// watch 持续查询 Consul 获取服务实例并更新解析器状态。
func (r *consulResolver) watch() {

	queryOptions := &consul.QueryOptions{
		WaitIndex: 0,
	}

	for {
		select {
		case <-r.ctx.Done():
			return
		default:
		}

		instances, meta, err := r.getInstances(queryOptions)
		if err != nil {
			cachedInstances, err := r.loadFromLocalCache()
			if err == nil {
				fmt.Println("Consul服务出现问题,启动本地缓存配置.")
				r.updateInstances(cachedInstances)
			}
			time.Sleep(retryInterval)
			continue
		}

		r.updateInstances(instances)
		queryOptions.WaitIndex = meta.LastIndex
	}
}

// getInstances 查询 Consul 获取服务实例。
func (r *consulResolver) getInstances(queryOptions *consul.QueryOptions) ([]*consul.ServiceEntry, *consul.QueryMeta, error) {
	entries, meta, err := r.client.Health().Service(r.serviceName, "", true, queryOptions)
	if err != nil {
		return nil, nil, err
	}
	if len(entries) > 0 {
		if err = r.writeToLocalCache(entries); err != nil {
			fmt.Printf("写入本地缓存时出错: %v\n", err)
		}
	}
	return entries, meta, nil
}

// updateInstances 处理获取的服务实例，更新解析器状态，并移除不再可用的实例。
func (r *consulResolver) updateInstances(instances []*consul.ServiceEntry) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var updatedInstances []resolver.Address
	for _, entry := range instances {
		address := entry.Service.Address
		if address == "" {
			address = entry.Node.Address
		}
		port := entry.Service.Port

		instanceAddr := fmt.Sprintf("%s:%d", address, port)

		r.instances[instanceAddr] = struct{}{}

		updatedInstances = append(updatedInstances, resolver.Address{
			Addr:     instanceAddr,
			Metadata: entry.Service.Meta,
		})
	}

	r.cc.UpdateState(resolver.State{Addresses: updatedInstances})

	for addr := range r.instances {
		found := false
		for _, entry := range instances {
			address := entry.Service.Address
			if address == "" {
				address = entry.Node.Address
			}
			port := entry.Service.Port

			instanceAddr := fmt.Sprintf("%s:%d", address, port)
			if addr == instanceAddr {
				found = true
				break
			}
		}
		if !found {
			delete(r.instances, addr)
		}
	}
}

// writeToLocalCache 将实例写入本地缓存文件
func (r *consulResolver) writeToLocalCache(entries []*consul.ServiceEntry) error {
	// 检查文件是否存在
	if _, err := os.Stat(r.cacheFile); os.IsNotExist(err) {
		// 文件不存在，创建目录
		if err = os.MkdirAll(filepath.Dir(r.cacheFile), 0755); err != nil {
			return err
		}
	}

	// 打开文件，如果不存在则创建
	file, err := os.OpenFile(r.cacheFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// 使用 JSON 编码并写入文件
	encoder := json.NewEncoder(file)
	if err = encoder.Encode(entries); err != nil {
		return err
	}
	return nil
}

// loadFromLocalCache 从本地缓存加载实例。
func (r *consulResolver) loadFromLocalCache() ([]*consul.ServiceEntry, error) {
	// 打开文件
	file, err := os.Open(r.cacheFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// 使用 JSON 解码并加载节点
	decoder := json.NewDecoder(file)
	var instances []*consul.ServiceEntry
	if err = decoder.Decode(&instances); err != nil {
		return nil, err
	}
	return instances, nil
}

// ResolveNow 这里是一个空操作。
// 这只是一个提示，如果不需要可以忽略
func (r *consulResolver) ResolveNow(resolver.ResolveNowOptions) {}

// Close 取消上下文并等待 watch 协程结束。
func (r *consulResolver) Close() {
	r.cancel()
}
