package usecase

import (
	"context"
	"errors"
	"fmt"

	"inventory_ms/internal/domain"
)

type ReserveInput struct {
	OrderID string
	Items   []ReserveItemInput
}

type ReserveItemInput struct {
	ProductID string
	Amount    int
}

type ReserveUseCase struct {
	inventoryRepo   domain.InventoryRepository
	reservationRepo domain.ReservationRepository
	transactor      domain.Transactor
}

func NewReserveUseCase(
	inv domain.InventoryRepository,
	res domain.ReservationRepository,
	tx domain.Transactor,
) *ReserveUseCase {
	return &ReserveUseCase{
		inventoryRepo:   inv,
		reservationRepo: res,
		transactor:      tx,
	}
}

func (uc *ReserveUseCase) Execute(ctx context.Context, input ReserveInput) error {
	// Step 1 — idempotency check (outside transaction, cheap read).
	_, err := uc.reservationRepo.GetByOrderID(ctx, input.OrderID)
	if err == nil {
		// Reservation already exists — return success.
		return nil
	}
	if !errors.Is(err, domain.ErrNotFound) {
		return fmt.Errorf("check existing reservation: %w", err)
	}

	return uc.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		// Step 2 — lock inventory rows (SELECT FOR UPDATE).
		productIDs := make([]string, len(input.Items))
		for i, it := range input.Items {
			productIDs[i] = it.ProductID
		}

		rows, err := uc.inventoryRepo.LockByIDs(ctx, productIDs)
		if err != nil {
			return fmt.Errorf("lock inventory: %w", err)
		}

		// Index locked rows by product ID for O(1) lookup.
		stock := make(map[string]*domain.Inventory, len(rows))
		for i := range rows {
			stock[rows[i].ProductID] = &rows[i]
		}

		// Step 3 — validate stock for every requested item.
		for _, it := range input.Items {
			inv, ok := stock[it.ProductID]
			if !ok {
				return fmt.Errorf("product %s: %w", it.ProductID, domain.ErrNotFound)
			}
			if !inv.CanReserve(it.Amount) {
				return fmt.Errorf("product %s: %w", it.ProductID, domain.ErrInsufficientStock)
			}
		}

		// Step 4 — mutate in-memory inventory (Reserve only raises reserved_amount).
		for _, it := range input.Items {
			if err := stock[it.ProductID].Reserve(it.Amount); err != nil {
				return fmt.Errorf("reserve product %s: %w", it.ProductID, err)
			}
		}

		// Step 5 — persist updated inventory in one batch write.
		updated := make([]domain.Inventory, 0, len(stock))
		for _, inv := range stock {
			updated = append(updated, *inv)
		}
		if err := uc.inventoryRepo.UpdateBatch(ctx, updated); err != nil {
			return fmt.Errorf("update inventory batch: %w", err)
		}

		// Step 6 — build and persist the reservation.
		reservation := buildReservation(input, stock)
		if err := uc.reservationRepo.Create(ctx, reservation); err != nil {
			return fmt.Errorf("create reservation: %w", err)
		}

		// Step 7 — transaction committed by Transactor on nil return.
		return nil
	})
}

func buildReservation(input ReserveInput, stock map[string]*domain.Inventory) domain.Reservation {
	items := make([]domain.ReservationItem, len(input.Items))
	var total float64

	for i, it := range input.Items {
		linePrice := stock[it.ProductID].ProductPrice * float64(it.Amount)
		items[i] = domain.ReservationItem{
			ProductID:  it.ProductID,
			Amount:     it.Amount,
			TotalPrice: linePrice,
		}
		total += linePrice
	}

	return domain.Reservation{
		OrderID:    input.OrderID,
		Status:     domain.StatusReserved,
		TotalPrice: total,
		Items:      items,
	}
}
