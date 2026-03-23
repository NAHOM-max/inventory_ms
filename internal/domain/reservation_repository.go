package domain

import "context"

type ReservationRepository interface {
	GetByOrderID(ctx context.Context, orderID string) (Reservation, error)
	// LockByOrderID acquires a row-level lock on the reservation (SELECT FOR UPDATE).
	// Must be called within a transaction.
	LockByOrderID(ctx context.Context, orderID string) (Reservation, error)
	Create(ctx context.Context, r Reservation) error
	UpdateStatus(ctx context.Context, orderID string, status ReservationStatus) error
	GetItems(ctx context.Context, orderID string) ([]ReservationItem, error)
}
