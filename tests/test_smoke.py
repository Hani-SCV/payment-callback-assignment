"""Minimal happy-path smoke test.

Add tests for the remaining invariants described in README.md.
"""

import pytest
from sqlalchemy import select

from app.models import Payment

TOSS_URL = "/v1/payment-callbacks/toss/return"
pytestmark = pytest.mark.usefixtures("reset_database")


def test_toss_return_completes_payment(client, db_session):
    response = client.post(
        TOSS_URL,
        json={
            "paymentKey": "toss_key_demo_1001",
            "orderId": "pay_demo_toss_001",
            "amount": 129900,
        },
    )

    assert response.status_code == 200
    assert response.json() == {
        "result": "completed",
        "payment_id": "pay_demo_toss_001",
        "order_status": "PAID",
    }

    payment = db_session.scalar(
        select(Payment).where(Payment.public_id == "pay_demo_toss_001")
    )
    assert payment is not None
    assert payment.status == "COMPLETED"
