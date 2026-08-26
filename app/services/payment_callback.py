from decimal import Decimal

from sqlalchemy import select, func
from sqlalchemy.orm import Session

from app.models import Order, Payment, PaymentEvent, OutboxMessage
from app.errors import TransactionMismatchError, PaymentValidationError
from app.schemas import StripeWebhookRequest

def complete_toss_payment(
    session: Session,
    payment_id: str,
    transaction_id: str,
    amount: str,
    currency: str,
) -> Payment:
    amount = Decimal(amount)
    
    with session.begin():
        # Payment 잠금
        payment = session.scalar(
            select(Payment)
            .where(Payment.public_id == payment_id)
            .with_for_update()
        )

        if payment is None:
            raise PaymentValidationError(
                code="PAYMENT_NOT_FOUND",
                message="Payment not found.",
            )

        # Order 잠금
        order = session.scalar(
            select(Order)
            .where(Order.id == payment.order_id)
            .with_for_update()
        )

        if order is None:
            raise PaymentValidationError(
                code="ORDER_NOT_FOUND",
                message="Order not found.",
            )

        # 결제 제공자 검증
        if payment.provider != "TOSS":
            raise PaymentValidationError(
                code="PROVIDER_MISMATCH",
                message="Payment provider does not match.",
            )

        # 이미 처리된 동일 콜백인지 확인
        if payment.status == "COMPLETED":
            if payment.external_transaction_id == transaction_id:
                return payment

            raise TransactionMismatchError()

        # 처리 가능한 결제 상태인지 확인
        if payment.status != "PENDING":
            raise PaymentValidationError(
                code="INVALID_PAYMENT_STATUS",
                message="Payment is not in a processable state.",
            )

        # 주문 상태 검증
        if order.status != "PAYMENT_PENDING":
            raise PaymentValidationError(
                code="INVALID_ORDER_STATUS",
                message="Order is not in a payable state.",
            )

        # 금액 검증
        if payment.amount != amount:
            raise PaymentValidationError(
                code="AMOUNT_MISMATCH",
                message="Payment amount does not match.",
            )

        # 통화 검증
        if payment.currency != currency:
            raise PaymentValidationError(
                code="CURRENCY_MISMATCH",
                message="Payment currency does not match.",
            )

        # 결제 완료
        payment.status = "COMPLETED"
        payment.external_transaction_id = transaction_id
        payment.completed_at = func.now()

        # 주문 완료
        order.status = "PAID"

        # Payment 이벤트 기록
        record_payment_event(
            session,
            payment,
            transaction_id,
            amount,
            currency,
        )

        # Outbox 메시지 기록
        record_outbox_event(
            session,
            payment,
            transaction_id,
            amount,
            currency,
        )

        return payment

def complete_stripe_payment(
    session: Session,
    payload: StripeWebhookRequest,
) -> Payment:
    event_id = payload.id
    event_type = payload.type
    data = payload.data.object

    transaction_id = data["id"]
    payment_id = data["client_reference_id"]
    amount = Decimal(data["amount_total"]) / Decimal("100")
    currency = data["currency"].upper()
    payment_status = data["payment_status"]

    with session.begin():
        payment = session.scalar(
            select(Payment)
            .where(Payment.public_id == payment_id)
            .with_for_update()
        )

        if payment is None:
            raise PaymentValidationError(
                code="PAYMENT_NOT_FOUND",
                message="Payment not found.",
            )

        order = session.scalar(
            select(Order)
            .where(Order.id == payment.order_id)
            .with_for_update()
        )

        if order is None:
            raise PaymentValidationError(
                code="ORDER_NOT_FOUND",
                message="Order not found.",
            )

        if payment.provider != "STRIPE":
            raise PaymentValidationError(
                code="PROVIDER_MISMATCH",
                message="Payment provider does not match.",
            )

        # 이미 처리한 Stripe 이벤트인지 확인
        existing_event = session.scalar(
            select(PaymentEvent)
            .where(PaymentEvent.event_id == event_id)
        )

        if existing_event is not None:
            if existing_event.payment_id != payment.id:
                raise PaymentValidationError(
                    code="EVENT_PAYMENT_MISMATCH",
                    message="Stripe event is associated with a different payment.",
                )

            return payment

        if event_type != "checkout.session.completed":
            raise PaymentValidationError(
                code="UNSUPPORTED_EVENT",
                message="Unsupported Stripe event type.",
            )

        if payment_status != "paid":
            raise PaymentValidationError(
                code="PAYMENT_NOT_PAID",
                message="Stripe payment is not paid.",
            )

        # 이미 완료된 결제인데 새로운 이벤트가 들어온 경우
        if payment.status == "COMPLETED":
            if payment.external_transaction_id == transaction_id:
                return payment

            raise TransactionMismatchError()

        if payment.status != "PENDING":
            raise PaymentValidationError(
                code="INVALID_PAYMENT_STATUS",
                message="Payment is not in a processable state.",
            )

        if order.status != "PAYMENT_PENDING":
            raise PaymentValidationError(
                code="INVALID_ORDER_STATUS",
                message="Order is not in a payable state.",
            )

        if payment.amount != amount:
            raise PaymentValidationError(
                code="AMOUNT_MISMATCH",
                message="Payment amount does not match.",
            )

        if payment.currency != currency:
            raise PaymentValidationError(
                code="CURRENCY_MISMATCH",
                message="Payment currency does not match.",
            )

        payment.status = "COMPLETED"
        payment.external_transaction_id = transaction_id
        payment.completed_at = func.now()

        order.status = "PAID"

        record_payment_event(
            session,
            payment,
            transaction_id,
            amount,
            currency,
            event_id,
        )

        record_outbox_event(
            session,
            payment,
            transaction_id,
            amount,
            currency,
        )

        return payment

