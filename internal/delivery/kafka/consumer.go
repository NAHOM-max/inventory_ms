package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	kafkago "github.com/segmentio/kafka-go"

	"inventory_ms/internal/domain"
	"inventory_ms/internal/usecase"
)

const serviceName = "inventory-service"

type ConsumerConfig struct {
	Brokers    []string
	MaxRetries int
}

type DeliveryConfirmedConsumer struct {
	reader     *kafkago.Reader
	uc         *usecase.HandleDeliveryConfirmedUseCase
	dlq        DLQProducer
	maxRetries int
}

func NewDeliveryConfirmedConsumer(cfg ConsumerConfig, uc *usecase.HandleDeliveryConfirmedUseCase, dlq DLQProducer) *DeliveryConfirmedConsumer {
	reader := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers: cfg.Brokers,
		Topic:   "delivery.confirmed",
		GroupID: "inventory-service-group",
	})
	return &DeliveryConfirmedConsumer{
		reader:     reader,
		uc:         uc,
		dlq:        dlq,
		maxRetries: cfg.MaxRetries,
	}
}

func (c *DeliveryConfirmedConsumer) Run(ctx context.Context) {
	defer c.reader.Close()
	defer c.dlq.Close()

	for {
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("kafka fetch error: %v", err)
			continue
		}

		c.handle(ctx, msg)

		// Always commit — offset advances regardless of processing outcome.
		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			log.Printf("kafka commit error event_id=%s: %v", string(msg.Key), err)
		}
	}
}

func (c *DeliveryConfirmedConsumer) handle(ctx context.Context, msg kafkago.Message) {
	eventID := string(msg.Key)
	log.Printf("event received event_id=%s topic=%s", eventID, msg.Topic)

	var event usecase.DeliveryConfirmedEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		log.Printf("unmarshal error event_id=%s: %v — sending to DLQ", eventID, err)
		c.sendToDLQ(ctx, eventID, event, err, 0)
		return
	}

	lastErr := c.withRetry(ctx, eventID, event)
	if lastErr != nil {
		c.sendToDLQ(ctx, eventID, event, lastErr, c.maxRetries)
	}
}

// withRetry executes the use case up to maxRetries times for retryable errors.
// Returns nil on success, the last error otherwise.
func (c *DeliveryConfirmedConsumer) withRetry(ctx context.Context, eventID string, event usecase.DeliveryConfirmedEvent) error {
	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			log.Printf("retry attempt %d event_id=%s shipment_id=%s order_id=%s",
				attempt, eventID, event.ShipmentID, event.OrderID)
			time.Sleep(backoff(attempt))
		}

		lastErr = c.uc.Execute(ctx, eventID, event)
		if lastErr == nil {
			log.Printf("event processed event_id=%s shipment_id=%s order_id=%s",
				eventID, event.ShipmentID, event.OrderID)
			return nil
		}

		if errors.Is(lastErr, domain.ErrNonRetryable) {
			log.Printf("non-retryable error event_id=%s: %v — skipping retries", eventID, lastErr)
			return lastErr
		}
	}
	return lastErr
}

func (c *DeliveryConfirmedConsumer) sendToDLQ(ctx context.Context, eventID string, event usecase.DeliveryConfirmedEvent, cause error, retryCount int) {
	log.Printf("sending event to DLQ event_id=%s shipment_id=%s order_id=%s service=%s retries=%d reason=%v",
		eventID, event.ShipmentID, event.OrderID, serviceName, retryCount, cause)

	dlqEvent := DLQEvent{
		OriginalEvent: event,
		Error:         cause.Error(),
		Service:       serviceName,
		RetryCount:    retryCount,
		FailedAt:      time.Now().UTC(),
	}

	if err := c.dlq.Send(ctx, eventID, dlqEvent); err != nil {
		log.Printf("DLQ publish failed event_id=%s: %v", eventID, err)
	}
}

// backoff returns a simple exponential delay capped at 30s.
func backoff(attempt int) time.Duration {
	d := time.Duration(attempt) * 2 * time.Second
	if d > 30*time.Second {
		return 30 * time.Second
	}
	return d
}
