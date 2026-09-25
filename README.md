# Payment Callback Assignment

다중 결제 서비스의 콜백 처리 로직을 구현한 Backend Assignment입니다.

Toss, Stripe, Alipay의 서로 다른 콜백 요청 형식을 처리하고, 결제 및 주문 상태를 안전하게 변경하도록 구현했습니다.

## 구현 범위

FastAPI를 사용해 다음 결제 Provider의 콜백 처리를 구현했습니다.

* Toss 결제 완료 콜백
* Stripe Checkout webhook
* Alipay notify

## 주요 구현

* 결제 및 주문 상태 검증
* 금액 및 통화 검증
* 중복 콜백 멱등성 처리
* 동시 콜백에 대한 트랜잭션 및 Row Lock
* Payment Event 기록
* Outbox Message 기록
* 오류 발생 시 트랜잭션 롤백
* Provider별 콜백 요청 형식 처리

## API

| Provider | Endpoint | Content-Type |
| -------- | -------- | ------------ |
| Toss | `POST /v1/payment-callbacks/toss/return` | JSON |
| Stripe | `POST /v1/payment-callbacks/stripe/webhook` | JSON |
| Alipay | `POST /v1/payment-callbacks/alipay/notify` | Form URL Encoded |

## Architecture

FastAPI 구현에서는 Repository 계층을 별도로 두지 않고, Service 계층에서 SQLAlchemy Session을 직접 사용해 DB 조회 및 변경을 처리합니다.

```text
Router
   ↓
Service
   ↓
SQLAlchemy Session
   ↓
MySQL
```

트랜잭션은 Service 계층에서 관리하며, 하나의 콜백 처리 과정에서 Payment, Order, PaymentEvent, Outbox를 하나의 트랜잭션으로 처리합니다.

동시 콜백에 대해서는 DB Row Lock을 사용해 동일한 결제에 대한 상태 변경을 안전하게 처리합니다.

## Tech Stack

* Python
* FastAPI
* SQLAlchemy
* MySQL
* Docker
* Pytest

## 테스트

```bash
docker compose run --rm api pytest -q
```

## 원본 과제

전체 과제 요구사항과 참고 자료는 [`docs/ASSIGNMENT.md`](docs/ASSIGNMENT.md)에서 확인할 수 있습니다.
