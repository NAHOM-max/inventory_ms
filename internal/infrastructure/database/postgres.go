package database

import (
	"context"
	"database/sql"

	"inventory_ms/internal/domain"
	_ "github.com/lib/pq"
)

// PostgresInventoryRepo implements repository.InventoryRepository
type PostgresInventoryRepo struct {
	db *sql.DB
}

func NewPostgresInventoryRepo(db *sql.DB) *PostgresInventoryRepo {
	return &PostgresInventoryRepo{db: db}
}

func (r *PostgresInventoryRepo) GetByID(ctx context.Context, id string) (*domain.InventoryItem, error) {
	// TODO: implement query
	return nil, nil
}

func (r *PostgresInventoryRepo) Update(ctx context.Context, item *domain.InventoryItem) error {
	// TODO: implement update
	return nil
}

func (r *PostgresInventoryRepo) List(ctx context.Context) ([]*domain.InventoryItem, error) {
	// TODO: implement list query
	return nil, nil
}

// PostgresReservationRepo implements repository.ReservationRepository
type PostgresReservationRepo struct {
	db *sql.DB
}

func NewPostgresReservationRepo(db *sql.DB) *PostgresReservationRepo {
	return &PostgresReservationRepo{db: db}
}

func (r *PostgresReservationRepo) Create(ctx context.Context, res *domain.Reservation) error {
	// TODO: implement insert
	return nil
}

func (r *PostgresReservationRepo) GetByID(ctx context.Context, id string) (*domain.Reservation, error) {
	// TODO: implement query
	return nil, nil
}

func (r *PostgresReservationRepo) Update(ctx context.Context, res *domain.Reservation) error {
	// TODO: implement update
	return nil
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
