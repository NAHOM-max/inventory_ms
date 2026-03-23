package domain

import "context"

type InventoryRepository interface {
	GetByIDs(ctx context.Context, productIDs []string) ([]Inventory, error)
	Update(ctx context.Context, item Inventory) error
	UpdateBatch(ctx context.Context, items []Inventory) error
	// LockByIDs acquires a pessimistic row-level lock (SELECT FOR UPDATE).
	// Must be called within a transaction.
	LockByIDs(ctx context.Context, productIDs []string) ([]Inventory, error)
}
