-- =============================================================
-- Seed: inventory + reservations
-- =============================================================
-- Reservation layout
--
--  order-reserved-1  (RESERVED)
--    prod-laptop      × 3   →  3 × 1299.99 =  3899.97
--    prod-phone       × 2   →  2 ×  799.99 =  1599.98
--    total = 5499.95
--
--  order-reserved-2  (RESERVED)
--    prod-headphones  × 5   →  5 ×  249.99 =  1249.95
--    prod-monitor     × 1   →  1 ×  549.99 =   549.99
--    total = 1799.94
--
--  order-delivered-1 (DELIVERED)
--    prod-keyboard    × 4   →  4 ×   89.99 =   359.96
--    prod-mouse       × 4   →  4 ×   49.99 =   199.96
--    total = 559.92
--
--  order-returned-1  (RETURNED)
--    prod-webcam      × 2   →  2 ×  129.99 =   259.98
--    total = 259.98
--
-- reserved_amount per product (only RESERVED orders count):
--   prod-laptop      += 3
--   prod-phone       += 2
--   prod-headphones  += 5
--   prod-monitor     += 1
--   (all others stay 0)
-- =============================================================

BEGIN;

-- -------------------------------------------------------------
-- Reset — safe to run repeatedly
-- -------------------------------------------------------------
TRUNCATE reservation_items, reservations, inventory RESTART IDENTITY CASCADE;

-- -------------------------------------------------------------
-- 1. Inventory  (10 products)
-- -------------------------------------------------------------
INSERT INTO inventory
    (product_id, product_name, product_price, product_weight, available_amount, reserved_amount)
VALUES
--  id                  name                 price     weight  avail  reserved
    ('prod-laptop',      'Laptop',            1299.99,  2.10,   120,   3),   -- 3 locked by order-reserved-1
    ('prod-phone',       'Smartphone',         799.99,  0.19,   200,   2),   -- 2 locked by order-reserved-1
    ('prod-headphones',  'Headphones',         249.99,  0.31,   150,   5),   -- 5 locked by order-reserved-2
    ('prod-monitor',     'Monitor 27"',        549.99,  5.80,    80,   1),   -- 1 locked by order-reserved-2
    ('prod-keyboard',    'Mechanical Keyboard', 89.99,  0.95,   175,   0),
    ('prod-mouse',       'Wireless Mouse',      49.99,  0.12,   190,   0),
    ('prod-webcam',      'Webcam 1080p',       129.99,  0.24,   100,   0),
    ('prod-tablet',      'Tablet 10"',         499.99,  0.48,    90,   0),
    ('prod-charger',     'USB-C Charger 65W',   39.99,  0.18,   160,   0),
    ('prod-ssd',         'SSD 1TB',            109.99,  0.07,   140,   0);

-- -------------------------------------------------------------
-- 2. Reservations
-- -------------------------------------------------------------
INSERT INTO reservations (order_id, status, total_price, created_at)
VALUES
    ('order-reserved-1',  'RESERVED',  5499.95, '2024-01-10 09:00:00'),
    ('order-reserved-2',  'RESERVED',  1799.94, '2024-01-11 11:30:00'),
    ('order-delivered-1', 'DELIVERED',  559.92, '2024-01-05 14:00:00'),
    ('order-returned-1',  'RETURNED',   259.98, '2024-01-07 16:45:00');

-- -------------------------------------------------------------
-- 3. Reservation items
-- id = <order_id>-<product_id>  (matches reservation_repo.Create)
-- -------------------------------------------------------------
INSERT INTO reservation_items (id, order_id, product_id, amount, total_price)
VALUES
    -- order-reserved-1
    ('order-reserved-1-prod-laptop',     'order-reserved-1',  'prod-laptop',     3,  3899.97),
    ('order-reserved-1-prod-phone',      'order-reserved-1',  'prod-phone',      2,  1599.98),

    -- order-reserved-2
    ('order-reserved-2-prod-headphones', 'order-reserved-2',  'prod-headphones', 5,  1249.95),
    ('order-reserved-2-prod-monitor',    'order-reserved-2',  'prod-monitor',    1,   549.99),

    -- order-delivered-1
    ('order-delivered-1-prod-keyboard',  'order-delivered-1', 'prod-keyboard',   4,   359.96),
    ('order-delivered-1-prod-mouse',     'order-delivered-1', 'prod-mouse',      4,   199.96),

    -- order-returned-1
    ('order-returned-1-prod-webcam',     'order-returned-1',  'prod-webcam',     2,   259.98);

COMMIT;

-- =============================================================
-- Test scenarios enabled by this seed
-- =============================================================
--
-- [reserve endpoint]
--   Happy path  — reserve prod-tablet × 2, prod-ssd × 3 (new order, ample stock)
--   Idempotency — POST reserve with order-reserved-1 again → 201, no state change
--   Conflict    — reserve prod-laptop × 120 (only 117 free: 120 avail − 3 reserved)
--
-- [return endpoint]
--   Happy path  — return order-reserved-1 → status RETURNED, reserved_amount decremented
--   Idempotency — return order-returned-1 again → 200, no state change
--   Conflict    — return order-delivered-1 → 409 invalid transition (DELIVERED → RETURNED)
--
-- [deliver endpoint]
--   Happy path  — deliver order-reserved-2 → status DELIVERED, both amounts decremented
--   Idempotency — deliver order-delivered-1 again → 200, no state change
--   Conflict    — deliver order-returned-1 → 409 invalid transition (RETURNED → DELIVERED)
--
-- [concurrency]
--   Fire multiple simultaneous reserve requests for prod-laptop (117 free units)
--   with amounts that sum to more than 117 — only requests that fit should succeed.
-- =============================================================
