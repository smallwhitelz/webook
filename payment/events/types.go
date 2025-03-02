package events

// PaymentEvent 是最简设计
// 有一些人喜欢把支付详情放进来，单目前看来是没有必要的
// 后续如果接入大数据之类的，那么就可以考虑提供payment详情了
type PaymentEvent struct {
	BizTradeNo string
	Status     uint8
}

func (PaymentEvent) Topic() string {
	return "payment_events"
}
