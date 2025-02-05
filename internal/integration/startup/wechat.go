package startup

import (
	"webook/internal/service/oauth2/wechat"
	"webook/pkg/logger"
)

func InitWechatService(l logger.V1) wechat.Service {
	return wechat.NewService("", "", l)
}
