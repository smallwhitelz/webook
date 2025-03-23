package events

import (
	"github.com/IBM/sarama"
	"webook/pkg/logger"
	"webook/search/service"
)

type SyncDataEvent struct {
	IndexName string
	DocID     string
	// 这里应该是 BizTags
	Data string
}
type SyncDataEventConsumer struct {
	svc    service.SyncService
	client sarama.Client
	l      logger.V1
}
