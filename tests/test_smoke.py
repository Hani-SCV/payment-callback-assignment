"""Minimal happy-path smoke test.

Add tests for the remaining invariants described in README.md.
"""

import pytest
from sqlalchemy import select

from app.models import OutboxMessage, Payment, PaymentEvent

TOSS_URL = "/v1/payment-callbacks/toss/return"
STRIPE_URL = "/v1/payment-callbacks/stripe/webhook"

pytestmark = pytest.mark.usefixtures("reset_database")


def test_toss_return_completes_payment(client, db_session):
    """정상 Toss 콜백 처리 시 Payment와 Order를 완료하고 Event와 Outbox를 생성"""
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
        select(Payment).where(
            Payment.public_id == "pay_demo_toss_001"
        )
    )

    assert payment is not None
    assert payment.status == "COMPLETED"

    payment_event = db_session.scalar(
        select(PaymentEvent).where(
            PaymentEvent.payment_id == payment.id
        )
    )

    outbox = db_session.scalar(
        select(OutboxMessage).where(
            OutboxMessage.aggregate_id == "pay_demo_toss_001"
        )
    )

    assert payment_event is not None
    assert payment_event.event_type == "PAYMENT_COMPLETED"

    assert outbox is not None
    assert outbox.status == "PENDING"


def test_toss_return_is_idempotent(client, db_session):
    """동일한 Toss 콜백을 재전송해도 중복 Event와 Outbox를 생성 방지 테스트"""
    payload = {
        "paymentKey": "toss_key_demo_1001",
        "orderId": "pay_demo_toss_001",
        "amount": 129900,
    }

    first_response = client.post(TOSS_URL, json=payload)
    second_response = client.post(TOSS_URL, json=payload)

    assert first_response.status_code == 200
    assert second_response.status_code == 200

    assert second_response.json() == {
        "result": "completed",
        "payment_id": "pay_demo_toss_001",
        "order_status": "PAID",
    }

    payment = db_session.scalar(
        select(Payment).where(
            Payment.public_id == "pay_demo_toss_001"
        )
    )

    assert payment is not None
    assert payment.status == "COMPLETED"

    payment_events = db_session.scalars(
        select(PaymentEvent).where(
            PaymentEvent.payment_id == payment.id
        )
    ).all()

    assert len(payment_events) == 1

    outbox_messages = db_session.scalars(
        select(OutboxMessage).where(
            OutboxMessage.aggregate_id == "pay_demo_toss_001"
        )
    ).all()

    assert len(outbox_messages) == 1


def test_toss_return_rejects_different_transaction_id(client):
    """완료된 Payment에 다른 transaction_id가 들어오면 요청을 거부"""
    first_payload = {
        "paymentKey": "toss_key_demo_1001",
        "orderId": "pay_demo_toss_001",
        "amount": 129900,
    }

    second_payload = {
        "paymentKey": "toss_key_demo_9999",
        "orderId": "pay_demo_toss_001",
        "amount": 129900,
    }

    first_response = client.post(TOSS_URL, json=first_payload)
    second_response = client.post(TOSS_URL, json=second_payload)

    assert first_response.status_code == 200
    assert second_response.status_code == 400

    assert second_response.json() == {
        "error": {
            "code": "TRANSACTION_MISMATCH",
            "message": "Transaction ID does not match the completed payment.",
        }
    }


def test_toss_return_rejects_amount_mismatch(client):
    """Toss 콜백의 결제 금액이 Payment 금액과 다르면 요청 거부"""
    response = client.post(
        TOSS_URL,
        json={
            "paymentKey": "toss_key_demo_1001",
            "orderId": "pay_demo_toss_001",
            "amount": 100000,
        },
    )

    assert response.status_code == 400
    assert response.json() == {
        "error": {
            "code": "AMOUNT_MISMATCH",
            "message": "Payment amount does not match.",
        }
    }


def test_toss_return_currency_mismatch(client, db_session):
    """Payment의 통화와 Toss 콜백의 통화가 다르면 요청 거부"""
    payment = db_session.scalar(
        select(Payment).where(
            Payment.public_id == "pay_demo_toss_001"
        )
    )

    assert payment is not None

    payment.currency = "USD"
    db_session.commit()

    response = client.post(
        TOSS_URL,
        json={
            "paymentKey": "toss_key_demo_1001",
            "orderId": "pay_demo_toss_001",
            "amount": 129900,
        },
    )

    assert response.status_code == 400
    assert response.json() == {
        "error": {
            "code": "CURRENCY_MISMATCH",
            "message": "Payment currency does not match.",
        }
    }


