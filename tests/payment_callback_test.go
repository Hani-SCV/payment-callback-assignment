package tests

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/Hani-SCV/payment-callback-assignment/internal/app"
	"github.com/Hani-SCV/payment-callback-assignment/internal/config"
	"github.com/Hani-SCV/payment-callback-assignment/internal/model"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const (
	TossReturnPath    = "/v1/payment-callbacks/toss/return"
	StripeWebhookPath = "/v1/payment-callbacks/stripe/webhook"
	AlipayNotifyPath  = "/v1/payment-callbacks/alipay/notify"
)

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

func setupTestRequest(path string, body string) (*http.Request, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(
		http.MethodPost,
		path,
		bytes.NewBufferString(body),
	)

	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	return req, rec
}

func setupAlipayTestRequest(
	path string,
	form url.Values,
) (*http.Request, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(
		http.MethodPost,
		path,
		strings.NewReader(form.Encode()),
	)

	req.Header.Set(
		"Content-Type",
		"application/x-www-form-urlencoded",
	)

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

	req, rec := setupTestRequest(TossReturnPath, body)

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

	req, firstRec := setupTestRequest(TossReturnPath, body)
	router.ServeHTTP(firstRec, req)

	req, secondRec := setupTestRequest(TossReturnPath, body)
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

	req, firstRec := setupTestRequest(TossReturnPath, firstBody)
	router.ServeHTTP(firstRec, req)
	require.Equal(t, http.StatusOK, firstRec.Code)

	req, secondRec := setupTestRequest(TossReturnPath, secondBody)
	router.ServeHTTP(secondRec, req)
	require.Equal(t, http.StatusConflict, secondRec.Code)

	var payment model.Payment
	require.NoError(t,
		db.Where("public_id = ?", "pay_demo_toss_001").
			First(&payment).Error,
	)

	assert.Equal(t, model.PaymentStatusCompleted, payment.Status)
	assert.Equal(t, "toss_key_demo_1001", *payment.ExternalTransactionID)
}

func TestTossReturnRejectsAmountMismatch(t *testing.T) {
	db, router := setupTest(t)

	body := `{
		"paymentKey": "toss_key_demo_1001",
		"orderId": "pay_demo_toss_001",
		"amount": 100000
	}`

	req, rec := setupTestRequest(TossReturnPath, body)

	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusUnprocessableEntity, rec.Code)

	var payment model.Payment
	require.NoError(t,
		db.Where("public_id = ?", "pay_demo_toss_001").
			First(&payment).Error,
	)
	assert.Equal(t, model.PaymentStatusPending, payment.Status)

	var order model.Order
	require.NoError(t,
		db.Where("public_id = ?", "ord_demo_1001").
			First(&order).Error,
	)
	assert.Equal(t, "PAYMENT_PENDING", order.Status)
}

func TestTossReturnRejectsInvalidProvider(t *testing.T) {
	db, router := setupTest(t)

	var order model.Order
	require.NoError(t,
		db.Where("public_id = ?", "ord_demo_1001").
			First(&order).Error,
	)

	stripePayment := model.Payment{
		PublicID: "pay_demo_toss_invalid_provider",
		OrderID:  order.ID,
		Provider: "STRIPE",
		Status:   "PENDING",
		Amount:   decimal.NewFromInt(129900),
		Currency: "KRW",
	}

	require.NoError(t, db.Create(&stripePayment).Error)

	body := `{
		"paymentKey": "toss_key_demo_1001",
		"orderId": "pay_demo_toss_invalid_provider",
		"amount": 129900
	}`

	req, rec := setupTestRequest(TossReturnPath, body)
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnprocessableEntity, rec.Code)

	var payment model.Payment
	require.NoError(t,
		db.Where("public_id = ?", "pay_demo_toss_invalid_provider").
			First(&payment).Error,
	)

	assert.Equal(t, "STRIPE", payment.Provider)
	assert.Equal(t, model.PaymentStatusPending, payment.Status)
	assert.Nil(t, payment.ExternalTransactionID)
	assert.Nil(t, payment.CompletedAt)
}

