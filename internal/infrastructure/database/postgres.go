package database

import (
	"context"
	"database/sql"
	"fmt"

	"inventory_ms/internal/domain"
	_ "github.com/lib/pq"
)

// txKey is the context key used to propagate an active *sql.Tx.
type txKey struct{}

// PostgresTransactor implements domain.Transactor.
type PostgresTransactor struct {
	db *sql.DB
}

func NewPostgresTransactor(db *sql.DB) *PostgresTransactor {
	return &PostgresTransactor{db: db}
}

func (t *PostgresTransactor) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := t.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	if err := fn(context.WithValue(ctx, txKey{}, tx)); err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

// TxFromContext extracts the active transaction from ctx, if any.
// Repository implementations call this to reuse the transaction connection.
func TxFromContext(ctx context.Context) (*sql.Tx, bool) {
	tx, ok := ctx.Value(txKey{}).(*sql.Tx)
	return tx, ok
}

// PostgresInventoryRepo implements domain.InventoryRepository.
type PostgresInventoryRepo struct {
	db *sql.DB
}

func NewPostgresInventoryRepo(db *sql.DB) *PostgresInventoryRepo {
	return &PostgresInventoryRepo{db: db}
}

func (r *PostgresInventoryRepo) GetByIDs(ctx context.Context, productIDs []string) ([]domain.Inventory, error) {
	// TODO: implement
	return nil, nil
}

func (r *PostgresInventoryRepo) Update(ctx context.Context, item domain.Inventory) error {
	// TODO: implement
	return nil
}

func (r *PostgresInventoryRepo) UpdateBatch(ctx context.Context, items []domain.Inventory) error {
	// TODO: implement
	return nil
}

func (r *PostgresInventoryRepo) LockByIDs(ctx context.Context, productIDs []string) ([]domain.Inventory, error) {
	// TODO: implement SELECT FOR UPDATE
	return nil, nil
}

// PostgresReservationRepo implements domain.ReservationRepository.
type PostgresReservationRepo struct {
	db *sql.DB
}

func NewPostgresReservationRepo(db *sql.DB) *PostgresReservationRepo {
	return &PostgresReservationRepo{db: db}
}

func (r *PostgresReservationRepo) GetByOrderID(ctx context.Context, orderID string) (domain.Reservation, error) {
	// TODO: implement
	return domain.Reservation{}, domain.ErrNotFound
}

func (r *PostgresReservationRepo) LockByOrderID(ctx context.Context, orderID string) (domain.Reservation, error) {
	// TODO: implement SELECT FOR UPDATE
	return domain.Reservation{}, domain.ErrNotFound
}

func (r *PostgresReservationRepo) Create(ctx context.Context, res domain.Reservation) error {
	// TODO: implement
	return nil
}

func (r *PostgresReservationRepo) UpdateStatus(ctx context.Context, orderID string, status domain.ReservationStatus) error {
	// TODO: implement
	return nil
}

func (r *PostgresReservationRepo) GetItems(ctx context.Context, orderID string) ([]domain.ReservationItem, error) {
	// TODO: implement
	return nil, nil
}

// NewPostgresDB opens and verifies a postgres connection.
func NewPostgresDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