def test_stripe_webhook_completes_payment(client, db_session):
    """정상 Stripe webhook 처리 시 Payment와 Order를 완료하고 Event와 Outbox를 생성"""
    response = client.post(
        STRIPE_URL,
        json={
            "id": "evt_demo_1001",
            "type": "checkout.session.completed",
            "data": {
                "object": {
                    "id": "cs_demo_1001",
                    "client_reference_id": "pay_demo_stripe_001",
                    "amount_total": 7700,
                    "currency": "usd",
                    "payment_status": "paid",
                }
            },
        },
    )

    assert response.status_code == 200
    assert response.json() == {
        "result": "completed",
        "payment_id": "pay_demo_stripe_001",
        "order_status": "PAID",
    }

    payment = db_session.scalar(
        select(Payment).where(
            Payment.public_id == "pay_demo_stripe_001"
        )
    )

    assert payment is not None
    assert payment.status == "COMPLETED"
    assert payment.external_transaction_id == "cs_demo_1001"

    payment_event = db_session.scalar(
        select(PaymentEvent).where(
            PaymentEvent.payment_id == payment.id
        )
    )

    outbox = db_session.scalar(
        select(OutboxMessage).where(
            OutboxMessage.aggregate_id == payment.public_id
        )
    )

    assert payment_event is not None
    assert payment_event.event_type == "PAYMENT_COMPLETED"

    assert outbox is not None
    assert outbox.status == "PENDING"


def test_stripe_webhook_is_idempotent(client, db_session):
    """동일 event_id의 Stripe webhook 재전송 시 중복 처리를 방지"""
    first_payload = {
        "id": "evt_demo_1001",
        "type": "checkout.session.completed",
        "data": {
            "object": {
                "id": "cs_demo_1001",
                "client_reference_id": "pay_demo_stripe_001",
                "amount_total": 7700,
                "currency": "usd",
                "payment_status": "paid",
            }
        },
    }

    redelivery_payload = {
        "id": "evt_demo_1001",
        "type": "checkout.session.completed",
        "data": {
            "object": {
                "id": "cs_demo_1099",
                "client_reference_id": "pay_demo_stripe_001",
                "amount_total": 7700,
                "currency": "usd",
                "payment_status": "paid",
            }
        },
    }

    first_response = client.post(
        STRIPE_URL,
        json=first_payload,
    )

    second_response = client.post(
        STRIPE_URL,
        json=redelivery_payload,
    )

    assert first_response.status_code == 200
    assert second_response.status_code == 200

    payment = db_session.scalar(
        select(Payment).where(
            Payment.public_id == "pay_demo_stripe_001"
        )
    )

    assert payment is not None
    assert payment.status == "COMPLETED"
    assert payment.external_transaction_id == "cs_demo_1001"

    events = db_session.scalars(
        select(PaymentEvent).where(
            PaymentEvent.payment_id == payment.id
        )
    ).all()

    assert len(events) == 1
    assert events[0].event_id == "evt_demo_1001"

    outbox_messages = db_session.scalars(
        select(OutboxMessage).where(
            OutboxMessage.aggregate_id == payment.public_id
        )
    ).all()

    assert len(outbox_messages) == 1


def test_stripe_webhook_rejects_unpaid_session(client):
    """Stripe payment_status가 paid가 아니면 결제 실패 테스트"""
    response = client.post(
        STRIPE_URL,
        json={
            "id": "evt_demo_unpaid_001",
            "type": "checkout.session.completed",
            "data": {
                "object": {
                    "id": "cs_demo_unpaid_001",
                    "client_reference_id": "pay_demo_stripe_001",
                    "amount_total": 7700,
                    "currency": "usd",
                    "payment_status": "unpaid",
                }
            },
        },
    )

    assert response.status_code == 400
    assert response.json()["error"]["code"] == "PAYMENT_NOT_PAID"


def test_stripe_webhook_rejects_unsupported_currency(client):
    """Stripe webhook의 통화가 Payment의 통화와 다르면 결제 실패 테스트"""
    response = client.post(
        STRIPE_URL,
        json={
            "id": "evt_demo_currency_001",
            "type": "checkout.session.completed",
            "data": {
                "object": {
                    "id": "cs_demo_currency_001",
                    "client_reference_id": "pay_demo_stripe_001",
                    "amount_total": 7700,
                    "currency": "jpy",
                    "payment_status": "paid",
                }
            },
        },
    )

    assert response.status_code == 400
    assert response.json()["error"]["code"] == "CURRENCY_MISMATCH"

