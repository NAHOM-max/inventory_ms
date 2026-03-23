-- Migration: 001_create_tables
-- Direction: UP
-- Wrapped in a transaction so the migration is fully atomic:
-- either all objects are created or none are.

BEGIN;

-- ============================================================
-- 1. inventory
-- ============================================================
CREATE TABLE inventory (
    product_id       TEXT        PRIMARY KEY,
    product_name     TEXT        NOT NULL,
    product_price    NUMERIC     NOT NULL,
    product_weight   NUMERIC     NOT NULL,
    available_amount INT         NOT NULL CHECK (available_amount >= 0),
    reserved_amount  INT         NOT NULL CHECK (reserved_amount  >= 0),

    -- reserved stock can never exceed physical stock
    CONSTRAINT chk_reserved_lte_available
        CHECK (reserved_amount <= available_amount)
);

-- ============================================================
-- 2. reservations
-- ============================================================
CREATE TABLE reservations (
    order_id    TEXT        PRIMARY KEY,
    status      TEXT        NOT NULL
                            CHECK (status IN ('RESERVED', 'RETURNED', 'DELIVERED')),
    total_price NUMERIC     NOT NULL,
    created_at  TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================
-- 3. reservation_items
-- ============================================================
CREATE TABLE reservation_items (
    id          TEXT        PRIMARY KEY,
    order_id    TEXT        NOT NULL
                            REFERENCES reservations (order_id) ON DELETE CASCADE,
    product_id  TEXT        NOT NULL
                            REFERENCES inventory (product_id),
    amount      INT         NOT NULL CHECK (amount > 0),
    total_price NUMERIC     NOT NULL
);

-- ============================================================
-- Indexes
-- ============================================================

-- Speeds up GetItems(order_id) and ON DELETE CASCADE resolution
CREATE INDEX idx_reservation_items_order_id
    ON reservation_items (order_id);

-- Speeds up joins / lookups by product across all reservations
CREATE INDEX idx_reservation_items_product_id
    ON reservation_items (product_id);

COMMIT;
