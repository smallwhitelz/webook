package auth

import (
	"context"
	"github.com/golang-jwt/jwt/v5"
	"webook/internal/service/sms"
)

type SMSService struct {
	svc sms.Service
	key []byte
}

// Send 提高安全性，短信服务是要收费的，一般公司一个业务组会去申请短信额度，总不能谁都能用，所以这里就需要安全性
// 这里采用认token，不认人的方式，
func (s *SMSService) Send(ctx context.Context, tplToken string, args []string, numbers ...string) error {
	var claims SMSClaims
	_, err := jwt.ParseWithClaims(tplToken, &claims, func(token *jwt.Token) (interface{}, error) {
		return s.key, nil
	})
	if err != nil {
		return err
	}
	return s.svc.Send(ctx, claims.Tpl, args, numbers...)
}

// SMSClaims 从token解析出来tpl模版去发送
type SMSClaims struct {
	jwt.RegisteredClaims
	Tpl string
	// 可以额外加字段
}
