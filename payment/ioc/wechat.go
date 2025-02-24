package ioc

import (
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/native"
	"webook/payment/repository"
	"webook/payment/services/wechat"
	"webook/pkg/logger"
)

func InitWechatNativeService(cli *core.Client,
	repo repository.PaymentRepository, l logger.V1) *wechat.NativePaymentService {
	return wechat.NewNativePaymentService("11", "11",
		repo, &native.NativeApiService{
			Client: cli,
		}, l)
}
