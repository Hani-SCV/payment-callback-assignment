from typing import Annotated

from fastapi import APIRouter, Depends, Form
from sqlalchemy.orm import Session

from app.db import get_db
from app.errors import AssignmentIncompleteError
from app.schemas import (
    AlipayNotifyRequest,
    ErrorResponse,
    PaymentCallbackResponse,
    StripeWebhookRequest,
    TossReturnRequest,
)

router = APIRouter(prefix="/v1/payment-callbacks", tags=["payment-callbacks"])

ERROR_RESPONSES = {
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
    del payload, session
    raise AssignmentIncompleteError()


@router.post(
    "/stripe/webhook",
    response_model=PaymentCallbackResponse,
    responses=ERROR_RESPONSES,
)
def receive_stripe_webhook(
    payload: StripeWebhookRequest,
    session: Annotated[Session, Depends(get_db)],
) -> PaymentCallbackResponse:
    del payload, session
    raise AssignmentIncompleteError()


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
    payload = AlipayNotifyRequest(
        order_id=order_id,
        trade_no=trade_no,
        total_amount=total_amount,
    )
    del payload, session
    raise AssignmentIncompleteError()
