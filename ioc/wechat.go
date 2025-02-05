package ioc

import (
	"os"
	"webook/internal/service/oauth2/wechat"
	"webook/pkg/logger"
)

func InitWechatService(l logger.V1) wechat.Service {
	err := os.Setenv("WECHAT_APP_ID", "wx7256bc69ab349c72")
	if err != nil {
		panic("设置环境变量失败")
	}
	appID, ok := os.LookupEnv("WECHAT_APP_ID")
	if !ok {
		panic("找不到环境变量 WECHAT_APP_ID")
	}
	err = os.Setenv("WECHAT_APP_SECRET", "wx7256bc69ab349c72")
	if err != nil {
		panic("设置环境变量失败")
	}
	appSecret, ok := os.LookupEnv("WECHAT_APP_SECRET")
	if !ok {
		panic("找不到环境变量 WECHAT_APP_SECRET")
	}
	return wechat.NewService(appID, appSecret, l)
}
