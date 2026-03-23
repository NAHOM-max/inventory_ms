package usecase

import (
	"context"
	"fmt"

	"inventory_ms/internal/domain"
)

type ReturnUseCase struct {
	inventoryRepo   domain.InventoryRepository
	reservationRepo domain.ReservationRepository
	transactor      domain.Transactor
}

func NewReturnUseCase(
	inv domain.InventoryRepository,
	res domain.ReservationRepository,
	tx domain.Transactor,
) *ReturnUseCase {
	return &ReturnUseCase{inventoryRepo: inv, reservationRepo: res, transactor: tx}
}

func (uc *ReturnUseCase) Execute(ctx context.Context, orderID string) error {
	if orderID == "" {
		return domain.NewValidationError("order_id", "must not be empty")
	}

	return uc.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		reservation, err := uc.reservationRepo.LockByOrderID(ctx, orderID)
		if err != nil {
			return fmt.Errorf("lock reservation %s: %w", orderID, err)
		}

		// Idempotent: already returned.
		if reservation.Status == domain.StatusReturned {
			return nil
		}

		// Guard: only RESERVED → RETURNED is valid.
		if reservation.Status != domain.StatusReserved {
			return domain.NewTransitionError(orderID, reservation.Status, domain.StatusReturned)
		}

		items, err := uc.reservationRepo.GetItems(ctx, orderID)
		if err != nil {
			return fmt.Errorf("get items for order %s: %w", orderID, err)
		}

		productIDs := make([]string, len(items))
		for i, it := range items {
			productIDs[i] = it.ProductID
		}

		rows, err := uc.inventoryRepo.LockByIDs(ctx, productIDs)
		if err != nil {
			return fmt.Errorf("lock inventory: %w", err)
		}

		stock := make(map[string]*domain.Inventory, len(rows))
		for i := range rows {
			stock[rows[i].ProductID] = &rows[i]
		}

		for _, it := range items {
			inv, ok := stock[it.ProductID]
			if !ok {
				return fmt.Errorf("product %s: %w", it.ProductID, domain.ErrNotFound)
			}
			if err := inv.Release(it.Amount); err != nil {
				return domain.NewStockError(it.ProductID, it.Amount, inv.ReservedAmount, domain.ErrInsufficientReserve)
			}
		}

		updated := make([]domain.Inventory, 0, len(stock))
		for _, inv := range stock {
			updated = append(updated, *inv)
		}
		if err := uc.inventoryRepo.UpdateBatch(ctx, updated); err != nil {
			return fmt.Errorf("update inventory batch: %w", err)
		}

		if err := uc.reservationRepo.UpdateStatus(ctx, orderID, domain.StatusReturned); err != nil {
			return fmt.Errorf("update reservation status: %w", err)
		}

		return nil
	})
}