def test_alipay_notify_completes_payment(client, db_session):
    """정상 Alipay notify
    → Payment COMPLETED
    → Order PAID
    → PaymentEvent 생성
    → Outbox 생성
    """
    response = client.post(
        "/v1/payment-callbacks/alipay/notify",
        data={
            "order_id": "pay_demo_alipay_001",
            "trade_no": "alipay_trade_demo_1001",
            "total_amount": "88.80",
        },
    )

    assert response.status_code == 200
    assert response.json() == {
        "result": "completed",
        "payment_id": "pay_demo_alipay_001",
        "order_status": "PAID",
    }

    payment = db_session.scalar(
        select(Payment).where(
            Payment.public_id == "pay_demo_alipay_001"
        )
    )

    assert payment is not None
    assert payment.status == "COMPLETED"
    assert payment.external_transaction_id == "alipay_trade_demo_1001"

    payment_event = db_session.scalar(
        select(PaymentEvent).where(
            PaymentEvent.payment_id == payment.id
        )
    )

    assert payment_event is not None
    assert payment_event.event_type == "PAYMENT_COMPLETED"

    outbox = db_session.scalar(
        select(OutboxMessage).where(
            OutboxMessage.aggregate_id == payment.public_id
        )
    )

    assert outbox is not None
    assert outbox.status == "PENDING"


def test_alipay_notify_is_idempotent(client, db_session):
    """동일한 Alipay notify 재전송
    → 200
    → 중복 Event 생성 X
    → 중복 Outbox 생성 X
    """
    payload = {
        "order_id": "pay_demo_alipay_001",
        "trade_no": "alipay_trade_demo_1001",
        "total_amount": "88.80",
    }

    first_response = client.post(
        "/v1/payment-callbacks/alipay/notify",
        data=payload,
    )
    second_response = client.post(
        "/v1/payment-callbacks/alipay/notify",
        data=payload,
    )

    assert first_response.status_code == 200
    assert second_response.status_code == 200

    payment = db_session.scalar(
        select(Payment).where(
            Payment.public_id == "pay_demo_alipay_001"
        )
    )

    assert payment is not None
    assert payment.status == "COMPLETED"
    assert payment.external_transaction_id == "alipay_trade_demo_1001"

    events = db_session.scalars(
        select(PaymentEvent).where(
            PaymentEvent.payment_id == payment.id
        )
    ).all()

    assert len(events) == 1

    outbox_messages = db_session.scalars(
        select(OutboxMessage).where(
            OutboxMessage.aggregate_id == payment.public_id
        )
    ).all()

    assert len(outbox_messages) == 1


def test_alipay_notify_rejects_different_trade_no(client):
    """이미 완료된 Payment에 다른 trade_no 요청
    → 400 TRANSACTION_MISMATCH
    """
    first_response = client.post(
        "/v1/payment-callbacks/alipay/notify",
        data={
            "order_id": "pay_demo_alipay_001",
            "trade_no": "alipay_trade_demo_1001",
            "total_amount": "88.80",
        },
    )

    second_response = client.post(
        "/v1/payment-callbacks/alipay/notify",
        data={
            "order_id": "pay_demo_alipay_001",
            "trade_no": "alipay_trade_demo_9999",
            "total_amount": "88.80",
        },
    )

    assert first_response.status_code == 200

    assert second_response.status_code == 400
    assert second_response.json() == {
        "error": {
            "code": "TRANSACTION_MISMATCH",
            "message": "Transaction ID does not match the completed payment.",
        }
    }


def test_alipay_notify_rejects_amount_mismatch(client):
    """Alipay notify 금액과 Payment 금액이 다르면 거부
    → 400 AMOUNT_MISMATCH
    """
    response = client.post(
        "/v1/payment-callbacks/alipay/notify",
        data={
            "order_id": "pay_demo_alipay_001",
            "trade_no": "alipay_trade_demo_1001",
            "total_amount": "100.00",
        },
    )

    assert response.status_code == 400
    assert response.json() == {
        "error": {
            "code": "AMOUNT_MISMATCH",
            "message": "Payment amount does not match.",
        }
    }


