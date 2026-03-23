-- Migration: 001_create_tables
-- Direction: DOWN
-- Drops objects in reverse dependency order so foreign-key
-- constraints are never violated during teardown.

BEGIN;

-- Indexes are dropped automatically with their table,
-- but listed explicitly here for clarity.
DROP INDEX IF EXISTS idx_reservation_items_product_id;
DROP INDEX IF EXISTS idx_reservation_items_order_id;

-- reservation_items references both reservations and inventory,
-- so it must be dropped before either parent table.
DROP TABLE IF EXISTS reservation_items;

-- reservations has no remaining dependents at this point.
DROP TABLE IF EXISTS reservations;

-- inventory has no remaining dependents at this point.
DROP TABLE IF EXISTS inventory;

COMMIT;
