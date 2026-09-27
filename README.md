# Payment Callback Assignment - Go

다중 결제 서비스의 콜백 처리 로직을 Go로 구현한 Backend Assignment입니다.

Toss, Stripe, Alipay의 서로 다른 콜백 요청 형식을 처리하고, 중복 및 동시 요청에서도 결제와 주문 상태의 일관성을 유지하도록 구현했습니다.

## 주요 구현

* Toss 결제 완료 콜백
* Stripe Checkout webhook
* Alipay notify
* Provider별 요청 형식 및 금액 단위 처리
* 결제 및 주문 상태 검증
* 금액 및 통화 검증
* 중복 콜백 멱등성 처리
* 동시 콜백에 대한 Transaction 및 Row Lock
* Payment Event 기록
* Outbox Message 기록
* 오류 발생 시 Transaction Rollback

## API

| Provider | Endpoint | Content-Type |
| -------- | -------- | ------------ |
| Toss | `POST /v1/payment-callbacks/toss/return` | JSON |
| Stripe | `POST /v1/payment-callbacks/stripe/webhook` | JSON |
| Alipay | `POST /v1/payment-callbacks/alipay/notify` | Form URL Encoded |

## Architecture

```text
HTTP Request
     ↓
Handler
     ↓
Service
     ↓
Repository
     ↓
GORM
     ↓
MySQL
```

Handler는 HTTP 요청과 응답을 처리하고, Service에서 결제 콜백의 비즈니스 로직과 Transaction을 관리합니다.

Repository는 DB 조회 및 변경을 담당하며, Transaction 내부에서는 동일한 Transaction DB를 사용합니다.

## Transaction

하나의 결제 완료 과정에서 다음 작업을 하나의 Transaction으로 처리합니다.

```text
Payment 상태 변경
      +
Order 상태 변경
      +
Payment Event 생성
      +
Outbox Message 생성
```

처리 중 오류가 발생하면 전체 변경사항을 Rollback하여 일부 데이터만 변경된 상태가 남지 않도록 합니다.

## Concurrency Control

동일한 결제에 대한 동시 콜백을 처리할 때 DB Row Lock을 사용합니다.

```text
Callback A ──┐
             ├── Payment Row Lock
Callback B ──┘
```

Payment와 Order를 Transaction 내부에서 조회하고 Row Lock을 획득한 뒤 상태를 검증하고 변경합니다.

이를 통해 동시에 들어온 콜백이 동일한 결제를 중복 완료하거나 거래 ID를 덮어쓰는 것을 방지합니다.

## Idempotency

이미 완료된 Payment에 동일한 PG 거래 ID가 다시 전달되면 기존 완료 상태를 유지하고 성공 응답을 반환합니다.

반대로 이미 완료된 Payment에 다른 거래 ID가 전달되는 경우에는 기존 거래 정보를 보호하기 위해 충돌로 처리합니다.

```text
같은 transaction ID
    → 이미 처리된 요청
    → 200 OK

다른 transaction ID
    → 기존 거래와 충돌
    → Error
```

## Outbox Pattern

결제 완료와 후속 이벤트 발행을 분리하기 위해 Outbox Message를 Transaction 안에서 함께 생성합니다.

```text
Payment 완료
    +
Payment Event 생성
    +
Outbox Message 생성
        ↓
      COMMIT
```

결제 상태는 변경되었지만 이벤트 기록이 누락되는 상황을 방지할 수 있도록 구성했습니다.

## Project Structure

```text
cmd/
└── server/
    └── main.go

internal/
├── app/
│   ├── app.go
│   ├── dependencies.go
│   └── router.go
├── config/
├── database/
├── errors/
├── handler/
├── model/
├── repository/
├── request/
├── response/
└── service/

tests/
├── payment_callback_test.go
└── test_helper.go
```

## Tech Stack

* Go 1.26
* net/http
* GORM
* MySQL 8.4
* Docker
* Testify

## Test

전체 테스트 실행:

```bash
go test ./...
```

상세 테스트:

```bash
go test ./tests -v
```

Race Detector:

```bash
go test ./... -race
```

## Run

DB 실행:

```bash
docker compose up -d
```

서버 실행:

```bash
go run ./cmd/server
```

## 원본 과제

전체 과제 요구사항과 참고 자료는 [`docs/ASSIGNMENT.md`](docs/ASSIGNMENT.md)에서 확인할 수 있습니다.