def test_alipay_notify_rejects_canceled_order(client):
    """취소된 주문의 Alipay notify는 결제를 완료하지 않음
    → 400 INVALID_ORDER_STATUS
    """
    response = client.post(
        "/v1/payment-callbacks/alipay/notify",
        data={
            "order_id": "pay_demo_alipay_canceled_order_001",
            "trade_no": "alipay_trade_demo_canceled_001",
            "total_amount": "88.00",
        },
    )

    assert response.status_code == 400
    assert response.json() == {
        "error": {
            "code": "INVALID_ORDER_STATUS",
            "message": "Order is not in a payable state.",
        }
    }

def test_toss_return_rejects_provider_mismatch(client):
    """Toss callback이 Stripe Payment를 대상으로 하면 provider mismatch로 거부"""
    response = client.post(
        TOSS_URL,
        json={
            "paymentKey": "toss_key_demo_1001",
            "orderId": "pay_demo_stripe_001",
            "amount": 129900,
        },
    )

    assert response.status_code == 400
    assert response.json() == {
        "error": {
            "code": "PROVIDER_MISMATCH",
            "message": "Payment provider does not match.",
        }
    }


def test_toss_return_rejects_payment_not_found(client):
    """존재하지 않는 Payment에 대한 Toss callback을 거부"""
    response = client.post(
        TOSS_URL,
        json={
            "paymentKey": "toss_key_demo_1001",
            "orderId": "pay_demo_unknown_001",
            "amount": 129900,
        },
    )

    assert response.status_code == 400
    assert response.json() == {
        "error": {
            "code": "PAYMENT_NOT_FOUND",
            "message": "Payment not found.",
        }
    }


def test_stripe_webhook_rejects_unsupported_event(client):
    """지원하지 않는 Stripe event type은 처리하지 않음"""
    response = client.post(
        STRIPE_URL,
        json={
            "id": "evt_demo_unsupported_001",
            "type": "payment_intent.created",
            "data": {
                "object": {
                    "id": "pi_demo_001",
                    "client_reference_id": "pay_demo_stripe_001",
                    "amount_total": 7700,
                    "currency": "usd",
                    "payment_status": "paid",
                }
            },
        },
    )

    assert response.status_code == 400
    assert response.json() == {
        "error": {
            "code": "UNSUPPORTED_EVENT",
            "message": "Unsupported Stripe event type.",
        }
    }


def test_stripe_webhook_rejects_different_transaction_id(client):
    """완료된 Stripe Payment에 다른 transaction_id가 들어오면 거부"""
    first_response = client.post(
        STRIPE_URL,
        json={
            "id": "evt_demo_stripe_first_001",
            "type": "checkout.session.completed",
            "data": {
                "object": {
                    "id": "cs_demo_stripe_1001",
                    "client_reference_id": "pay_demo_stripe_001",
                    "amount_total": 7700,
                    "currency": "usd",
                    "payment_status": "paid",
                }
            },
        },
    )

    second_response = client.post(
        STRIPE_URL,
        json={
            "id": "evt_demo_stripe_second_001",
            "type": "checkout.session.completed",
            "data": {
                "object": {
                    "id": "cs_demo_stripe_9999",
                    "client_reference_id": "pay_demo_stripe_001",
                    "amount_total": 7700,
                    "currency": "usd",
                    "payment_status": "paid",
                }
            },
        },
    )

    assert first_response.status_code == 200

    assert second_response.status_code == 400
    assert second_response.json() == {
        "error": {
            "code": "TRANSACTION_MISMATCH",
            "message": "Transaction ID does not match the completed payment.",
        }
    }

def test_stripe_webhook_rejects_event_associated_with_different_payment(
    client,
    db_session,
):
    """이미 다른 Payment에 기록된 event_id가 들어오면 잘못된 결제 연결을 거부"""
    event = db_session.scalar(
        select(PaymentEvent).where(
            PaymentEvent.event_id == "evt_demo_completed_001"
        )
    )

    assert event is not None

    response = client.post(
        STRIPE_URL,
        json={
            "id": "evt_demo_completed_001",
            "type": "checkout.session.completed",
            "data": {
                "object": {
                    "id": "cs_demo_1001",
                    "client_reference_id": "pay_demo_stripe_001",
                    "amount_total": 7700,
                    "currency": "usd",
                    "payment_status": "paid",
                }
            },
        },
    )

    assert response.status_code == 400
    assert response.json() == {
        "error": {
            "code": "EVENT_PAYMENT_MISMATCH",
            "message": "Stripe event is associated with a different payment.",
        }
    }


