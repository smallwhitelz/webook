package simpleim

import (
	"context"
	"encoding/json"
	"github.com/IBM/sarama"
	"strconv"
)

type Service struct {
	producer sarama.SyncProducer
}

func (s *Service) Receive(ctx context.Context, sender int64, msg Message) error {
	// 1. 我要确定接收者是谁
	members := s.findMembers()
	for _, mem := range members {
		// 我自己的不用转发
		if mem == sender {
			continue
		}
		event := &Event{
			Msg:      msg,
			Receiver: mem,
		}
		val, _ := json.Marshal(event)
		_, _, err := s.producer.SendMessage(&sarama.ProducerMessage{
			Topic: eventName,
			Key:   sarama.StringEncoder(strconv.FormatInt(mem, 10)),
			Value: sarama.ByteEncoder(val),
		})
		if err != nil {
			// 记录日志
			continue
		}
	}
	return nil
}

func (s *Service) findMembers() []int64 {
	// 这里就是查询 IM 中的群组服务拿到成员
	// 模拟拿到的结果（模拟数据库查询的结果）
	return []int64{1, 2, 3, 4}
}
