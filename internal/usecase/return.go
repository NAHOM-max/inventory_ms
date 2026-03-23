package usecase

import (
	"context"

	"inventory_ms/internal/repository"
)

type ReturnUseCase struct {
	inventoryRepo   repository.InventoryRepository
	reservationRepo repository.ReservationRepository
}

func NewReturnUseCase(inv repository.InventoryRepository, res repository.ReservationRepository) *ReturnUseCase {
	return &ReturnUseCase{inventoryRepo: inv, reservationRepo: res}
}

func (uc *ReturnUseCase) Execute(ctx context.Context, reservationID string) error {
	// TODO: implement return logic
	return nil
}
