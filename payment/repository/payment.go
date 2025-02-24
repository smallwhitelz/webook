package repository

import (
	"context"
	"time"
	"webook/payment/domain"
	"webook/payment/repository/dao"
)

type paymentRepository struct {
	dao dao.PaymentDAO
}

func NewPaymentRepository(dao dao.PaymentDAO) PaymentRepository {
	return &paymentRepository{dao: dao}
}

func (p *paymentRepository) AddPayment(ctx context.Context, pmt domain.Payment) error {
	return p.dao.Insert(ctx, p.toEntity(pmt))
}

func (p *paymentRepository) UpdatePayment(ctx context.Context, pmt domain.Payment) error {
	return p.dao.UpdateTxnIDAndStatus(ctx, pmt.BizTradeNO, pmt.TxnID, pmt.Status)
}

func (p *paymentRepository) FindExpiredPayment(ctx context.Context, offset int, limit int, t time.Time) ([]domain.Payment, error) {
	//TODO implement me
	panic("implement me")
}

func (p *paymentRepository) GetPayment(ctx context.Context, bizTradeNO string) (domain.Payment, error) {
	//TODO implement me
	panic("implement me")
}

func (p *paymentRepository) toEntity(pmt domain.Payment) dao.Payment {
	return dao.Payment{
		Amt:         pmt.Amt.Total,
		Currency:    pmt.Amt.Currency,
		BizTradeNO:  pmt.BizTradeNO,
		Description: pmt.Description,
		Status:      domain.PaymentStatusInit,
	}
}
