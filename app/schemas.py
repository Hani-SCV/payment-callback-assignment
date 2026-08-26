from typing import Any, Literal

from pydantic import BaseModel, ConfigDict, Field


class TossReturnRequest(BaseModel):
    payment_key: str = Field(alias="paymentKey", min_length=1, max_length=80)
    order_id: str = Field(alias="orderId", min_length=1, max_length=40)
    amount: int = Field(gt=0, strict=True)

    model_config = ConfigDict(extra="forbid")


class StripeWebhookData(BaseModel):
    object: dict[str, Any]

    model_config = ConfigDict(extra="allow")


class StripeWebhookRequest(BaseModel):
    id: str = Field(min_length=1, max_length=80)
    type: str = Field(min_length=1, max_length=80)
    data: StripeWebhookData

    model_config = ConfigDict(extra="allow")


class AlipayNotifyRequest(BaseModel):
    order_id: str = Field(min_length=1, max_length=40)
    trade_no: str = Field(min_length=1, max_length=80)
    total_amount: str = Field(min_length=1, max_length=32)

    model_config = ConfigDict(extra="forbid")


class PaymentCallbackResponse(BaseModel):
    result: Literal["completed", "already_completed"]
    payment_id: str
    order_status: Literal["PAID"]


class ErrorDetail(BaseModel):
    code: str
    message: str


class ErrorResponse(BaseModel):
    error: ErrorDetail
