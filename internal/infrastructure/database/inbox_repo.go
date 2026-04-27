package database

import (
	"context"
	"database/sql"
	"fmt"

	"inventory_ms/internal/domain"
)

type PostgresInboxRepo struct {
	db *sql.DB
}

func NewPostgresInboxRepo(db *sql.DB) *PostgresInboxRepo {
	return &PostgresInboxRepo{db: db}
}

func (r *PostgresInboxRepo) Exists(ctx context.Context, eventID string) (bool, error) {
	var processed bool
	err := conn(ctx, r.db).QueryRowContext(ctx,
		`SELECT processed FROM inbox_events WHERE event_id = $1`, eventID,
	).Scan(&processed)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("inbox exists: %w", err)
	}
	return processed, nil
}

func (r *PostgresInboxRepo) Save(ctx context.Context, event domain.InboxEvent) error {
	_, err := conn(ctx, r.db).ExecContext(ctx,
		`INSERT INTO inbox_events (event_id, event_type, payload) VALUES ($1, $2, $3)`,
		event.EventID, event.EventType, event.Payload,
	)
	if err != nil {
		return fmt.Errorf("inbox save: %w", err)
	}
	return nil
}

func (r *PostgresInboxRepo) MarkProcessed(ctx context.Context, eventID string) error {
	_, err := conn(ctx, r.db).ExecContext(ctx,
		`UPDATE inbox_events SET processed = TRUE, updated_at = CURRENT_TIMESTAMP WHERE event_id = $1`,
		eventID,
	)
	if err != nil {
		return fmt.Errorf("inbox mark processed: %w", err)
	}
	return nil
}
