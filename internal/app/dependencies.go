package app

import (
	"github.com/Hani-SCV/payment-callback-assignment/internal/database"
	"github.com/Hani-SCV/payment-callback-assignment/internal/handler"
	"github.com/Hani-SCV/payment-callback-assignment/internal/repository"
	"github.com/Hani-SCV/payment-callback-assignment/internal/service"
	"gorm.io/gorm"
)

type Dependencies struct {
	PaymentHandler *handler.PaymentCallbackHandler
}

func NewDependencies(db *gorm.DB) *Dependencies {
	transactionManager := database.NewTransactionManager(db)
	orderRepository := repository.NewOrderRepository(db)
	paymentRepository := repository.NewPaymentRepository(db)
	paymentEventRepository := repository.NewPaymentEventRepository(db)
	outboxMessageRepository := repository.NewOutboxMessageRepository(db)

	paymentService := service.NewPaymentCallbackService(
		transactionManager,
		orderRepository,
		paymentRepository,
		paymentEventRepository,
		outboxMessageRepository,
	)

	paymentHandler := handler.NewPaymentCallbackHandler(
		paymentService,
	)

	return &Dependencies{
		PaymentHandler: paymentHandler,
	}
}