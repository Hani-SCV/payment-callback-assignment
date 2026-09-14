package service

import (
	"context"

	"encoding/json"

	"github.com/Hani-SCV/payment-callback-assignment/internal/database"
	"github.com/Hani-SCV/payment-callback-assignment/internal/errors"
	"github.com/Hani-SCV/payment-callback-assignment/internal/model"
	"github.com/Hani-SCV/payment-callback-assignment/internal/repository"
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
		return errors.ErrInvalidPaymentStatus
	}

	event := &model.PaymentEvent{
		PaymentID: payment.ID,
		EventID:   &transactionID,
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
	req model.TossReturnRequest,
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
