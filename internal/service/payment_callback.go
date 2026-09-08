package service

import (
	"context"

	"encoding/json"

	"github.com/Hani-SCV/payment-callback-assignment/internal/database"
	"github.com/Hani-SCV/payment-callback-assignment/internal/errors"
	"github.com/Hani-SCV/payment-callback-assignment/internal/model"
	"github.com/Hani-SCV/payment-callback-assignment/internal/repository"
	"gorm.io/gorm"
)

type PaymentCallbackService struct {
	transactionManager      *database.TransactionManager
	orderRepository         *repository.OrderRepository
	paymentRepository       *repository.PaymentRepository
	paymentEventRepository  *repository.PaymentEventRepository
	outboxMessageRepository *repository.OutboxMessageRepository
}

type PaymentEventPayload struct {
	Provider      string `json:"provider"`
	TransactionID string `json:"transaction_id"`
	Amount        int64  `json:"amount"`
	Currency      string `json:"currency"`
}

type OutboxMessagePayload struct {
	Provider      string `json:"provider"`
	PaymentID     string `json:"public_id"`
	TransactionID string `json:"transaction_id"`
	Amount        int64  `json:"amount"`
	Currency      string `json:"currency"`
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

func createPaymentEvent(
	ctx context.Context,
	repo *repository.PaymentEventRepository,
	payment *model.Payment,
	transactionID string,
	amount int64,
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
	amount int64,
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
		Status:           "PENDING",
	}

	return repo.Create(ctx, event)
}

func (s *PaymentCallbackService) ProcessTossReturn(
	ctx context.Context,
	req model.TossReturnRequest,
) error {
	return s.transactionManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		orderRepository := s.orderRepository.WithTx(tx)
		paymentRepository := s.paymentRepository.WithTx(tx)
		paymentEventRepository := s.paymentEventRepository.WithTx(tx)
		outboxMessageRepository := s.outboxMessageRepository.WithTx(tx)

		// Payment 잠금
		payment, err := paymentRepository.FindByPublicIDForUpdate(
			ctx,
			req.OrderID,
		)
		if err != nil {
			return err
		}

		// Order 잠금
		order, err := orderRepository.FindByIDForUpdate(
			ctx,
			payment.OrderID,
		)
		if err != nil {
			return err
		}

		// TODO: 결제 제공자 검증
		if payment.Provider != "TOSS" {
			return errors.ErrInvalidProvider
		}
		// TODO: 이미 처리된 동일 콜백인지 확인
		if payment.Status == "COMPLETED" {
			return errors.ErrInvalidPaymentStatus
		}

		// TODO: 처리 가능한 결제 상태인지 확인
		if payment.Status != "PENDING" {
			return errors.ErrInvalidPaymentStatus
		}

		// TODO: 주문 상태 검증
		if order.Status != "PAYMENT_PENDING" {
			return errors.ErrInvalidOrderStatus
		}

		// TODO: 금액 검증
		if payment.Amount != req.Amount {
			return errors.ErrInvalidAmount
		}

		// TODO: 통화 검증
		if payment.Currency != "KRW" {
			return errors.ErrInvalidCurrency
		}

		// TODO: 결제 완료
		if err := paymentRepository.Complete(ctx, payment.ID); err != nil {
			return errors.ErrInvalidPaymentStatus
		}

		// TODO: 주문 완료
		if err := orderRepository.MarkAsPaid(ctx, order.ID); err != nil {
			return errors.ErrInvalidOrderStatus
		}

		// TODO: PaymentEvent 저장
		if err := createPaymentEvent(
			ctx,
			paymentEventRepository,
			payment,
			req.PaymentKey,
			req.Amount,
			payment.Currency,
		); err != nil {
			return err
		}

		// TODO: Outbox 저장
		if err := createOutboxMessage(
			ctx,
			outboxMessageRepository,
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
