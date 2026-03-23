package domain

import "context"

// Transactor runs fn inside a single database transaction.
// The context passed to fn carries the active transaction so
// repository implementations can extract it via ctx.Value.
// If fn returns an error the transaction is rolled back; otherwise committed.
type Transactor interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}
