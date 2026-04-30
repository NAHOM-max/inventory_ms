package kafka

import (
	"context"
	"time"

	"inventory_ms/internal/usecase"
)

// DLQEvent wraps the original event with failure metadata.
type DLQEvent struct {
	OriginalEvent usecase.DeliveryConfirmedEvent `json:"original_event"`
	Error         string                         `json:"error"`
	Service       string                         `json:"service"`
	RetryCount    int                            `json:"retry_count"`
	FailedAt      time.Time                      `json:"failed_at"`
}

// DLQProducer publishes failed events to the dead-letter queue.
type DLQProducer interface {
	Send(ctx context.Context, eventID string, event DLQEvent) error
	Close() error
}
