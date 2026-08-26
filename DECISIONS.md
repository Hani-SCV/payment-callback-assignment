# Decisions

## Provider별 입력 변환 및 공통 완료 처리

Toss, Stripe, Alipay는 callback 형식과 금액 단위가 다르므로 각 provider의 입력을 service 계층에서 내부 결제 처리에 필요한 형태로 변환한 후 공통적인 완료 처리 로직을 수행하도록 구성했습니다.

금액은 Decimal을 사용하며, provider별 금액 단위를 내부 Payment 금액 단위로 변환한 뒤 금액과 통화를 함께 검증합니다.

* Toss: KRW 정수 금액
* Stripe: 최소 통화 단위 → 내부 금액 단위로 변환
* Alipay: 금액 문자열 → Decimal 변환

## 트랜잭션 제어 및 동시성 제어

Payment 상태 변경, Order 상태 변경, PaymentEvent 생성, OutboxMessage 생성을 하나의 DB transaction으로 처리했습니다.

결제 완료 과정에서는 Payment와 해당 Order를 row-level lock으로 조회합니다. 특히 Order를 잠금으로써 하나의 주문에 여러 Payment가 연결되어 있더라도 동시에 두 Payment가 완료되는 것을 방지했습니다.

먼저 Order를 완료한 요청만 결제를 완료할 수 있으며, 이미 PAID 상태인 주문에 연결된 다른 Payment는 완료 처리하지 않습니다.

오류가 발생하면 전체 transaction을 rollback하여 일부 데이터만 변경되는 상황을 방지합니다.

## 중복 처리 및 거래 식별

중복 callback은 동일 이벤트와 동일 거래를 구분하여 처리합니다.

Stripe는 event_id를 기준으로 이미 처리된 이벤트인지 확인하고, Toss와 Alipay는 Payment의 external_transaction_id를 기준으로 동일 거래의 재전송 여부를 확인합니다.

동일 거래의 재전송은 기존 데이터를 다시 생성하거나 변경하지 않고 성공으로 응답합니다.

반면 이미 완료된 Payment에 다른 거래 ID가 들어오거나, Stripe event가 다른 Payment에 연결되어 있는 경우에는 오류로 처리하여 기존 결제 정보가 다른 거래로 덮어써지는 것을 방지했습니다.

## 결제 완료 기록

결제 완료 시 PaymentEvent와 OutboxMessage를 각각 한 건 생성합니다.

두 데이터는 결제 완료에 필요한 최소한의 정보만 저장하며 callback 원문 전체나 과제 처리에 필요하지 않은 provider 필드는 저장하지 않습니다.

Event와 Outbox 생성 역시 Payment 및 Order 변경과 동일한 transaction에서 처리하여 결제 완료 상태와 기록 간의 불일치를 방지했습니다.

## 다른 결제 시도 정리

하나의 주문에 여러 Payment가 존재할 수 있으므로 하나의 Payment가 완료되면 같은 주문의 다른 미완료 Payment는 다시 완료 가능한 상태로 남지 않도록 처리했습니다.

이를 통해 실패 후 다른 PG로 재시도한 결제나 금액이 서로 다른 결제 시도가 동일 주문에서 중복 완료되는 것을 방지했습니다.

