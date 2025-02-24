package dao

import (
	"context"
	"gorm.io/gorm"
	"time"
	"webook/payment/domain"
)

type PaymentGORMDAO struct {
	db *gorm.DB
}

func NewPaymentGORMDAO(db *gorm.DB) PaymentDAO {
	return &PaymentGORMDAO{db: db}
}

func (p *PaymentGORMDAO) Insert(ctx context.Context, pmt Payment) error {
	now := time.Now().UnixMilli()
	pmt.Ctime = now
	pmt.Utime = now
	return p.db.WithContext(ctx).Create(&pmt).Error
}

func (p *PaymentGORMDAO) UpdateTxnIDAndStatus(ctx context.Context, bizTradeNo string, txnID string, status domain.PaymentStatus) error {
	return p.db.WithContext(ctx).Model(&Payment{}).
		Where("biz_trade_no = ?", bizTradeNo).
		Updates(map[string]any{
			"txn_id": txnID,
			"status": status.AsUint8(),
			"utime":  time.Now().UnixMilli(),
		}).Error
}

func (p *PaymentGORMDAO) FindExpiredPayment(ctx context.Context, offset int, limit int, t time.Time) ([]Payment, error) {
	//TODO implement me
	panic("implement me")
}

func (p *PaymentGORMDAO) GetPayment(ctx context.Context, bizTradeNO string) (Payment, error) {
	//TODO implement me
	panic("implement me")
}
