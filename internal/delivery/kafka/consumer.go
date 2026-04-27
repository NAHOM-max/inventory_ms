package kafka

import (
	"context"
	"encoding/json"
	"log"

	kafkago "github.com/segmentio/kafka-go"

	"inventory_ms/internal/usecase"
)

type DeliveryConfirmedConsumer struct {
	reader *kafkago.Reader
	uc     *usecase.HandleDeliveryConfirmedUseCase
}

func NewDeliveryConfirmedConsumer(brokers []string, uc *usecase.HandleDeliveryConfirmedUseCase) *DeliveryConfirmedConsumer {
	reader := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers: brokers,
		Topic:   "delivery.confirmed",
		GroupID: "inventory-service-group",
	})
	return &DeliveryConfirmedConsumer{reader: reader, uc: uc}
}

func (c *DeliveryConfirmedConsumer) Run(ctx context.Context) {
	defer c.reader.Close()
	for {
		msg, err := c.reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("kafka read error: %v", err)
			continue
		}

		eventID := string(msg.Key)
		log.Printf("event received event_id=%s topic=%s", eventID, msg.Topic)

		var event usecase.DeliveryConfirmedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("unmarshal error event_id=%s: %v", eventID, err)
			continue
		}

		if err := c.uc.Execute(ctx, eventID, event); err != nil {
			log.Printf("handle error event_id=%s shipment_id=%s order_id=%s: %v",
				eventID, event.ShipmentID, event.OrderID, err)
		}
	}
}