func TestTossReturnRejectsInvalidCurrency(t *testing.T) {
	db, router := setupTest(t)

	require.NoError(t,
		db.Model(&model.Payment{}).
			Where("public_id = ?", "pay_demo_toss_001").
			Update("currency", "CNY").Error,
	)

	body := `{
		"paymentKey": "toss_key_demo_1001",
		"orderId": "pay_demo_toss_001",
		"amount": 129900
	}`

	req, rec := setupTestRequest(TossReturnPath, body)
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnprocessableEntity, rec.Code)

	var payment model.Payment
	require.NoError(t,
		db.Where("public_id = ?", "pay_demo_toss_001").
			First(&payment).Error,
	)

	assert.Equal(t, "CNY", payment.Currency)
	assert.Equal(t, model.PaymentStatusPending, payment.Status)
	assert.Nil(t, payment.ExternalTransactionID)
	assert.Nil(t, payment.CompletedAt)
}

func TestTossReturnRejectsInvalidOrderStatus(t *testing.T) {
	db, router := setupTest(t)

	require.NoError(t,
		db.Model(&model.Order{}).
			Where("public_id = ?", "ord_demo_1001").
			Update("status", "PAID").Error,
	)

	body := `{
		"paymentKey": "toss_key_demo_1001",
		"orderId": "pay_demo_toss_001",
		"amount": 129900
	}`

	req, rec := setupTestRequest(TossReturnPath, body)
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusConflict, rec.Code)

	var order model.Order
	require.NoError(t,
		db.Where("public_id = ?", "ord_demo_1001").
			First(&order).Error,
	)

	assert.Equal(t, "PAID", order.Status)

	var payment model.Payment
	require.NoError(t,
		db.Where("public_id = ?", "pay_demo_toss_001").
			First(&payment).Error,
	)

	assert.Equal(t, model.PaymentStatusPending, payment.Status)
	assert.Nil(t, payment.ExternalTransactionID)
	assert.Nil(t, payment.CompletedAt)
}

func TestTossReturnRejectsInvalidPaymentStatus(t *testing.T) {
	db, router := setupTest(t)

	require.NoError(t,
		db.Model(&model.Payment{}).
			Where("public_id = ?", "pay_demo_toss_001").
			Update("status", "FAILED").Error,
	)

	body := `{
		"paymentKey": "toss_key_demo_1001",
		"orderId": "pay_demo_toss_001",
		"amount": 129900
	}`

	req, rec := setupTestRequest(TossReturnPath, body)
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusConflict, rec.Code)

	var payment model.Payment
	require.NoError(t,
		db.Where("public_id = ?", "pay_demo_toss_001").
			First(&payment).Error,
	)

	assert.Equal(t, model.PaymentStatusFailed, payment.Status)
	assert.Nil(t, payment.ExternalTransactionID)
	assert.Nil(t, payment.CompletedAt)
}

