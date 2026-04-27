package domain

import (
	"context"
	"encoding/json"
	"time"
)

type InboxEvent struct {
	ID        string
	EventID   string
	EventType string
	Payload   json.RawMessage
	Processed bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type InboxRepository interface {
	Exists(ctx context.Context, eventID string) (bool, error)
	Save(ctx context.Context, event InboxEvent) error
	MarkProcessed(ctx context.Context, eventID string) error
}
