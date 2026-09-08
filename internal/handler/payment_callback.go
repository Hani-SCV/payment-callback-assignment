package handler

import (
	"encoding/json"
	"net/http"

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
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if err := h.service.ProcessTossReturn(r.Context(), req); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(map[string]string{
		"message": "received",
	})
}