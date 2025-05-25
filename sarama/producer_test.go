package sarama

import (
	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
	"testing"
)

var addr = []string{"43.154.97.245:9094"}

func TestSyncProducer(t *testing.T) {
	cfg := sarama.NewConfig()
	cfg.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer(addr, cfg)
	// 最典型的场景，利用Partitioner来保证同一个业务的消息一定发送在同一个分区，从而保证业务消息的有序性
	cfg.Producer.Partitioner = sarama.NewRoundRobinPartitioner // 轮训
	//cfg.Producer.Partitioner = sarama.NewRandomPartitioner // 随机
	//cfg.Producer.Partitioner = sarama.NewHashPartitioner // 根据Key进行Hash
	//cfg.Producer.Partitioner = sarama.NewManualPartitioner // 手动指定
	//cfg.Producer.Partitioner = sarama.NewConsistentCRCHashPartitioner // 一致性hash，很少用
	//cfg.Producer.Partitioner = sarama.NewCustomPartitioner() // 自定义hash
	assert.NoError(t, err)
	for i := 0; i < 100; i++ {
		_, _, err = producer.SendMessage(&sarama.ProducerMessage{
			Topic: "test_topic",
			Value: sarama.StringEncoder("这是一条消息"),
			// 会在生产者和消费者之间传递的
			Headers: []sarama.RecordHeader{
				{
					Key:   []byte("key1"),
					Value: []byte("value1"),
				},
			},
			Metadata: "这是 metadata ",
		})
	}
}

func TestAsyncProducer(t *testing.T) {
	cfg := sarama.NewConfig()
	cfg.Producer.Return.Successes = true
	cfg.Producer.Return.Errors = true
	producer, err := sarama.NewAsyncProducer(addr, cfg)
	assert.NoError(t, err)
	msgs := producer.Input()
	msgs <- &sarama.ProducerMessage{
		Topic: "test_topic",
		Value: sarama.StringEncoder("这是一条消息"),
		// 会在生产者和消费者之间传递的
		Headers: []sarama.RecordHeader{
			{
				Key:   []byte("key1"),
				Value: []byte("value1"),
			},
		},
		Metadata: "这是 metadata ",
	}
	select {
	case msg := <-producer.Successes():
		t.Log("发送成功", string(msg.Value.(sarama.StringEncoder)))
	case err := <-producer.Errors():
		t.Log("发送失败", err.Err, err.Msg)
	}
}
