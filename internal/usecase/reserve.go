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
	return &ReserveUseCase{inventoryRepo: inv, reservationRepo: res, transactor: tx}
}

func (uc *ReserveUseCase) Execute(ctx context.Context, input ReserveInput) error {
	if err := validateReserveInput(input); err != nil {
		return err
	}

	// Cheap pre-check outside the transaction for the common case.
	_, err := uc.reservationRepo.GetByOrderID(ctx, input.OrderID)
	if err == nil {
		return nil // already reserved — idempotent success
	}
	if !errors.Is(err, domain.ErrNotFound) {
		return fmt.Errorf("check existing reservation: %w", err)
	}

	return uc.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		productIDs := make([]string, len(input.Items))
		for i, it := range input.Items {
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

		// Validate all items before mutating any.
		for _, it := range input.Items {
			inv, ok := stock[it.ProductID]
			if !ok {
				return fmt.Errorf("product %s: %w", it.ProductID, domain.ErrNotFound)
			}
			if !inv.CanReserve(it.Amount) {
				free := inv.AvailableAmount - inv.ReservedAmount
				return domain.NewStockError(it.ProductID, it.Amount, free, domain.ErrInsufficientStock)
			}
		}

		for _, it := range input.Items {
			if err := stock[it.ProductID].Reserve(it.Amount); err != nil {
				return fmt.Errorf("reserve product %s: %w", it.ProductID, err)
			}
		}

		updated := make([]domain.Inventory, 0, len(stock))
		for _, inv := range stock {
			updated = append(updated, *inv)
		}
		if err := uc.inventoryRepo.UpdateBatch(ctx, updated); err != nil {
			return fmt.Errorf("update inventory batch: %w", err)
		}

		reservation := buildReservation(input, stock)
		if err := uc.reservationRepo.Create(ctx, reservation); err != nil {
			// Another concurrent request won the race and already created it — idempotent.
			if errors.Is(err, domain.ErrAlreadyExists) {
				return nil
			}
			return fmt.Errorf("create reservation: %w", err)
		}

		return nil
	})
}

func validateReserveInput(input ReserveInput) error {
	if input.OrderID == "" {
		return domain.NewValidationError("order_id", "must not be empty")
	}
	if len(input.Items) == 0 {
		return domain.NewValidationError("items", "must contain at least one item")
	}
	for _, it := range input.Items {
		if it.ProductID == "" {
			return domain.NewValidationError("items.product_id", "must not be empty")
		}
		if it.Amount <= 0 {
			return domain.NewValidationError("items.amount", fmt.Sprintf("must be positive, got %d for product %s", it.Amount, it.ProductID))
		}
	}
	return nil
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
