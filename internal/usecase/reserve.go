package usecase

import (
	"context"

	"inventory_ms/internal/repository"
)

type ReserveUseCase struct {
	inventoryRepo   repository.InventoryRepository
	reservationRepo repository.ReservationRepository
}

func NewReserveUseCase(inv repository.InventoryRepository, res repository.ReservationRepository) *ReserveUseCase {
	return &ReserveUseCase{inventoryRepo: inv, reservationRepo: res}
}

func (uc *ReserveUseCase) Execute(ctx context.Context, itemID string, quantity int) error {
	// TODO: implement reserve logic
	return nil
}
