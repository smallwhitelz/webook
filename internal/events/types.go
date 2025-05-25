package events

// Consumer 给消费者定义一个统一的启动方法
type Consumer interface {
	Start() error
}
