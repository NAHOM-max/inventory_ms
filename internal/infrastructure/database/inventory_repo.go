package database

import (
	"context"
	"database/sql"
	"fmt"

	"inventory_ms/internal/domain"

	"github.com/lib/pq"
)

type PostgresInventoryRepo struct {
	db *sql.DB
}

func NewPostgresInventoryRepo(db *sql.DB) *PostgresInventoryRepo {
	return &PostgresInventoryRepo{db: db}
}

// GetByIDs fetches inventory rows by product IDs without locking.
func (r *PostgresInventoryRepo) GetByIDs(ctx context.Context, productIDs []string) ([]domain.Inventory, error) {
	return r.queryByIDs(ctx, productIDs, false)
}

// LockByIDs fetches inventory rows with SELECT FOR UPDATE.
// Must be called within a transaction.
func (r *PostgresInventoryRepo) LockByIDs(ctx context.Context, productIDs []string) ([]domain.Inventory, error) {
	return r.queryByIDs(ctx, productIDs, true)
}

func (r *PostgresInventoryRepo) queryByIDs(ctx context.Context, productIDs []string, lock bool) ([]domain.Inventory, error) {
	q := `
		SELECT product_id, product_name, product_price, product_weight,
		       available_amount, reserved_amount
		FROM inventory
		WHERE product_id = ANY($1)`
	if lock {
		q += " FOR UPDATE"
	}

	rows, err := conn(ctx, r.db).QueryContext(ctx, q, pq.Array(productIDs))
	if err != nil {
		return nil, fmt.Errorf("query inventory: %w", err)
	}
	defer rows.Close()

	return scanInventoryRows(rows)
}

// Update writes a single inventory row.
func (r *PostgresInventoryRepo) Update(ctx context.Context, item domain.Inventory) error {
	q := `
		UPDATE inventory
		SET product_name     = $1,
		    product_price    = $2,
		    product_weight   = $3,
		    available_amount = $4,
		    reserved_amount  = $5
		WHERE product_id = $6`

	_, err := conn(ctx, r.db).ExecContext(ctx, q,
		item.ProductName,
		item.ProductPrice,
		item.ProductWeight,
		item.AvailableAmount,
		item.ReservedAmount,
		item.ProductID,
	)
	if err != nil {
		return fmt.Errorf("update inventory %s: %w", item.ProductID, err)
	}
	return nil
}

// UpdateBatch updates multiple inventory rows in a single round-trip using UNNEST.
func (r *PostgresInventoryRepo) UpdateBatch(ctx context.Context, items []domain.Inventory) error {
	if len(items) == 0 {
		return nil
	}

	ids := make([]string, len(items))
	names := make([]string, len(items))
	prices := make([]float64, len(items))
	weights := make([]float64, len(items))
	available := make([]int, len(items))
	reserved := make([]int, len(items))

	for i, it := range items {
		ids[i] = it.ProductID
		names[i] = it.ProductName
		prices[i] = it.ProductPrice
		weights[i] = it.ProductWeight
		available[i] = it.AvailableAmount
		reserved[i] = it.ReservedAmount
	}

	q := `
		UPDATE inventory AS inv
		SET product_name     = v.product_name,
		    product_price    = v.product_price,
		    product_weight   = v.product_weight,
		    available_amount = v.available_amount,
		    reserved_amount  = v.reserved_amount
		FROM (
			SELECT
				UNNEST($1::text[])    AS product_id,
				UNNEST($2::text[])    AS product_name,
				UNNEST($3::numeric[]) AS product_price,
				UNNEST($4::numeric[]) AS product_weight,
				UNNEST($5::int[])     AS available_amount,
				UNNEST($6::int[])     AS reserved_amount
		) AS v
		WHERE inv.product_id = v.product_id`

	_, err := conn(ctx, r.db).ExecContext(ctx, q,
		pq.Array(ids),
		pq.Array(names),
		pq.Array(prices),
		pq.Array(weights),
		pq.Array(available),
		pq.Array(reserved),
	)
	if err != nil {
		return fmt.Errorf("batch update inventory: %w", err)
	}
	return nil
}

func scanInventoryRows(rows *sql.Rows) ([]domain.Inventory, error) {
	var result []domain.Inventory
	for rows.Next() {
		var inv domain.Inventory
		if err := rows.Scan(
			&inv.ProductID,
			&inv.ProductName,
			&inv.ProductPrice,
			&inv.ProductWeight,
			&inv.AvailableAmount,
			&inv.ReservedAmount,
		); err != nil {
			return nil, fmt.Errorf("scan inventory row: %w", err)
		}
		result = append(result, inv)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate inventory rows: %w", err)
	}
	return result, nil
}

