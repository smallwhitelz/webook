package events

import (
	"context"
	"encoding/json"
	"github.com/IBM/sarama"
)

type SaramaProducer struct {
	p sarama.SyncProducer
}

func (s *SaramaProducer) ProducePaymentEvent(ctx context.Context, evt PaymentEvent) error {
	val, err := json.Marshal(evt)
	if err != nil {
		return err
	}
	_, _, err = s.p.SendMessage(&sarama.ProducerMessage{
		Key:   sarama.StringEncoder(evt.BizTradeNo),
		Topic: evt.Topic(),
		Value: sarama.ByteEncoder(val),
	})
	return err
}
