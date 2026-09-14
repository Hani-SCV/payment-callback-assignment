package handler

import (
	"encoding/json"
	stderrors "errors"
	"net/http"

	"github.com/Hani-SCV/payment-callback-assignment/internal/errors"
	"github.com/Hani-SCV/payment-callback-assignment/internal/model"
	"github.com/Hani-SCV/payment-callback-assignment/internal/service"
)

type PaymentCallbackHandler struct {
	service *service.PaymentCallbackService
}

func NewPaymentCallbackHandler(service *service.PaymentCallbackService) *PaymentCallbackHandler {
	return &PaymentCallbackHandler{
		service: service,
	}
}

func (h *PaymentCallbackHandler) TossReturn(w http.ResponseWriter, r *http.Request) {
	var req model.TossReturnRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusInternalServerError)
		return
	}

	if err := h.service.ProcessTossReturn(r.Context(), req); err != nil {
		if stderrors.Is(err, errors.ErrInvalidProvider) ||
			stderrors.Is(err, errors.ErrInvalidPaymentStatus) ||
			stderrors.Is(err, errors.ErrInvalidOrderStatus) ||
			stderrors.Is(err, errors.ErrInvalidAmount) ||
			stderrors.Is(err, errors.ErrInvalidCurrency) {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(map[string]string{
		"message": "received",
	})
}

