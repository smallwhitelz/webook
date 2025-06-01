package job

// Job 为了便于控制，方便扩展，我们使用自己的接口
// 在这个基础上，可以考虑引入监控、告警、重试等（都是装饰器模式）
type Job interface {
	Name() string
	Run() error
}
