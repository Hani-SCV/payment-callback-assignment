from typing import Annotated

from decimal import Decimal
from fastapi import APIRouter, Depends, Form
from sqlalchemy.orm import Session

from app.db import get_db
from app.schemas import (
    ErrorResponse,
    PaymentCallbackResponse,
    StripeWebhookRequest,
    TossReturnRequest,
)

from app.services.payment_callback import (
    complete_toss_payment, 
    complete_stripe_payment,
    complete_alipay_payment,
)

router = APIRouter(prefix="/v1/payment-callbacks", tags=["payment-callbacks"])

ERROR_RESPONSES = {
    400: {"model": ErrorResponse},
    404: {"model": ErrorResponse},
    409: {"model": ErrorResponse},
    422: {"model": ErrorResponse},
    501: {"model": ErrorResponse},
}


@router.post(
    "/toss/return",
    response_model=PaymentCallbackResponse,
    responses=ERROR_RESPONSES,
)
def receive_toss_return(
    payload: TossReturnRequest,
    session: Annotated[Session, Depends(get_db)],
) -> PaymentCallbackResponse:

    payment = complete_toss_payment(
        session=session,
        payment_id=payload.order_id,
        transaction_id=payload.payment_key,
        amount=payload.amount,
        currency="KRW",
    )

    return PaymentCallbackResponse(
        result="completed",
        payment_id=payment.public_id,
        order_status=payment.order.status,
    )


@router.post(
    "/stripe/webhook",
    response_model=PaymentCallbackResponse,
    responses=ERROR_RESPONSES,
)
def receive_stripe_webhook(
    payload: StripeWebhookRequest,
    session: Annotated[Session, Depends(get_db)],
) -> PaymentCallbackResponse:
    payment = complete_stripe_payment(
        session=session,
        payload=payload,
    )

    return PaymentCallbackResponse(
        result="completed",
        payment_id=payment.public_id,
        order_status=payment.order.status,
    )


@router.post(
    "/alipay/notify",
    response_model=PaymentCallbackResponse,
    responses=ERROR_RESPONSES,
)
def receive_alipay_notify(
    session: Annotated[Session, Depends(get_db)],
    order_id: Annotated[str, Form(min_length=1, max_length=40)],
    trade_no: Annotated[str, Form(min_length=1, max_length=80)],
    total_amount: Annotated[str, Form(min_length=1, max_length=32)],
) -> PaymentCallbackResponse:
    payment = complete_alipay_payment(
        session=session,
        order_id=order_id,
        trade_no=trade_no,
        total_amount=total_amount,
    )

    return PaymentCallbackResponse(
        result="completed",
        payment_id=payment.public_id,
        order_status=payment.order.status,
    )
