package wechat

import (
	"context"
	"errors"
	"fmt"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/native"
	"time"
	"webook/payment/domain"
	"webook/payment/repository"
	"webook/pkg/logger"
)

var errUnknownTransactionState = errors.New("未知的微信事务状态")

type NativePaymentService struct {
	appID string
	mchID string
	// 支付通知 回调的URL
	notifyURL string
	// 自己的支付记录
	repo repository.PaymentRepository
	svc  *native.NativeApiService

	l logger.V1

	// 在微信 native 里面，分别是
	// SUCCESS：支付成功
	// REFUND：转入退款
	// NOTPAY：未支付
	// CLOSED：已关闭
	// REVOKED：已撤销（付款码支付）
	// USERPAYING：用户支付中（付款码支付）
	// PAYERROR：支付失败(其他原因，如银行返回失败)
	nativeCBTypeToStatus map[string]domain.PaymentStatus
}

func NewNativePaymentService(appID string, mchID string, repo repository.PaymentRepository, svc *native.NativeApiService, l logger.V1) *NativePaymentService {
	return &NativePaymentService{
		appID:     appID,
		mchID:     mchID,
		notifyURL: "http://wechat.meoying.com/pay/callback",
		repo:      repo, svc: svc, l: l,
		nativeCBTypeToStatus: map[string]domain.PaymentStatus{
			"SUCCESS":  domain.PaymentStatusSuccess,
			"PAYERROR": domain.PaymentStatusFailed,
			"NOTPAY":   domain.PaymentStatusInit,
			"CLOSED":   domain.PaymentStatusFailed,
			"REVOKED":  domain.PaymentStatusFailed,
			"REFUND":   domain.PaymentStatusRefund,
			// 其它状态你都可以加
		},
	}

}

func (n *NativePaymentService) Prepay(ctx context.Context, pmt domain.Payment) (string, error) {
	pmt.Status = domain.PaymentStatusInit
	err := n.repo.AddPayment(ctx, pmt)
	if err != nil {
		return "", err
	}
	resp, _, err := n.svc.Prepay(ctx, native.PrepayRequest{
		Appid:       core.String(n.appID),
		Mchid:       core.String(n.mchID),
		Description: core.String(pmt.Description),
		OutTradeNo:  core.String(pmt.BizTradeNO),
		// 最好这个带上
		TimeExpire: core.Time(time.Now().Add(time.Minute * 30)),
		Amount: &native.Amount{
			Total:    core.Int64(pmt.Amt.Total),
			Currency: core.String(pmt.Amt.Currency),
		},
	})
	if err != nil {
		return "", err
	}
	return *resp.CodeUrl, err
}

func (n *NativePaymentService) HandleCallback(ctx context.Context, txn *payments.Transaction) error {
	return n.updateByTxn(ctx, txn)
}

func (n *NativePaymentService) updateByTxn(ctx context.Context, txn *payments.Transaction) error {
	status, ok := n.nativeCBTypeToStatus[*txn.TradeState]
	if !ok {
		return fmt.Errorf("%w, 微信的状态是 %s", errUnknownTransactionState, *txn.TradeState)
	}
	// 很显然，就是更新一下我们本地数据里面payment的状态
	return n.repo.UpdatePayment(ctx, domain.Payment{
		// 微信过来的 Transaction Id
		TxnID:      *txn.TransactionId,
		BizTradeNO: *txn.OutTradeNo,
		Status:     status,
	})
}
