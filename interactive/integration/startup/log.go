package startup

import "webook/pkg/logger"

func InitLogger() logger.V1 {
	return logger.NewNopLogger()
}