def test_alipay_notify_rejects_provider_mismatch(client):
    """Alipay notify가 Stripe Payment를 대상으로 하면 provider mismatch로 거부"""
    response = client.post(
        "/v1/payment-callbacks/alipay/notify",
        data={
            "order_id": "pay_demo_stripe_001",
            "trade_no": "alipay_trade_demo_1001",
            "total_amount": "77.00",
        },
    )

    assert response.status_code == 400
    assert response.json() == {
        "error": {
            "code": "PROVIDER_MISMATCH",
            "message": "Payment provider does not match.",
        }
    }


def test_alipay_notify_rejects_payment_not_found(client):
    """존재하지 않는 Payment에 대한 Alipay notify를 거부"""
    response = client.post(
        "/v1/payment-callbacks/alipay/notify",
        data={
            "order_id": "pay_demo_alipay_unknown_001",
            "trade_no": "alipay_trade_demo_1001",
            "total_amount": "88.80",
        },
    )

    assert response.status_code == 400
    assert response.json() == {
        "error": {
            "code": "PAYMENT_NOT_FOUND",
            "message": "Payment not found.",
        }
    }

def test_toss_and_stripe_cannot_both_complete_same_order(
    client,
    db_session,
):
    """같은 주문에 연결된 서로 다른 Payment가 동시에 완료되는 것을 방지"""

    toss_response = client.post(
        TOSS_URL,
        json={
            "paymentKey": "toss_key_demo_1001",
            "orderId": "pay_demo_toss_001",
            "amount": 129900,
        },
    )

    assert toss_response.status_code == 200

    stripe_response = client.post(
        STRIPE_URL,
        json={
            "id": "evt_demo_competing_001",
            "type": "checkout.session.completed",
            "data": {
                "object": {
                    "id": "cs_demo_competing_001",
                    "client_reference_id": "pay_demo_stripe_competing_001",
                    "amount_total": 12990000,
                    "currency": "krw",
                    "payment_status": "paid",
                }
            },
        },
    )

    assert stripe_response.status_code == 400

    toss_payment = db_session.scalar(
        select(Payment).where(
            Payment.public_id == "pay_demo_toss_001"
        )
    )

    stripe_payment = db_session.scalar(
        select(Payment).where(
            Payment.public_id == "pay_demo_stripe_competing_001"
        )
    )

    assert toss_payment is not None
    assert stripe_payment is not None

    assert toss_payment.status == "COMPLETED"
    assert stripe_payment.status == "PENDING"

    events = db_session.scalars(
        select(PaymentEvent).where(
            PaymentEvent.payment_id.in_(
                [toss_payment.id, stripe_payment.id]
            )
        )
    ).all()

    assert len(events) == 1

    outbox_messages = db_session.scalars(
        select(OutboxMessage).where(
            OutboxMessage.aggregate_id.in_(
                [
                    toss_payment.public_id,
                    stripe_payment.public_id,
                ]
            )
        )
    ).all()

    assert len(outbox_messages) == 1

def test_payment_with_different_amount_cannot_complete_same_order(
    client,
    db_session,
):
    """같은 주문의 금액이 다른 Payment가 중복 완료되는 것을 방지"""

    toss_response = client.post(
        TOSS_URL,
        json={
            "paymentKey": "toss_key_demo_split_001",
            "orderId": "pay_demo_toss_split_001",
            "amount": 50000,
        },
    )

    assert toss_response.status_code == 200

    stripe_response = client.post(
        STRIPE_URL,
        json={
            "id": "evt_demo_split_001",
            "type": "checkout.session.completed",
            "data": {
                "object": {
                    "id": "cs_demo_split_001",
                    "client_reference_id": "pay_demo_stripe_split_001",
                    "amount_total": 12990000,
                    "currency": "krw",
                    "payment_status": "paid",
                }
            },
        },
    )

    assert stripe_response.status_code == 400

    toss_payment = db_session.scalar(
        select(Payment).where(
            Payment.public_id == "pay_demo_toss_split_001"
        )
    )

    stripe_payment = db_session.scalar(
        select(Payment).where(
            Payment.public_id == "pay_demo_stripe_split_001"
        )
    )

    assert toss_payment is not None
    assert stripe_payment is not None

    assert toss_payment.status == "COMPLETED"
    assert stripe_payment.status == "PENDING"