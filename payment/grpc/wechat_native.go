package grpc

import (
	"context"
	"google.golang.org/grpc"
	"webook/api/proto/gen/payment/v1"
	"webook/payment/domain"
	"webook/payment/services/wechat"
)

type WechatServiceServer struct {
	pmtv1.UnimplementedWechatPaymentServiceServer
	svc *wechat.NativePaymentService
}

func NewWechatServiceServer(svc *wechat.NativePaymentService) *WechatServiceServer {
	return &WechatServiceServer{svc: svc}
}
func (w *WechatServiceServer) Register(server *grpc.Server) {
	pmtv1.RegisterWechatPaymentServiceServer(server, w)
}

func (w *WechatServiceServer) NativePrePay(ctx context.Context, req *pmtv1.PrePayRequest) (*pmtv1.NativePrePayResponse, error) {
	codeURL, err := w.svc.Prepay(ctx, domain.Payment{
		Amt: domain.Amount{
			Currency: req.Amt.Currency,
			Total:    req.Amt.Total,
		},
		BizTradeNO:  req.BizTradeNo,
		Description: req.Description,
	})
	if err != nil {
		return nil, err
	}
	return &pmtv1.NativePrePayResponse{
		CodeUrl: codeURL,
	}, nil
}

func (w *WechatServiceServer) GetPayment(ctx context.Context, req *pmtv1.GetPaymentRequest) (*pmtv1.GetPaymentResponse, error) {
	p, err := w.svc.GetPayment(ctx, req.GetBizTradeNo())
	if err != nil {
		return nil, err
	}
	return &pmtv1.GetPaymentResponse{
		Status: pmtv1.PaymentStatus(p.Status),
	}, nil
}
