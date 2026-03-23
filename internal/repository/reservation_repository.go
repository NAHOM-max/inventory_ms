package repository

import (
	"context"

	"inventory_ms/internal/domain"
)

type ReservationRepository interface {
	Create(ctx context.Context, r *domain.Reservation) error
	GetByID(ctx context.Context, id string) (*domain.Reservation, error)
	Update(ctx context.Context, r *domain.Reservation) error
}
