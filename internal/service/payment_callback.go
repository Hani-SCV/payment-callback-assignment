package service

import (
	"context"
	stderrors "errors"

	"encoding/json"

	"github.com/Hani-SCV/payment-callback-assignment/internal/database"
	"github.com/Hani-SCV/payment-callback-assignment/internal/errors"
	"github.com/Hani-SCV/payment-callback-assignment/internal/model"
	"github.com/Hani-SCV/payment-callback-assignment/internal/repository"
	"github.com/Hani-SCV/payment-callback-assignment/internal/request"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type PaymentCallbackService struct {
	transactionManager      *database.TransactionManager
	orderRepository         *repository.OrderRepository
	paymentRepository       *repository.PaymentRepository
	paymentEventRepository  *repository.PaymentEventRepository
	outboxMessageRepository *repository.OutboxMessageRepository
}

type Repositories struct {
	Order        *repository.OrderRepository
	Payment      *repository.PaymentRepository
	PaymentEvent *repository.PaymentEventRepository
	Outbox       *repository.OutboxMessageRepository
}

type PaymentEventPayload struct {
	Provider      string          `json:"provider"`
	TransactionID string          `json:"transaction_id"`
	Amount        decimal.Decimal `json:"amount"`
	Currency      string          `json:"currency"`
}

type OutboxMessagePayload struct {
	Provider      string          `json:"provider"`
	PaymentID     string          `json:"public_id"`
	TransactionID string          `json:"transaction_id"`
	Amount        decimal.Decimal `json:"amount"`
	Currency      string          `json:"currency"`
}

func NewPaymentCallbackService(
	transactionManager *database.TransactionManager,
	orderRepository *repository.OrderRepository,
	paymentRepository *repository.PaymentRepository,
	paymentEventRepository *repository.PaymentEventRepository,
	outboxMessageRepository *repository.OutboxMessageRepository,
) *PaymentCallbackService {
	return &PaymentCallbackService{
		transactionManager:      transactionManager,
		orderRepository:         orderRepository,
		paymentRepository:       paymentRepository,
		paymentEventRepository:  paymentEventRepository,
		outboxMessageRepository: outboxMessageRepository,
	}
}

func (s *PaymentCallbackService) repositoriesWithTx(tx *gorm.DB) *Repositories {
	return &Repositories{
		Order:        s.orderRepository.WithTx(tx),
		Payment:      s.paymentRepository.WithTx(tx),
		PaymentEvent: s.paymentEventRepository.WithTx(tx),
		Outbox:       s.outboxMessageRepository.WithTx(tx),
	}
}

func createPaymentEvent(
	ctx context.Context,
	repo *repository.PaymentEventRepository,
	payment *model.Payment,
	eventID string,
	transactionID string,
	amount decimal.Decimal,
	currency string,
) error {
	payload := PaymentEventPayload{
		Provider:      payment.Provider,
		TransactionID: transactionID,
		Amount:        amount,
		Currency:      currency,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	event := &model.PaymentEvent{
		PaymentID: payment.ID,
		EventID:   &eventID,
		EventType: "PAYMENT_COMPLETED",
		Payload:   jsonData,
	}

	return repo.Create(ctx, event)
}

func createOutboxMessage(
	ctx context.Context,
	repo *repository.OutboxMessageRepository,
	payment *model.Payment,
	transactionID string,
	amount decimal.Decimal,
	currency string,
) error {
	payload := OutboxMessagePayload{
		Provider:      payment.Provider,
		PaymentID:     payment.PublicID,
		TransactionID: transactionID,
		Amount:        amount,
		Currency:      currency,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return errors.ErrInvalidPaymentStatus
	}

	event := &model.OutboxMessage{
		DeduplicationKey: payment.PublicID,
		EventType:        "PAYMENT_COMPLETED",
		AggregateType:    "payment",
		AggregateID:      payment.PublicID,
		Payload:          jsonData,
		Status:           model.PaymentStatusPending,
	}

	return repo.Create(ctx, event)
}

func (s *PaymentCallbackService) ProcessTossReturn(
	ctx context.Context,
	req request.TossReturnRequest,
) error {
	return s.transactionManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		repos := s.repositoriesWithTx(tx)

		// Payment 잠금
		payment, err := repos.Payment.FindByPublicIDForUpdate(
			ctx,
			req.OrderID,
		)
		if err != nil {
			return err
		}

		// Order 잠금
		order, err := repos.Order.FindByIDForUpdate(
			ctx,
			payment.OrderID,
		)
		if err != nil {
			return err
		}

		// 결제 제공자 검증
		if payment.Provider != "TOSS" {
			return errors.ErrInvalidProvider
		}

		// 이미 처리된 결제 검증 (중복 콜백은 정상 처리한다.)
		if payment.Status == model.PaymentStatusCompleted {
			if payment.ExternalTransactionID != nil &&
				*payment.ExternalTransactionID == req.PaymentKey {
				return nil
			}
			return errors.ErrInvalidPaymentStatus
		}

		// 처리 가능한 결제 상태인지 확인
		if payment.Status != model.PaymentStatusPending {
			return errors.ErrInvalidPaymentStatus
		}

		// 주문 상태 검증
		if order.Status != "PAYMENT_PENDING" {
			return errors.ErrInvalidOrderStatus
		}

		// 금액 검증
		if !payment.Amount.Equal(req.Amount) {
			return errors.ErrInvalidAmount
		}

		// 통화 검증
		if payment.Currency != "KRW" {
			return errors.ErrInvalidCurrency
		}

		// 결제 완료
		if err := repos.Payment.Complete(
			ctx,
			payment.ID,
			req.PaymentKey,
		); err != nil {
			return errors.ErrInvalidPaymentStatus
		}

		// 주문 완료
		if err := repos.Order.MarkAsPaid(ctx, order.ID); err != nil {
			return errors.ErrInvalidOrderStatus
		}

		// PaymentEvent 저장
		if err := createPaymentEvent(
			ctx,
			repos.PaymentEvent,
			payment,
			req.PaymentKey,
			req.PaymentKey,
			req.Amount,
			payment.Currency,
		); err != nil {
			return err
		}

		// Outbox 저장
		if err := createOutboxMessage(
			ctx,
			repos.Outbox,
			payment,
			req.PaymentKey,
			req.Amount,
			payment.Currency,
		); err != nil {
			return err
		}

		return nil
	})
}

