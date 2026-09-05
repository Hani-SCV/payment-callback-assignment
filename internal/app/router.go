package app

import (
	"net/http"
)

func NewRouter(deps *Dependencies) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc(
		"/health",
		func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("OK"))
		},
	)

	mux.HandleFunc(
		"/v1/payment-callbacks/toss/return",
		deps.PaymentHandler.TossReturn,
	)

	return mux
}