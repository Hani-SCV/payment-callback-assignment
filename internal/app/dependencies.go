package app

import (
	"github.com/Hani-SCV/payment-callback-assignment/internal/config"
	"github.com/Hani-SCV/payment-callback-assignment/internal/database"
	"github.com/Hani-SCV/payment-callback-assignment/internal/handler"
	"github.com/Hani-SCV/payment-callback-assignment/internal/repository"
	"github.com/Hani-SCV/payment-callback-assignment/internal/service"
)

type Dependencies struct {
	PaymentHandler *handler.PaymentCallbackHandler
}

func NewDependencies(cfg *config.Config) (*Dependencies, error) {
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	transactionManager := database.NewTransactionManager(db)
	orderRepository := repository.NewOrderRepository(db)
	paymentRepository := repository.NewPaymentRepository(db)

	paymentService := service.NewPaymentCallbackService(
		transactionManager,
		orderRepository,
		paymentRepository,
	)

	paymentHandler := handler.NewPaymentCallbackHandler(
		paymentService,
	)

	return &Dependencies{
		PaymentHandler: paymentHandler,
	}, nil
}