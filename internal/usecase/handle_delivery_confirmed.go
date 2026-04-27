package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"inventory_ms/internal/domain"
)

type DeliveryConfirmedEvent struct {
	ShipmentID     string    `json:"shipment_id"`
	OrderID        string    `json:"order_id"`
	TrackingNumber string    `json:"tracking_number"`
	DeliveredAt    time.Time `json:"delivered_at"`
}

type HandleDeliveryConfirmedUseCase struct {
	inbox      domain.InboxRepository
	transactor domain.Transactor
}

func NewHandleDeliveryConfirmedUseCase(inbox domain.InboxRepository, tx domain.Transactor) *HandleDeliveryConfirmedUseCase {
	return &HandleDeliveryConfirmedUseCase{inbox: inbox, transactor: tx}
}

func (uc *HandleDeliveryConfirmedUseCase) Execute(ctx context.Context, eventID string, event DeliveryConfirmedEvent) error {
	processed, err := uc.inbox.Exists(ctx, eventID)
	if err != nil {
		return fmt.Errorf("check inbox: %w", err)
	}
	if processed {
		log.Printf("duplicate skipped event_id=%s shipment_id=%s order_id=%s", eventID, event.ShipmentID, event.OrderID)
		return nil
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	return uc.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		if err := uc.inbox.Save(ctx, domain.InboxEvent{
			EventID:   eventID,
			EventType: "delivery.confirmed",
			Payload:   payload,
		}); err != nil {
			return fmt.Errorf("save inbox event: %w", err)
		}

		log.Printf("Inventory service received delivery.confirmed for shipment %s event_id=%s order_id=%s",
			event.ShipmentID, eventID, event.OrderID)

		if err := uc.inbox.MarkProcessed(ctx, eventID); err != nil {
			return fmt.Errorf("mark processed: %w", err)
		}
		return nil
	})
}
