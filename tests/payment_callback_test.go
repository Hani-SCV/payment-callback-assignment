package tests

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Hani-SCV/payment-callback-assignment/internal/app"
	"github.com/Hani-SCV/payment-callback-assignment/internal/config"
	"github.com/Hani-SCV/payment-callback-assignment/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const TossReturnPath = "/v1/payment-callbacks/toss/return"

func setupTest(t *testing.T) (*gorm.DB, *http.ServeMux) {
	t.Helper()

	cfg := config.Load()

	db, err := gorm.Open(
		mysql.Open(cfg.DatabaseURL),
		&gorm.Config{},
	)
	require.NoError(t, err)

	resetTestDB(t, db)
	seedSyntheticData(t, db)

	deps, err := app.NewDependencies(cfg)
	require.NoError(t, err)

	router := app.NewRouter(deps)

	return db, router
}

func setupTestRequest(body string) (*http.Request, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(
		http.MethodPost,
		TossReturnPath,
		bytes.NewBufferString(body),
	)

	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	return req, rec
}

func TestTossReturnCompletesPayment(t *testing.T) {
	db, router := setupTest(t)

	body := `{
		"paymentKey": "toss_key_demo_1001",
		"orderId": "pay_demo_toss_001",
		"amount": 129900
	}`

	req, rec := setupTestRequest(body)

	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var payment model.Payment
	require.NoError(t,
		db.Where("public_id = ?", "pay_demo_toss_001").
			First(&payment).Error,
	)

	assert.Equal(t, "COMPLETED", payment.Status)

	var order model.Order
	require.NoError(t,
		db.Where("public_id = ?", "ord_demo_1001").
			First(&order).Error,
	)

	assert.Equal(t, "PAID", order.Status)

	var event model.PaymentEvent
	require.NoError(t,
		db.Where("payment_id = ?", payment.ID).
			First(&event).Error,
	)

	assert.Equal(t, "PAYMENT_COMPLETED", event.EventType)

	var outbox model.OutboxMessage
	require.NoError(t,
		db.Where("aggregate_id = ?", payment.PublicID).
			First(&outbox).Error,
	)

	assert.Equal(t, "PAYMENT_COMPLETED", outbox.EventType)
}

func TestTossReturnIsIdempotent(t *testing.T) {
	db, router := setupTest(t)

	body := `{
		"paymentKey": "toss_key_demo_1001",
		"orderId": "pay_demo_toss_001",
		"amount": 129900
	}`

	req, firstRec := setupTestRequest(body)
	router.ServeHTTP(firstRec, req)

	req, secondRec := setupTestRequest(body)
	router.ServeHTTP(secondRec, req)

	require.Equal(t, http.StatusOK, firstRec.Code)
	require.Equal(t, http.StatusOK, secondRec.Code)

	var payment model.Payment
	require.NoError(t,
			db.Where("public_id = ?", "pay_demo_toss_001").
					First(&payment).Error,
	)

	assert.Equal(t, "COMPLETED", payment.Status)

	var eventCount int64
	require.NoError(t,
			db.Model(&model.PaymentEvent{}).
					Where("payment_id = ?", payment.ID).
					Count(&eventCount).Error,
	)

	assert.Equal(t, int64(1), eventCount)

	var outboxCount int64
	require.NoError(t,
			db.Model(&model.OutboxMessage{}).
					Where("deduplication_key = ?", payment.PublicID).
					Count(&outboxCount).Error,
	)

	assert.Equal(t, int64(1), outboxCount)
}

func TestTossReturnRejectsDifferentTransactionID(t *testing.T) {
	db, router := setupTest(t)

	firstBody := `{
		"paymentKey": "toss_key_demo_1001",
		"orderId": "pay_demo_toss_001",
		"amount": 129900
	}`

	secondBody := `{
		"paymentKey": "toss_key_demo_9999",
		"orderId": "pay_demo_toss_001",
		"amount": 129900
	}`

	req, firstRec := setupTestRequest(firstBody)
	router.ServeHTTP(firstRec, req)
	require.Equal(t, http.StatusOK, firstRec.Code)
	
	req, secondRec := setupTestRequest(secondBody)
	router.ServeHTTP(secondRec, req)
	require.Equal(t, http.StatusBadRequest, secondRec.Code)

	var payment model.Payment
	require.NoError(t,
		db.Where("public_id = ?", "pay_demo_toss_001").
			First(&payment).Error,
	)

	assert.Equal(t, model.PaymentStatusCompleted, payment.Status)
	assert.Equal(t, "toss_key_demo_1001", *payment.ExternalTransactionID)
}
