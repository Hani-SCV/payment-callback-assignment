# Backend Assignment: 다중 결제 콜백 마이그레이션

## 상황

PHP에 있던 결제 완료 처리를 Python 백엔드로 옮기려 합니다. 
Toss, Stripe, Alipay는 요청 형식과 금액 단위가 서로 다릅니다. 
세 경로로 콜백이 들어와도 주문과 결제 기록이 꼬이지 않게 구현해 주세요.

이 저장소의 데이터는 모두 과제를 위해 만든 값입니다. 실제 회사명, URL, 주문 번호, 고객 정보는 들어 있지 않습니다.

## 구현할 것

다음 세 endpoint와 관련 테스트를 완성해 주세요.

- `POST /v1/payment-callbacks/toss/return` — JSON
- `POST /v1/payment-callbacks/stripe/webhook` — JSON
- `POST /v1/payment-callbacks/alipay/notify` — form-urlencoded

endpoint 경로와 `orders`, `payments`, `payment_events`, `outbox_messages` 테이블은 채점에 사용합니다. 
이 부분을 제외한 파일 구조와 내부 모델은 바꿔도 됩니다.

현재 `app/services` 구조를 그대로 따를 필요도 없습니다.
스키마를 바꾼다면 `db/migrations/`에 변경 SQL을 추가하고, `db/init/001_schema.sql`과 ORM 모델에도 같은 내용을 반영해 주세요.

## 참고 자료

- [legacy/payment_callback.php](legacy/payment_callback.php) — 기존 PHP 코드
- [docs/REQUIREMENTS_MEMO.md](docs/REQUIREMENTS_MEMO.md) — 제품팀 메모
- [docs/INCIDENT_2025_11.md](docs/INCIDENT_2025_11.md) — 이전 장애 기록
- [PRIVACY.md](PRIVACY.md) — 저장하거나 로그로 남기면 안 되는 데이터
- [fixtures](fixtures) — 요청 예시

## 요청 예시

### Toss return

```json
{
  "paymentKey": "toss_key_demo_1001",
  "orderId": "pay_demo_toss_001",
  "amount": 129900
}
```

- `orderId`: 내부 결제 참조
- `paymentKey`: PG 거래 참조
- `amount`: KRW 정수 금액

### Stripe Checkout webhook

```json
{
  "id": "evt_demo_1001",
  "type": "checkout.session.completed",
  "data": {
    "object": {
      "id": "cs_demo_1001",
      "client_reference_id": "pay_demo_stripe_001",
      "amount_total": 7700,
      "currency": "usd",
      "payment_status": "paid"
    }
  }
}
```

- `client_reference_id`: 내부 결제 참조
- Checkout Session `id`: PG 거래 참조
- `amount_total`: Stripe 최소 통화 단위
- 과제 데이터에는 KRW, JPY, USD, CNY가 나올 수 있습니다.

### Alipay notify

```text
order_id=pay_demo_alipay_001&trade_no=alipay_trade_demo_1001&total_amount=88.80
```

- `order_id`: 내부 결제 참조
- `trade_no`: PG 거래 참조
- `total_amount`: CNY 금액 문자열

위 예시는 과제에서 쓰는 필드만 추린 것입니다. 실제 요청에는 다른 필드가 더 들어올 수 있습니다.

## 이번 과제에서 가정하는 것

아래 검증은 endpoint에 도달하기 전에 끝났다고 가정합니다.

- Toss 승인 또는 조회 결과 확인
- Stripe 원문 body와 `Stripe-Signature` 검증
- Alipay 결제 브리지의 RSA 검증

SDK 호출이나 서명 검증은 구현하지 않아도 됩니다.

## 완료 기준

- 정상 콜백을 처리하면 payment와 order가 완료되고, 완료 기록과 발행 전 outbox가 각각 한 건 남아야 합니다.
- 재전송이나 서로 경쟁하는 콜백 때문에 거래 정보가 덮어써지거나, 같은 일이 두 번 기록되거나, 한 주문의 결제가 둘 이상 완료되면 안 됩니다. 
  더 이상 사용할 수 없는 결제 시도도 다시 완료 가능한 상태로 남겨 두지 마세요.
- 요청 검증이나 저장 중 오류가 나면 관련 데이터가 일부만 바뀌어 있으면 안 됩니다. 예상할 수 있는 잘못된 요청과 상태 충돌은 HTTP 500이 아닌 응답으로 처리해 주세요.
- 금액은 통화 단위까지 정확히 비교하고 DB 시각은 UTC로 저장합니다. 계산에 `float`를 사용하거나 callback 원문과 필요 없는 필드를 DB나 로그에 남기면 안 됩니다.

위 조건은 순서대로 들어온 요청과 동시에 들어온 요청에서 모두 같아야 합니다.

최초 완료 응답은 HTTP 200과 아래 JSON을 사용합니다.

```json
{
  "result": "completed",
  "payment_id": "pay_demo_toss_001",
  "order_status": "PAID"
}
```

같은 거래의 재전송은 데이터를 다시 쓰지 않고 HTTP 200으로 응답해 주세요. 
응답의 `result`에는 `"already_completed"`를 사용할 수 있습니다. 
오류 응답의 바깥 형태는 `{"error":{"code":"ERROR_CODE","message":"설명"}}`로 맞춰 주세요. 
HTTP status와 세부 code는 구현하면서 정하면 됩니다. 
FastAPI validation 응답은 기본 형식을 써도 됩니다.

## 실행

Docker와 Docker Compose가 필요합니다.

```bash
docker compose up --build
docker compose run --rm api pytest -q
```

- API 문서: `http://localhost:8000/docs`
- MySQL: `localhost:53306`
- API: `localhost:8000`

처음 받은 상태에서는 health check만 통과하고 세 callback은 HTTP 501을 반환합니다.
`make reset`은 이 Compose 프로젝트의 로컬 DB 볼륨을 삭제하고 시드를 다시 만듭니다.

## 필수 제출물

- 구현한 코드 및 테스트
- 구현 중 내린 결정을 정리한 `DECISIONS.md`
- 스키마를 변경했다면, `db/migrations/` 아래 변경 SQL

PG 호출, 서명 검증, 사용자 인증, outbox publisher, 배포 설정은 범위에 포함하지 않습니다. 
사용한 도구의 기록도 제출하지 않아도 됩니다.

## 제출 방법

1. 회사의 공개 저장소를 로컬에 clone해 작업합니다.
2. 완성한 코드나 commit은 공개 fork에 push하지 말고, 회사 저장소에도 pull request를 만들지 마세요.
3. 검사가 끝나면 소스, 테스트, 문서를 포함한 프로젝트 폴더 전체를 ZIP으로 압축해 안내받은 이메일 주소로 보내 주세요. 
   `.git`, `.env`, DB 볼륨, 캐시, 비밀값, 실제 사용자 데이터는 포함하지 마세요.