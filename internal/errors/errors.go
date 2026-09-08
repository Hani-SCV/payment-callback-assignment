package errors

type AppError struct {
	Code       string
	Message    string
	StatusCode int
}

func (e *AppError) Error() string {
	return e.Message
}

var (
	ErrInvalidProvider = &AppError{
		Code:       "INVALID_PROVIDER",
		Message:    "invalid payment provider",
		StatusCode: 422,
	}

	ErrInvalidAmount = &AppError{
		Code:       "INVALID_AMOUNT",
		Message:    "invalid payment amount",
		StatusCode: 422,
	}

	ErrInvalidCurrency = &AppError{
		Code:       "INVALID_CURRENCY",
		Message:    "invalid payment currency",
		StatusCode: 422,
	}

	ErrInvalidPaymentStatus = &AppError{
		Code:       "INVALID_PAYMENT_STATUS",
		Message:    "invalid payment status",
		StatusCode: 409,
	}

	ErrInvalidOrderStatus = &AppError{
		Code:       "INVALID_ORDER_STATUS",
		Message:    "invalid order status",
		StatusCode: 409,
	}
)