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
	return uc.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		// Step 1 — lock the reservation row for the duration of the transaction.
		reservation, err := uc.reservationRepo.LockByOrderID(ctx, orderID)
		if err != nil {
			return fmt.Errorf("lock reservation %s: %w", orderID, err)
		}

		// Step 2 — idempotency: already returned, safe to return success.
		if reservation.Status == domain.StatusReturned {
			return nil
		}

		// Step 2 — guard: only RESERVED reservations can be returned.
		if reservation.Status != domain.StatusReserved {
			return fmt.Errorf("order %s has status %s: %w", orderID, reservation.Status, domain.ErrInvalidTransition)
		}

		// Step 3 — fetch reservation items.
		items, err := uc.reservationRepo.GetItems(ctx, orderID)
		if err != nil {
			return fmt.Errorf("get items for order %s: %w", orderID, err)
		}

		// Step 4 — lock inventory rows (SELECT FOR UPDATE).
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

		// Step 5 — release reserved amounts in memory.
		for _, it := range items {
			inv, ok := stock[it.ProductID]
			if !ok {
				return fmt.Errorf("product %s: %w", it.ProductID, domain.ErrNotFound)
			}
			if err := inv.Release(it.Amount); err != nil {
				return fmt.Errorf("release product %s: %w", it.ProductID, err)
			}
		}

		// Step 6 — persist updated inventory.
		updated := make([]domain.Inventory, 0, len(stock))
		for _, inv := range stock {
			updated = append(updated, *inv)
		}
		if err := uc.inventoryRepo.UpdateBatch(ctx, updated); err != nil {
			return fmt.Errorf("update inventory batch: %w", err)
		}

		// Step 7 — mark reservation as RETURNED.
		if err := uc.reservationRepo.UpdateStatus(ctx, orderID, domain.StatusReturned); err != nil {
			return fmt.Errorf("update reservation status: %w", err)
		}

		return nil
	})
}
