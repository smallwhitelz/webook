package localsms

import (
	"context"
	"log"
)

type Service struct {
}

func NewService() *Service {
	return &Service{}
}

// Send 本地发送验证码模拟手机号登陆，真正的短信验证码发送是要收费的
func (s *Service) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	log.Println("验证码是", args)
	return nil
}
