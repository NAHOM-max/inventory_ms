package usecase

import (
	"context"

	"inventory_ms/internal/domain"
)

type DeliverUseCase struct {
	inventoryRepo   domain.InventoryRepository
	reservationRepo domain.ReservationRepository
}

func NewDeliverUseCase(inv domain.InventoryRepository, res domain.ReservationRepository) *DeliverUseCase {
	return &DeliverUseCase{inventoryRepo: inv, reservationRepo: res}
}

func (uc *DeliverUseCase) Execute(ctx context.Context, reservationID string) error {
	// TODO: implement deliver logic
	return nil
}
