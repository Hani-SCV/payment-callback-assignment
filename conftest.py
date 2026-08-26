from collections.abc import Generator
from datetime import datetime
from decimal import Decimal

import pytest
from fastapi.testclient import TestClient
from sqlalchemy import create_engine, delete
from sqlalchemy.orm import Session

from app.config import get_settings
from app.db import Base
from app.main import app
from app.models import Order, OutboxMessage, Payment, PaymentEvent

TEST_DATABASE_URL = get_settings().database_url
test_engine = create_engine(TEST_DATABASE_URL, pool_pre_ping=True)


def seed_synthetic_data(session: Session) -> None:
    created = datetime(2026, 1, 10, 1, 0)
    orders = {
        "pending_competing": Order(
            public_id="ord_demo_1001",
            customer_reference="customer_demo_a",
            status="PAYMENT_PENDING",
            created_at=created,
            updated_at=created,
        ),
        "completed": Order(
            public_id="ord_demo_1002",
            customer_reference="customer_demo_b",
            status="PAID",
            created_at=created,
            updated_at=created,
        ),
        "stripe": Order(
            public_id="ord_demo_1003",
            customer_reference="customer_demo_c",
            status="PAYMENT_PENDING",
            created_at=created,
            updated_at=created,
        ),
        "canceled": Order(
            public_id="ord_demo_1004",
            customer_reference="customer_demo_d",
            status="CANCELED",
            created_at=created,
            updated_at=created,
        ),
        "alipay": Order(
            public_id="ord_demo_1005",
            customer_reference="customer_demo_e",
            status="PAYMENT_PENDING",
            created_at=created,
            updated_at=created,
        ),
        "split": Order(
            public_id="ord_demo_1006",
            customer_reference="customer_demo_f",
            status="PAYMENT_PENDING",
            created_at=created,
            updated_at=created,
        ),
    }
    session.add_all(orders.values())
    session.flush()

    payments = [
        Payment(
            public_id="pay_demo_toss_001",
            order_id=orders["pending_competing"].id,
            provider="TOSS",
            status="PENDING",
            amount=Decimal("129900.00"),
            currency="KRW",
            created_at=created,
            updated_at=created,
        ),
        Payment(
            public_id="pay_demo_stripe_competing_001",
            order_id=orders["pending_competing"].id,
            provider="STRIPE",
            status="PENDING",
            amount=Decimal("129900.00"),
            currency="KRW",
            created_at=created,
            updated_at=created,
        ),
        Payment(
            public_id="pay_demo_stripe_completed_001",
            order_id=orders["completed"].id,
            provider="STRIPE",
            status="COMPLETED",
            amount=Decimal("45.00"),
            currency="USD",
            external_transaction_id="cs_demo_completed_001",
            completed_at=created,
            created_at=created,
            updated_at=created,
        ),
        Payment(
            public_id="pay_demo_stripe_001",
            order_id=orders["stripe"].id,
            provider="STRIPE",
            status="PENDING",
            amount=Decimal("77.00"),
            currency="USD",
            created_at=created,
            updated_at=created,
        ),
        Payment(
            public_id="pay_demo_alipay_canceled_order_001",
            order_id=orders["canceled"].id,
            provider="ALIPAY",
            status="PENDING",
            amount=Decimal("88.00"),
            currency="CNY",
            created_at=created,
            updated_at=created,
        ),
        Payment(
            public_id="pay_demo_alipay_001",
            order_id=orders["alipay"].id,
            provider="ALIPAY",
            status="PENDING",
            amount=Decimal("88.80"),
            currency="CNY",
            created_at=created,
            updated_at=created,
        ),
        # Same order, two pending attempts with different amounts.
        Payment(
            public_id="pay_demo_toss_split_001",
            order_id=orders["split"].id,
            provider="TOSS",
            status="PENDING",
            amount=Decimal("50000.00"),
            currency="KRW",
            created_at=created,
            updated_at=created,
        ),
        Payment(
            public_id="pay_demo_stripe_split_001",
            order_id=orders["split"].id,
            provider="STRIPE",
            status="PENDING",
            amount=Decimal("129900.00"),
            currency="KRW",
            created_at=created,
            updated_at=created,
        ),
    ]
    session.add_all(payments)
    session.flush()

    completed_payment = payments[2]
    session.add(
        PaymentEvent(
            payment_id=completed_payment.id,
            event_type="PAYMENT_COMPLETED",
            payload={
                "provider": "STRIPE",
                "transaction_id": "cs_demo_completed_001",
                "amount": "45.00",
                "currency": "USD",
            },
            created_at=created,
        )
    )
    session.add(
        OutboxMessage(
            deduplication_key="payment-completed:pay_demo_stripe_completed_001",
            event_type="PAYMENT_COMPLETED",
            aggregate_type="payment",
            aggregate_id="pay_demo_stripe_completed_001",
            payload={
                "provider": "STRIPE",
                "payment_id": "pay_demo_stripe_completed_001",
                "order_id": "ord_demo_1002",
                "amount": "45.00",
                "currency": "USD",
            },
            status="PENDING",
            created_at=created,
        )
    )


@pytest.fixture
def reset_database() -> Generator[None, None, None]:
    Base.metadata.create_all(test_engine)
    with Session(test_engine) as session, session.begin():
        session.execute(delete(OutboxMessage))
        session.execute(delete(PaymentEvent))
        session.execute(delete(Payment))
        session.execute(delete(Order))
        seed_synthetic_data(session)
    yield


@pytest.fixture
def client() -> Generator[TestClient, None, None]:
    with TestClient(app) as test_client:
        yield test_client


@pytest.fixture
def db_session() -> Generator[Session, None, None]:
    with Session(test_engine) as session:
        yield session