func (s *PaymentCallbackService) ProcessStripeWebhook(
	ctx context.Context,
	req request.StripeWebhookRequest,
) error {
	return s.transactionManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		repos := s.repositoriesWithTx(tx)

		// data.object에서 결제 정보 추출
		object := req.Data.Object
		transactionID := object.ID
		paymentID := object.ClientReferenceID
		amount := decimal.NewFromInt(object.AmountTotal)
		currency := object.Currency
		paymentStatus := object.PaymentStatus

		// Stripe webhook event ID로 중복 이벤트 여부 확인
		_, err := repos.PaymentEvent.FindByEventID(ctx, req.ID)
		if err == nil {
			return nil
		}

		if !stderrors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		// Stripe event type 확인
		if req.Type != "checkout.session.completed" {
			return errors.ErrInvalidEventType
		}

		// 결제 ID로 Payment 조회 및 row lock
		payment, err := repos.Payment.FindByPublicIDForUpdate(ctx, paymentID)
		if err != nil {
			return err
		}

		// Payment에 연결된 Order 조회 및 row lock
		order, err := repos.Order.FindByIDForUpdate(ctx, payment.OrderID)
		if err != nil {
			return err
		}

		// Provider가 STRIPE인지 검증
		if payment.Provider != "STRIPE" {
			return errors.ErrInvalidProvider
		}

		// 결제 상태 검증
		// 이미 완료된 결제인데 새로운 이벤트가 들어온 경우
		if payment.Status == model.PaymentStatusCompleted {
			if payment.ExternalTransactionID != nil &&
				*payment.ExternalTransactionID == transactionID {
				return nil
			}
			return errors.ErrInvalidPaymentStatus
		}

		// PENDING 상태가 아닌지 체크
		if payment.Status != model.PaymentStatusPending {
			return errors.ErrInvalidPaymentStatus
		}

		// Order 상태 검증
		if order.Status != "PAYMENT_PENDING" {
			return errors.ErrInvalidOrderStatus
		}

		// Stripe 결제 금액과 Payment 금액 비교
		if !payment.Amount.Equal(amount) {
			return errors.ErrInvalidAmount
		}

		// Stripe 통화와 Payment 통화 비교
		if payment.Currency != currency {
			return errors.ErrInvalidCurrency
		}

		if paymentStatus != "paid" {
			return errors.ErrInvalidPaymentStatus
		}

		// 결제 완료 처리
		if err := repos.Payment.Complete(ctx, payment.ID, transactionID); err != nil {
			return errors.ErrInvalidPaymentStatus
		}

		// Order PAID 처리
		if err := repos.Order.MarkAsPaid(ctx, order.ID); err != nil {
			return errors.ErrInvalidOrderStatus
		}

		// PaymentEvent 생성
		if err := createPaymentEvent(
			ctx,
			repos.PaymentEvent,
			payment,
			req.ID,
			transactionID,
			amount,
			currency,
		); err != nil {
			return err
		}

		// OutboxMessage 생성
		if err := createOutboxMessage(ctx, repos.Outbox, payment, transactionID, amount, currency); err != nil {
			return err
		}

		return nil
	})
}
