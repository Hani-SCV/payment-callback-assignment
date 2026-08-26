CREATE TABLE orders (
    id BIGINT NOT NULL AUTO_INCREMENT,
    public_id VARCHAR(40) NOT NULL,
    customer_reference VARCHAR(40) NOT NULL,
    status VARCHAR(24) NOT NULL,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    CONSTRAINT uq_orders_public_id UNIQUE (public_id),
    CONSTRAINT chk_orders_status CHECK (
        status IN ('PAYMENT_PENDING', 'PAID', 'CANCELED')
    )
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE payments (
    id BIGINT NOT NULL AUTO_INCREMENT,
    public_id VARCHAR(40) NOT NULL,
    order_id BIGINT NOT NULL,
    provider VARCHAR(24) NOT NULL,
    status VARCHAR(24) NOT NULL,
    amount DECIMAL(12, 2) NOT NULL,
    currency CHAR(3) NOT NULL,
    external_transaction_id VARCHAR(80),
    cancellation_reason VARCHAR(40),
    completed_at DATETIME(6),
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    CONSTRAINT uq_payments_public_id UNIQUE (public_id),
    CONSTRAINT fk_payments_order_id FOREIGN KEY (order_id) REFERENCES orders(id),
    CONSTRAINT chk_payments_status CHECK (
        status IN ('PENDING', 'COMPLETED', 'FAILED', 'CANCELED')
    ),
    CONSTRAINT chk_payments_amount CHECK (amount >= 0),
    CONSTRAINT chk_payments_completion_fields CHECK (
        (
            status = 'COMPLETED'
            AND external_transaction_id IS NOT NULL
            AND completed_at IS NOT NULL
        )
        OR status <> 'COMPLETED'
    ),
    INDEX ix_payments_order_id (order_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE payment_events (
    id BIGINT NOT NULL AUTO_INCREMENT,
    payment_id BIGINT NOT NULL,
    event_id VARCHAR(40) NULL UNIQUE,
    event_type VARCHAR(40) NOT NULL,
    payload JSON NOT NULL,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    CONSTRAINT fk_payment_events_payment_id FOREIGN KEY (payment_id)
        REFERENCES payments(id),
    INDEX ix_payment_events_payment_id (payment_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE outbox_messages (
    id BIGINT NOT NULL AUTO_INCREMENT,
    deduplication_key VARCHAR(120) NOT NULL,
    event_type VARCHAR(40) NOT NULL,
    aggregate_type VARCHAR(40) NOT NULL,
    aggregate_id VARCHAR(40) NOT NULL,
    payload JSON NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    published_at DATETIME(6),
    PRIMARY KEY (id),
    CONSTRAINT chk_outbox_messages_status CHECK (
        status IN ('PENDING', 'PUBLISHED')
    ),
    INDEX ix_outbox_messages_status_created_at (status, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
