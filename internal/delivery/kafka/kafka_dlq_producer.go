package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	kafkago "github.com/segmentio/kafka-go"
)

type KafkaDLQProducer struct {
	writer *kafkago.Writer
}

func NewKafkaDLQProducer(brokers []string) *KafkaDLQProducer {
	return &KafkaDLQProducer{
		writer: &kafkago.Writer{
			Addr:         kafkago.TCP(brokers...),
			Topic:        "delivery.confirmed.dlq",
			Balancer:     &kafkago.LeastBytes{},
			WriteTimeout: 10 * time.Second,
		},
	}
}

func (p *KafkaDLQProducer) Send(ctx context.Context, eventID string, event DLQEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal dlq event: %w", err)
	}

	err = p.writer.WriteMessages(ctx, kafkago.Message{
		Key:   []byte(eventID),
		Value: payload,
	})
	if err != nil {
		log.Printf("DLQ publish failed event_id=%s shipment_id=%s order_id=%s: %v",
			eventID, event.OriginalEvent.ShipmentID, event.OriginalEvent.OrderID, err)
		return fmt.Errorf("dlq publish: %w", err)
	}

	log.Printf("DLQ publish success event_id=%s shipment_id=%s order_id=%s retries=%d",
		eventID, event.OriginalEvent.ShipmentID, event.OriginalEvent.OrderID, event.RetryCount)
	return nil
}

func (p *KafkaDLQProducer) Close() error {
	return p.writer.Close()
}
