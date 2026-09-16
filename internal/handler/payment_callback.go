package handler

import (
	"encoding/json"
	stderrors "errors"
	"net/http"

	"github.com/Hani-SCV/payment-callback-assignment/internal/errors"
	"github.com/Hani-SCV/payment-callback-assignment/internal/request"
	"github.com/Hani-SCV/payment-callback-assignment/internal/response"
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
	var req request.TossReturnRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request")
		return
	}

	if err := h.service.ProcessTossReturn(r.Context(), req); err != nil {
		var appErr *errors.AppError

		if stderrors.As(err, &appErr) {
			writeError(
				w,
				appErr.StatusCode,
				appErr.Code,
				appErr.Message,
			)
			return
		}

		writeError(
			w,
			http.StatusInternalServerError,
			"INTERNAL_SERVER_ERROR",
			"internal server error",
		)
		return
	}

	writeJSON(w, http.StatusOK, response.PaymentCallbackResponse{
		Result:      "completed",
		PaymentID:   req.OrderID,
		OrderStatus: "PAID",
	})
}

func (h *PaymentCallbackHandler) StripeWebhook(w http.ResponseWriter, r *http.Request) {
	var req request.StripeWebhookRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request")
		return
	}

	// TODO: service.ProcessStripeWebhook(...)

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "received",
	})
}

func (h *PaymentCallbackHandler) AlipayNotify(w http.ResponseWriter, r *http.Request) {
	// TODO: form parsing
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, response.ErrorResponse{
		Error: response.ErrorDetail{
			Code:    code,
			Message: message,
		},
	})
}