def complete_alipay_payment(
    session: Session,
    order_id: str,
    trade_no: str,
    total_amount: str,
) -> Payment:
    amount = Decimal(total_amount)

    with session.begin():
        # Payment 잠금
        payment = session.scalar(
            select(Payment)
            .where(Payment.public_id == order_id)
            .with_for_update()
        )

        if payment is None:
            raise PaymentValidationError(
                code="PAYMENT_NOT_FOUND",
                message="Payment not found.",
            )

        # Order 잠금
        order = session.scalar(
            select(Order)
            .where(Order.id == payment.order_id)
            .with_for_update()
        )

        if order is None:
            raise PaymentValidationError(
                code="ORDER_NOT_FOUND",
                message="Order not found.",
            )

        # 결제 제공자 검증
        if payment.provider != "ALIPAY":
            raise PaymentValidationError(
                code="PROVIDER_MISMATCH",
                message="Payment provider does not match.",
            )

        # 이미 처리된 동일 notify인지 확인
        if payment.status == "COMPLETED":
            if payment.external_transaction_id == trade_no:
                return payment

            raise TransactionMismatchError()

        # 처리 가능한 결제 상태인지 확인
        if payment.status != "PENDING":
            raise PaymentValidationError(
                code="INVALID_PAYMENT_STATUS",
                message="Payment is not in a processable state.",
            )

        # 주문 상태 검증
        if order.status != "PAYMENT_PENDING":
            raise PaymentValidationError(
                code="INVALID_ORDER_STATUS",
                message="Order is not in a payable state.",
            )

        # 금액 검증
        if payment.amount != amount:
            raise PaymentValidationError(
                code="AMOUNT_MISMATCH",
                message="Payment amount does not match.",
            )

        # Alipay 결제 완료
        payment.status = "COMPLETED"
        payment.external_transaction_id = trade_no
        payment.completed_at = func.now()

        # 주문 완료
        order.status = "PAID"

        # Payment 이벤트 기록
        record_payment_event(
            session,
            payment,
            trade_no,
            amount,
            payment.currency,
        )

        # Outbox 메시지 기록
        record_outbox_event(
            session,
            payment,
            trade_no,
            amount,
            payment.currency,
        )

        return payment

def record_payment_event(
    session: Session,
    payment: Payment,
    transaction_id: str,
    amount: Decimal,
    currency: str,
    event_id: str | None = None,
) -> None:
    event = PaymentEvent(
        payment_id=payment.id,
        event_id=event_id,
        event_type="PAYMENT_COMPLETED",
        payload={
            "provider": payment.provider,
            "transaction_id": transaction_id,
            "amount": str(amount),
            "currency": currency,
        },
    )

    session.add(event)

def record_outbox_event(
    session: Session,
    payment: Payment,
    transaction_id: str,
    amount: Decimal,
    currency: str,
) -> None:
    event = OutboxMessage(
        deduplication_key=f"payment-completed:{payment.public_id}",
        event_type="PAYMENT_COMPLETED",
        aggregate_type="payment",
        aggregate_id=payment.public_id,
        payload={
            "provider": payment.provider,
            "payment_id": payment.public_id,
            "transaction_id": transaction_id,
            "amount": str(amount),
            "currency": currency,
        },
        status="PENDING",
    )

    session.add(event)