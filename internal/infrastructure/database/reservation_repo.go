package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"inventory_ms/internal/domain"
)

type PostgresReservationRepo struct {
	db *sql.DB
}

func NewPostgresReservationRepo(db *sql.DB) *PostgresReservationRepo {
	return &PostgresReservationRepo{db: db}
}

// GetByOrderID fetches a reservation without locking.
func (r *PostgresReservationRepo) GetByOrderID(ctx context.Context, orderID string) (domain.Reservation, error) {
	return r.queryByOrderID(ctx, orderID, false)
}

// LockByOrderID fetches a reservation with SELECT FOR UPDATE.
// Must be called within a transaction.
func (r *PostgresReservationRepo) LockByOrderID(ctx context.Context, orderID string) (domain.Reservation, error) {
	return r.queryByOrderID(ctx, orderID, true)
}

func (r *PostgresReservationRepo) queryByOrderID(ctx context.Context, orderID string, lock bool) (domain.Reservation, error) {
	q := `SELECT order_id, status, total_price FROM reservations WHERE order_id = $1`
	if lock {
		q += " FOR UPDATE"
	}

	var res domain.Reservation
	err := conn(ctx, r.db).QueryRowContext(ctx, q, orderID).Scan(
		&res.OrderID,
		&res.Status,
		&res.TotalPrice,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Reservation{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Reservation{}, fmt.Errorf("query reservation %s: %w", orderID, err)
	}
	return res, nil
}

// Create inserts the reservation header and all its items atomically.
// reservation_item IDs are derived as "<order_id>-<product_id>" — deterministic and unique per order.
func (r *PostgresReservationRepo) Create(ctx context.Context, res domain.Reservation) error {
	c := conn(ctx, r.db)

	_, err := c.ExecContext(ctx,
		`INSERT INTO reservations (order_id, status, total_price) VALUES ($1, $2, $3)`,
		res.OrderID, res.Status, res.TotalPrice,
	)
	if err != nil {
		return fmt.Errorf("insert reservation %s: %w", res.OrderID, err)
	}

	for _, it := range res.Items {
		itemID := res.OrderID + "-" + it.ProductID
		_, err := c.ExecContext(ctx,
			`INSERT INTO reservation_items (id, order_id, product_id, amount, total_price)
			 VALUES ($1, $2, $3, $4, $5)`,
			itemID, res.OrderID, it.ProductID, it.Amount, it.TotalPrice,
		)
		if err != nil {
			return fmt.Errorf("insert reservation item %s: %w", itemID, err)
		}
	}

	return nil
}

// UpdateStatus sets the status of a reservation by order ID.
func (r *PostgresReservationRepo) UpdateStatus(ctx context.Context, orderID string, status domain.ReservationStatus) error {
	_, err := conn(ctx, r.db).ExecContext(ctx,
		`UPDATE reservations SET status = $1 WHERE order_id = $2`,
		status, orderID,
	)
	if err != nil {
		return fmt.Errorf("update status for order %s: %w", orderID, err)
	}
	return nil
}

// GetItems returns all items belonging to a reservation.
func (r *PostgresReservationRepo) GetItems(ctx context.Context, orderID string) ([]domain.ReservationItem, error) {
	rows, err := conn(ctx, r.db).QueryContext(ctx,
		`SELECT product_id, amount, total_price FROM reservation_items WHERE order_id = $1`,
		orderID,
	)
	if err != nil {
		return nil, fmt.Errorf("query items for order %s: %w", orderID, err)
	}
	defer rows.Close()

	var items []domain.ReservationItem
	for rows.Next() {
		var it domain.ReservationItem
		if err := rows.Scan(&it.ProductID, &it.Amount, &it.TotalPrice); err != nil {
			return nil, fmt.Errorf("scan reservation item: %w", err)
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate reservation items: %w", err)
	}
	return items, nil
}
