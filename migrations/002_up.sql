BEGIN;

CREATE TABLE inbox_events (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id    TEXT        NOT NULL UNIQUE,
    event_type  TEXT        NOT NULL,
    payload     JSONB       NOT NULL,
    processed   BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP
);

COMMIT;
