package logger

// Logger 这种风格一般需要用户提前在日志处留好占位符
type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

func example() {
	var l Logger
	l.Info("用户的微信 id %d", 111)
}

// LoggerV1 结构化打日志
type LoggerV1 interface {
	Debug(msg string, args ...Field)
	Info(msg string, args ...Field)
	Warn(msg string, args ...Field)
	Error(msg string, args ...Field)
}

type Field struct {
	Key string
	Val any
}

func exampleV1() {
	var l LoggerV1
	l.Info("这是一个新用户", Field{Key: "union_id", Val: 123})
}

// 不建议
//type LoggerV2 interface {
//	// 它要求args必须是偶数，并且是以 key1,value1,key2,value2的形式传递
//	Debug(msg string, args ...any)
//	Info(msg string, args ...any)
//	Warn(msg string, args ...any)
//	Error(msg string, args ...any)
//}
