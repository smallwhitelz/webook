package events

import "context"

// Producer 使用kafka
type Producer interface {
	ProduceInconsistentEvent(ctx context.Context, evt InconsistentEvent) error
}
