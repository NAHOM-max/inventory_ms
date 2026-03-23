package repository

import (
	"context"

	"inventory_ms/internal/domain"
)

type InventoryRepository interface {
	GetByID(ctx context.Context, id string) (*domain.InventoryItem, error)
	Update(ctx context.Context, item *domain.InventoryItem) error
	List(ctx context.Context) ([]*domain.InventoryItem, error)
}