func TestTossReturnRejectsNonExistentPayment(t *testing.T) {
	_, router := setupTest(t)

	body := `{
		"paymentKey": "toss_key_demo_100",
		"orderId": "pay_demo_toss_01",
		"amount": 129900
	}`

	req, rec := setupTestRequest(TossReturnPath, body)
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestTossReturnRejectsInvalidJSON(t *testing.T) {
	_, router := setupTest(t)

	body := `invalid json`

	req, rec := setupTestRequest(TossReturnPath, body)
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestStripeWebhookCompletesPayment(t *testing.T) {
	db, router := setupTest(t)

	body := `{
		"id": "evt_demo_1001",
		"type": "checkout.session.completed",
		"data": {
			"object": {
				"id": "cs_demo_1001",
				"client_reference_id": "pay_demo_stripe_001",
				"amount_total": 7700,
				"currency": "usd",
				"payment_status": "paid"
			}
		}
	}`

	req, rec := setupTestRequest(StripeWebhookPath, body)
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var payment model.Payment
	require.NoError(t,
		db.Where("public_id = ?", "pay_demo_stripe_001").First(&payment).Error,
	)

	assert.Equal(t, model.PaymentStatusCompleted, payment.Status)
}

func TestStripeWebhookIsIdempotent(t *testing.T) {
	_, router := setupTest(t)

	body := `{
		"id": "evt_demo_1001",
		"type": "checkout.session.completed",
		"data": {
			"object": {
				"id": "cs_demo_1001",
				"client_reference_id": "pay_demo_stripe_001",
				"amount_total": 7700,
				"currency": "usd",
				"payment_status": "paid"
			}
		}
	}`

	firstReq, firstRec := setupTestRequest(StripeWebhookPath, body)
	secondReq, secondRec := setupTestRequest(StripeWebhookPath, body)

	router.ServeHTTP(firstRec, firstReq)
	router.ServeHTTP(secondRec, secondReq)

	require.Equal(t, http.StatusOK, firstRec.Code)
	require.Equal(t, http.StatusOK, secondRec.Code)
}

func TestStripeWebhookRejectsInvalidEventType(t *testing.T) {
	_, router := setupTest(t)

	body := `{
		"id": "evt_demo_1001",
		"type": "payment_intent.created",
		"data": {
			"object": {
				"id": "cs_demo_1001",
				"client_reference_id": "pay_demo_stripe_001",
				"amount_total": 7700,
				"currency": "usd",
				"payment_status": "paid"
			}
		}
	}`

	req, rec := setupTestRequest(StripeWebhookPath, body)
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestStripeWebhookRejectsAmountMismatch(t *testing.T) {
	_, router := setupTest(t)

	body := `{
		"id": "evt_demo_1001",
		"type": "checkout.session.completed",
		"data": {
			"object": {
				"id": "cs_demo_1001",
				"client_reference_id": "pay_demo_stripe_001",
				"amount_total": 9999,
				"currency": "usd",
				"payment_status": "paid"
			}
		}
	}`

	req, rec := setupTestRequest(StripeWebhookPath, body)
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestStripeWebhookRejectsCurrencyMismatch(t *testing.T) {
	_, router := setupTest(t)

	body := `{
		"id": "evt_demo_1001",
		"type": "checkout.session.completed",
		"data": {
			"object": {
				"id": "cs_demo_1001",
				"client_reference_id": "pay_demo_stripe_001",
				"amount_total": 7700,
				"currency": "krw",
				"payment_status": "paid"
			}
		}
	}`

	req, rec := setupTestRequest(StripeWebhookPath, body)
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestStripeWebhookRejectsUnpaidPayment(t *testing.T) {
	_, router := setupTest(t)

	body := `{
		"id": "evt_demo_1001",
		"type": "checkout.session.completed",
		"data": {
			"object": {
				"id": "cs_demo_1001",
				"client_reference_id": "pay_demo_stripe_001",
				"amount_total": 7700,
				"currency": "usd",
				"payment_status": "unpaid"
			}
		}
	}`

	req, rec := setupTestRequest(StripeWebhookPath, body)
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusConflict, rec.Code)
}

func TestStripeWebhookRejectsInvalidJSON(t *testing.T) {
	_, router := setupTest(t)

	body := `invalid json`

	req, rec := setupTestRequest(StripeWebhookPath, body)
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAlipayNotifyCompletesPayment(t *testing.T) {
	db, router := setupTest(t)

	form := url.Values{}
	form.Set("order_id", "pay_demo_alipay_001")
	form.Set("trade_no", "alipay_trade_demo_1001")
	form.Set("total_amount", "129900.00")

	req, rec := setupAlipayTestRequest(
		AlipayNotifyPath,
		form,
	)
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var payment model.Payment
	require.NoError(t,
		db.Where("public_id = ?", "pay_demo_alipay_001").
			First(&payment).Error,
	)
	assert.Equal(t, model.PaymentStatusCompleted, payment.Status)

	require.NotNil(t, payment.ExternalTransactionID)
	assert.Equal(t, "alipay_trade_demo_1001", *payment.ExternalTransactionID)

	var order model.Order
	require.NoError(t,
		db.First(&order, payment.OrderID).Error,
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

func TestAlipayNotifyIsIdempotent(t *testing.T) {
	db, router := setupTest(t)

	form := url.Values{}
	form.Set("order_id", "pay_demo_alipay_001")
	form.Set("trade_no", "alipay_trade_demo_1001")
	form.Set("total_amount", "129900.00")

	firstReq, firstRec := setupAlipayTestRequest(AlipayNotifyPath, form)
	router.ServeHTTP(firstRec, firstReq)
	secondReq, secondRec := setupAlipayTestRequest(AlipayNotifyPath, form)
	router.ServeHTTP(secondRec, secondReq)

	require.Equal(t, http.StatusOK, firstRec.Code)
	require.Equal(t, http.StatusOK, secondRec.Code)

	var payment model.Payment
	require.NoError(t,
		db.Where("public_id = ?", "pay_demo_alipay_001").
			First(&payment).Error,
	)

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

func TestAlipayNotifyRejectsDifferentTransactionID(t *testing.T) {
	db, router := setupTest(t)

	form := url.Values{}
	form.Set("order_id", "pay_demo_alipay_001")
	form.Set("trade_no", "alipay_trade_demo_1001")
	form.Set("total_amount", "129900.00")

	req, rec := setupAlipayTestRequest(AlipayNotifyPath, form)
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	form.Set("trade_no", "alipay_trade_demo_1002")

	req, rec = setupAlipayTestRequest(AlipayNotifyPath, form)
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusConflict, rec.Code)

	var payment model.Payment
	require.NoError(t,
		db.Where("public_id = ?", "pay_demo_alipay_001").
			First(&payment).Error,
	)

	require.NotNil(t, payment.ExternalTransactionID)
	assert.Equal(t, "alipay_trade_demo_1001", *payment.ExternalTransactionID)
}

func TestAlipayNotifyRejectsAmountMismatch(t *testing.T) {
	db, router := setupTest(t)

	form := url.Values{}
	form.Set("order_id", "pay_demo_alipay_001")
	form.Set("trade_no", "alipay_trade_demo_1001")
	form.Set("total_amount", "100000.00")

	req, rec := setupAlipayTestRequest(AlipayNotifyPath, form)
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnprocessableEntity, rec.Code)

	var payment model.Payment
	require.NoError(t,
		db.Where("public_id = ?", "pay_demo_alipay_001").
			First(&payment).Error,
	)
	assert.Equal(t, model.PaymentStatusPending, payment.Status)

	var order model.Order
	require.NoError(t,
		db.First(&order, payment.OrderID).Error,
	)
	assert.Equal(t, "PAYMENT_PENDING", order.Status)
}

func TestAlipayNotifyRejectsInvalidProvider(t *testing.T) {
	db, router := setupTest(t)

	require.NoError(t,
		db.Model(&model.Payment{}).
			Where("public_id = ?", "pay_demo_alipay_001").
			Update("provider", "TOSS").Error,
	)

	form := url.Values{}
	form.Set("order_id", "pay_demo_alipay_001")
	form.Set("trade_no", "alipay_trade_demo_1001")
	form.Set("total_amount", "129900.00")

	req, rec := setupAlipayTestRequest(AlipayNotifyPath, form)
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnprocessableEntity, rec.Code)

	var payment model.Payment
	require.NoError(t,
		db.Where("public_id = ?", "pay_demo_alipay_001").
			First(&payment).Error,
	)
	assert.Equal(t, model.PaymentStatusPending, payment.Status)
}

func TestAlipayNotifyRejectsInvalidCurrency(t *testing.T) {
	db, router := setupTest(t)

	require.NoError(t,
		db.Model(&model.Payment{}).
			Where("public_id = ?", "pay_demo_alipay_001").
			Update("currency", "KRW").Error,
	)

	form := url.Values{}
	form.Set("order_id", "pay_demo_alipay_001")
	form.Set("trade_no", "alipay_trade_demo_1001")
	form.Set("total_amount", "129900.00")

	req, rec := setupAlipayTestRequest(AlipayNotifyPath, form)
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnprocessableEntity, rec.Code)

	var payment model.Payment
	require.NoError(t,
		db.Where("public_id = ?", "pay_demo_alipay_001").
			First(&payment).Error,
	)
	assert.Equal(t, model.PaymentStatusPending, payment.Status)
}

func TestAlipayNotifyRejectsInvalidOrderStatus(t *testing.T) {
	db, router := setupTest(t)

	var payment model.Payment
	require.NoError(t,
		db.Where("public_id = ?", "pay_demo_alipay_001").
			First(&payment).Error,
	)

	require.NoError(t,
		db.Model(&model.Order{}).
			Where("id = ?", payment.OrderID).
			Update("status", "CANCELED").Error,
	)

	form := url.Values{}
	form.Set("order_id", "pay_demo_alipay_001")
	form.Set("trade_no", "alipay_trade_demo_1001")
	form.Set("total_amount", "129900.00")

	req, rec := setupAlipayTestRequest(AlipayNotifyPath, form)
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusConflict, rec.Code)

	require.NoError(t,
		db.Where("public_id = ?", "pay_demo_alipay_001").
			First(&payment).Error,
	)
	assert.Equal(t, model.PaymentStatusPending, payment.Status)
}

func TestAlipayNotifyRejectsInvalidPaymentStatus(t *testing.T) {
	db, router := setupTest(t)

	require.NoError(t,
		db.Model(&model.Payment{}).
			Where("public_id = ?", "pay_demo_alipay_001").
			Update("status", model.PaymentStatusCanceled).Error,
	)

	form := url.Values{}
	form.Set("order_id", "pay_demo_alipay_001")
	form.Set("trade_no", "alipay_trade_demo_1001")
	form.Set("total_amount", "129900.00")

	req, rec := setupAlipayTestRequest(AlipayNotifyPath, form)
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusConflict, rec.Code)

	var payment model.Payment
	require.NoError(t,
		db.Where("public_id = ?", "pay_demo_alipay_001").
			First(&payment).Error,
	)
	assert.Equal(t, model.PaymentStatusCanceled, payment.Status)
}

func TestAlipayNotifyRejectsNonExistentPayment(t *testing.T) {
	_, router := setupTest(t)

	form := url.Values{}
	form.Set("order_id", "pay_demo_alipay_not_found")
	form.Set("trade_no", "alipay_trade_demo_1001")
	form.Set("total_amount", "129900.00")

	req, rec := setupAlipayTestRequest(
		AlipayNotifyPath,
		form,
	)
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestAlipayNotifyRejectsInvalidAmount(t *testing.T) {
	_, router := setupTest(t)

	form := url.Values{}
	form.Set("order_id", "pay_demo_alipay_001")
	form.Set("trade_no", "alipay_trade_demo_1001")
	form.Set("total_amount", "invalid-amount")

	req, rec := setupAlipayTestRequest(
		AlipayNotifyPath,
		form,
	)
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestAlipayNotifyRejectsInvalidForm(t *testing.T) {
	_, router := setupTest(t)

	req := httptest.NewRequest(
		http.MethodPost,
		AlipayNotifyPath,
		strings.NewReader("%zz"),
	)
	req.Header.Set(
		"Content-Type",
		"application/x-www-form-urlencoded",
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}
