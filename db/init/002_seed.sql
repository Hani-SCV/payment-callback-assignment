-- All records below are synthetic. They do not originate from production data.
INSERT INTO orders (public_id, customer_reference, status, created_at, updated_at)
VALUES
    ('ord_demo_1001', 'customer_demo_a', 'PAYMENT_PENDING', '2026-01-10 01:00:00', '2026-01-10 01:00:00'),
    ('ord_demo_1002', 'customer_demo_b', 'PAID', '2026-01-11 02:00:00', '2026-01-11 02:05:00'),
    ('ord_demo_1003', 'customer_demo_c', 'PAYMENT_PENDING', '2026-01-12 03:00:00', '2026-01-12 03:00:00'),
    ('ord_demo_1004', 'customer_demo_d', 'CANCELED', '2026-01-13 04:00:00', '2026-01-13 04:10:00'),
    ('ord_demo_1005', 'customer_demo_e', 'PAYMENT_PENDING', '2026-01-14 05:00:00', '2026-01-14 05:00:00'),
    ('ord_demo_1006', 'customer_demo_f', 'PAYMENT_PENDING', '2026-01-15 06:00:00', '2026-01-15 06:00:00');

INSERT INTO payments (
    public_id,
    order_id,
    provider,
    status,
    amount,
    currency,
    external_transaction_id,
    completed_at,
    created_at,
    updated_at
)
VALUES
    (
        'pay_demo_toss_001',
        (SELECT id FROM orders WHERE public_id = 'ord_demo_1001'),
        'TOSS',
        'PENDING',
        129900.00,
        'KRW',
        NULL,
        NULL,
        '2026-01-10 01:01:00',
        '2026-01-10 01:01:00'
    ),
    (
        'pay_demo_stripe_competing_001',
        (SELECT id FROM orders WHERE public_id = 'ord_demo_1001'),
        'STRIPE',
        'PENDING',
        129900.00,
        'KRW',
        NULL,
        NULL,
        '2026-01-10 01:02:00',
        '2026-01-10 01:02:00'
    ),
    (
        'pay_demo_stripe_completed_001',
        (SELECT id FROM orders WHERE public_id = 'ord_demo_1002'),
        'STRIPE',
        'COMPLETED',
        45.00,
        'USD',
        'cs_demo_completed_001',
        '2026-01-11 02:05:00',
        '2026-01-11 02:01:00',
        '2026-01-11 02:05:00'
    ),
    (
        'pay_demo_stripe_001',
        (SELECT id FROM orders WHERE public_id = 'ord_demo_1003'),
        'STRIPE',
        'PENDING',
        77.00,
        'USD',
        NULL,
        NULL,
        '2026-01-12 03:01:00',
        '2026-01-12 03:01:00'
    ),
    (
        'pay_demo_alipay_canceled_order_001',
        (SELECT id FROM orders WHERE public_id = 'ord_demo_1004'),
        'ALIPAY',
        'PENDING',
        88.00,
        'CNY',
        NULL,
        NULL,
        '2026-01-13 04:01:00',
        '2026-01-13 04:01:00'
    ),
    (
        'pay_demo_alipay_001',
        (SELECT id FROM orders WHERE public_id = 'ord_demo_1005'),
        'ALIPAY',
        'PENDING',
        88.80,
        'CNY',
        NULL,
        NULL,
        '2026-01-14 05:01:00',
        '2026-01-14 05:01:00'
    ),
    -- Same order, two pending attempts with different amounts.
    (
        'pay_demo_toss_split_001',
        (SELECT id FROM orders WHERE public_id = 'ord_demo_1006'),
        'TOSS',
        'PENDING',
        50000.00,
        'KRW',
        NULL,
        NULL,
        '2026-01-15 06:01:00',
        '2026-01-15 06:01:00'
    ),
    (
        'pay_demo_stripe_split_001',
        (SELECT id FROM orders WHERE public_id = 'ord_demo_1006'),
        'STRIPE',
        'PENDING',
        129900.00,
        'KRW',
        NULL,
        NULL,
        '2026-01-15 06:02:00',
        '2026-01-15 06:02:00'
    );

INSERT INTO payment_events (payment_id, event_type, payload, created_at)
VALUES (
    (SELECT id FROM payments WHERE public_id = 'pay_demo_stripe_completed_001'),
    'PAYMENT_COMPLETED',
    JSON_OBJECT(
        'provider', 'STRIPE',
        'transaction_id', 'cs_demo_completed_001',
        'amount', '45.00',
        'currency', 'USD'
    ),
    '2026-01-11 02:05:00'
);

INSERT INTO outbox_messages (
    deduplication_key,
    event_type,
    aggregate_type,
    aggregate_id,
    payload,
    status,
    created_at
)
VALUES (
    'payment-completed:pay_demo_stripe_completed_001',
    'PAYMENT_COMPLETED',
    'payment',
    'pay_demo_stripe_completed_001',
    JSON_OBJECT(
        'provider', 'STRIPE',
        'payment_id', 'pay_demo_stripe_completed_001',
        'order_id', 'ord_demo_1002',
        'amount', '45.00',
        'currency', 'USD'
    ),
    'PENDING',
    '2026-01-11 02:05:00'
);
