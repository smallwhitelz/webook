package logger

func Error(err error) Field {
	return Field{Key: "error", Val: err}
}

func Int64(key string, val int64) Field {
	return Field{Key: key, Val: val}
}

func Int32(key string, val int32) Field {
	return Field{Key: key, Val: val}
}

func Int(key string, val int) Field {
	return Field{Key: key, Val: val}
}

func String(key string, val string) Field {
	return Field{Key: key, Val: val}
}
