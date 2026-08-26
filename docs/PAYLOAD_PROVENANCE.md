# 공급자 payload 구성 근거

과제 payload는 기존 PHP/Python 마이그레이션 코드에서 사용한 입력 경계를 바탕으로
채용 과제에 필요한 필드만 남긴 것입니다.

- Toss: 성공 return에서 `paymentKey`, `orderId`, `amount`를 사용합니다.
- Stripe: `checkout.session.completed`의 Checkout Session `id`,
  `client_reference_id`, `amount_total`, `currency`, `payment_status`를 사용합니다.
- Alipay: form notify의 `order_id`, `trade_no`, `total_amount`를 사용합니다.

실제 요청 원문을 복사한 것이 아니며 모든 식별자·금액은 새로 만든 값입니다. 실제
서명, 헤더, 고객 객체, billing 정보, metadata, 내부 URL은 포함하지 않았습니다.
공급자 사양 전체를 재현하는 것이 목적이 아니므로 과제 범위는 README에 명시된
필드와 통화에 한정합니다.
