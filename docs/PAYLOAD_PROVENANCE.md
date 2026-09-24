# 공급자 payload 구성 근거

- Toss: 성공 return에서 `paymentKey`, `orderId`, `amount`를 사용합니다.
- Stripe: `checkout.session.completed`의 Checkout Session `id`,
  `client_reference_id`, `amount_total`, `currency`, `payment_status`를 사용합니다.
- Alipay: form notify의 `order_id`, `trade_no`, `total_amount`를 사용합니다.

공급자 사양 전체를 재현하는 것이 목적이 아니므로 과제 범위는 README에 명시된 필드와 통화에 한정합니다.
