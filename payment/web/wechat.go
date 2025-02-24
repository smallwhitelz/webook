package web

import (
	"github.com/gin-gonic/gin"
	"github.com/wechatpay-apiv3/wechatpay-go/core/notify"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments"
	"net/http"
	"webook/payment/services/wechat"
	"webook/pkg/logger"
)

type WechatHandler struct {
	handler   *notify.Handler
	l         logger.V1
	nativeSvc *wechat.NativePaymentService
}

func NewWechatHandler(handler *notify.Handler,
	l logger.V1, nativeSvc *wechat.NativePaymentService) *WechatHandler {
	return &WechatHandler{handler: handler, l: l, nativeSvc: nativeSvc}
}

func (h *WechatHandler) RegisterRoutes(server *gin.Engine) {
	server.GET("/hello", func(context *gin.Context) {
		context.String(http.StatusOK, "我进来了")
	})
	server.Any("/pay/callback", h.HandleNative)
}

func (h *WechatHandler) HandleNative(ctx *gin.Context) {
	// 用来接收解密后的数据的
	transaction := new(payments.Transaction)
	_, err := h.handler.ParseNotifyRequest(ctx, ctx.Request, transaction)
	if err != nil {
		ctx.String(http.StatusBadRequest, "参数解析失败")
		h.l.Error("解析微信支付回调失败", logger.Error(err))
		// 在这里，可以考虑进一步加监控和告警
		// 进来这里，绝大多数是黑客攻击
		return
	}
	err = h.nativeSvc.HandleCallback(ctx, transaction)
	if err != nil {
		ctx.String(http.StatusInternalServerError, "系统异常")
		// 说明处理回调失败了
		h.l.Error("处理微信支付回调失败", logger.Error(err),
			logger.String("biz_trade_no", *transaction.OutTradeNo))
		return
	}
	ctx.String(http.StatusOK, "OK")
}
