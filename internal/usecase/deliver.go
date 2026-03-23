package usecase

import (
	"context"

	"inventory_ms/internal/repository"
)

type DeliverUseCase struct {
	inventoryRepo   repository.InventoryRepository
	reservationRepo repository.ReservationRepository
}

func NewDeliverUseCase(inv repository.InventoryRepository, res repository.ReservationRepository) *DeliverUseCase {
	return &DeliverUseCase{inventoryRepo: inv, reservationRepo: res}
}

func (uc *DeliverUseCase) Execute(ctx context.Context, reservationID string) error {
	// TODO: implement deliver logic
	return nil
}
