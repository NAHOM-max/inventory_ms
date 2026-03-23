package usecase

import (
	"context"
	"fmt"

	"inventory_ms/internal/domain"
)

type DeliverUseCase struct {
	inventoryRepo   domain.InventoryRepository
	reservationRepo domain.ReservationRepository
	transactor      domain.Transactor
}

func NewDeliverUseCase(
	inv domain.InventoryRepository,
	res domain.ReservationRepository,
	tx domain.Transactor,
) *DeliverUseCase {
	return &DeliverUseCase{inventoryRepo: inv, reservationRepo: res, transactor: tx}
}

func (uc *DeliverUseCase) Execute(ctx context.Context, orderID string) error {
	return uc.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		// Step 1 — lock the reservation row.
		reservation, err := uc.reservationRepo.LockByOrderID(ctx, orderID)
		if err != nil {
			return fmt.Errorf("lock reservation %s: %w", orderID, err)
		}

		// Step 2 — idempotency: already delivered, safe to return success.
		if reservation.Status == domain.StatusDelivered {
			return nil
		}

		// Step 2 — guard: only RESERVED reservations can be delivered.
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

		// Step 5 — deliver each item: reduces both available_amount and reserved_amount.
		for _, it := range items {
			inv, ok := stock[it.ProductID]
			if !ok {
				return fmt.Errorf("product %s: %w", it.ProductID, domain.ErrNotFound)
			}
			if err := inv.Deliver(it.Amount); err != nil {
				return fmt.Errorf("deliver product %s: %w", it.ProductID, err)
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

		// Step 7 — mark reservation as DELIVERED.
		if err := uc.reservationRepo.UpdateStatus(ctx, orderID, domain.StatusDelivered); err != nil {
			return fmt.Errorf("update reservation status: %w", err)
		}

		return nil
	})
}
