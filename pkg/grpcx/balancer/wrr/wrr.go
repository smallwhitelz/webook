package wrr

import (
	"context"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"math"
	"sync"
)

const Name = "custom_weighted_round_robin"

func newBuilder() balancer.Builder {
	return base.NewBalancerBuilder(Name, &PickerBuilder{}, base.Config{HealthCheck: true})
}

func init() {
	balancer.Register(newBuilder())
}

type PickerBuilder struct {
}

func (p *PickerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	conns := make([]*weightConn, 0, len(info.ReadySCs))
	for sc, sci := range info.ReadySCs {
		md, _ := sci.Address.Metadata.(map[string]any)
		weightVal, _ := md["weight"]
		weight, _ := weightVal.(float64)
		conns = append(conns, &weightConn{
			SubConn:       sc,
			weight:        int(weight),
			currentWeight: int(weight),
		})
	}
	return &Picker{
		conns: conns,
	}
}

type Picker struct {
	conns []*weightConn
	lock  sync.Mutex
}

func (p *Picker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	p.lock.Lock()
	defer p.lock.Unlock()
	if len(p.conns) == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}
	var total int
	var maxCC *weightConn
	for _, c := range p.conns {
		total += c.weight
		c.currentWeight = c.currentWeight + c.weight
		if maxCC == nil || maxCC.currentWeight < c.currentWeight {
			maxCC = c
		}
	}
	maxCC.currentWeight = maxCC.currentWeight - total
	return balancer.PickResult{
		SubConn: maxCC.SubConn,
		Done: func(info balancer.DoneInfo) {
			// 要在这里进一步调整weight/currentWeight
			// failover要在这里做文章
			// 根据调用结果的具体错误信息进行容错
			// 1. 如果要是出发了限流
			// 1.1 你可以考虑直接挪走这个节点，后面再挪回来
			// 1.2 你可以考虑直接将weight/currentWeight调整到极低
			// 2. 触发了熔断呢？
			// 3. 降级呢？
		},
	}, nil
}

// PickV1 动态调整权重
func (p *Picker) PickV1(info balancer.PickInfo) (balancer.PickResult, error) {
	p.lock.Lock()
	defer p.lock.Unlock()
	if len(p.conns) == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}
	var totalWeight int
	var maxCC *weightConn
	for _, c := range p.conns {
		c.mutex.Lock()
		totalWeight = totalWeight + c.efficientWeight
		c.currentWeight = c.currentWeight + c.efficientWeight
		if maxCC == nil || maxCC.currentWeight < c.currentWeight {
			maxCC = c
		}
		c.mutex.Unlock()
	}
	maxCC.mutex.Lock()
	maxCC.currentWeight = maxCC.currentWeight - totalWeight
	maxCC.mutex.Unlock()
	return balancer.PickResult{
		SubConn: maxCC.SubConn,
		Done: func(info balancer.DoneInfo) {
			maxCC.mutex.Lock()
			defer maxCC.mutex.Unlock()
			if info.Err != nil && maxCC.efficientWeight == 0 {
				return
			}
			// MaxUint32 可以替换为你认为的最大值。
			// 例如说你预期节点的权重是在 100 - 200 之间
			// 那么你可以设置经过动态调整之后的权重不会超过 500。
			if info.Err == nil && maxCC.efficientWeight >= 500 {
				return
			}
			if info.Err != nil {
				maxCC.efficientWeight--
			} else {
				maxCC.efficientWeight++
			}
			// 要在这里进一步调整weight/currentWeight
			// failover要在这里做文章
			// 根据调用结果的具体错误信息进行容错
			// 1. 如果要是出发了限流
			// 1.1 你可以考虑直接挪走这个节点，后面再挪回来
			// 1.2 你可以考虑直接将weight/currentWeight调整到极低
			// 2. 触发了熔断呢？
			// 3. 降级呢？
		},
	}, nil
}

// PickV2 客户端容错
func (p *Picker) PickV2(info balancer.PickInfo) (balancer.PickResult, error) {
	p.lock.Lock()
	defer p.lock.Unlock()
	if len(p.conns) == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}
	var totalWeight int
	var maxCC *weightConn
	for _, c := range p.conns {
		c.mutex.Lock()
		totalWeight = totalWeight + c.efficientWeight
		c.currentWeight = c.currentWeight + c.efficientWeight
		if maxCC == nil || maxCC.currentWeight < c.currentWeight {
			maxCC = c
		}
		c.mutex.Unlock()
	}
	maxCC.mutex.Lock()
	maxCC.currentWeight = maxCC.currentWeight - totalWeight
	maxCC.mutex.Unlock()
	return balancer.PickResult{
		SubConn: maxCC.SubConn,
		Done: func(info balancer.DoneInfo) {
			maxCC.mutex.Lock()
			defer maxCC.mutex.Unlock()
			if info.Err != nil && maxCC.efficientWeight == 0 {
				return
			}
			// MaxUint32 可以替换为你认为的最大值。
			// 例如说你预期节点的权重是在 100 - 200 之间
			// 那么你可以设置经过动态调整之后的权重不会超过 500。
			switch info.Err {
			case nil:
				if maxCC.efficientWeight == math.MaxInt32 {
					return
				}
				// 增加权重
				maxCC.efficientWeight++
			case context.DeadlineExceeded:
				// 超时可以考虑动态调整
				// 比如第一次超时降低1，第二次连续超时降低2
				maxCC.efficientWeight = maxCC.efficientWeight - 10
			default:
				// 检测服务端的错误
				code := status.Code(info.Err)
				switch code {
				// 假如服务端返回这个代表熔断
				case codes.Unavailable:
					// 直接降低到 1，我们可以预期接下来几乎不会选中它。
					// 但是本身没有降低到 0，所以它又存在被选中的机会，
					// 那么后续会慢慢恢复过来
					maxCC.efficientWeight = 1
				case codes.ResourceExhausted:
					// 直接减半，可以快速降低选中该节点的概率
					maxCC.efficientWeight = maxCC.efficientWeight / 2
				// 假定我们服务端返回这个代表降级
				case codes.Aborted:
					// 降级可以考虑和限流采用类似的策略，你也可以调整减少的幅度
					maxCC.efficientWeight = maxCC.efficientWeight / 2
				default:
					if maxCC.efficientWeight == 1 {
						// 降无可降
						return
					}
					maxCC.efficientWeight--
				}
			}
		},
	}, nil
}

type weightConn struct {
	balancer.SubConn
	mutex           sync.Mutex
	weight          int
	currentWeight   int
	efficientWeight int
}
