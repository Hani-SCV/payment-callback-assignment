package tests

import (
	"testing"
	"time"

	"github.com/Hani-SCV/payment-callback-assignment/internal/model"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func resetTestDB(t *testing.T, db *gorm.DB) {
	t.Helper()

	tables := []string{
		"outbox_messages",
		"payment_events",
		"payments",
		"orders",
	}

	for _, table := range tables {
		require.NoError(t, db.Exec("DELETE FROM "+table).Error)
	}
}

func seedSyntheticData(t *testing.T, db *gorm.DB) {
	t.Helper()

	created := time.Date(2026, 1, 10, 1, 0, 0, 0, time.UTC)

	order := model.Order{
		PublicID:          "ord_demo_1001",
		CustomerReference: "customer_demo_a",
		Status:            "PAYMENT_PENDING",
		CreatedAt:         created,
		UpdatedAt:         created,
	}

	require.NoError(t, db.Create(&order).Error)

	payment := model.Payment{
		PublicID:  "pay_demo_toss_001",
		OrderID:   order.ID,
		Provider:  "TOSS",
		Status:    "PENDING",
		Amount:    decimal.NewFromInt(129900),
		Currency:  "KRW",
		CreatedAt: created,
		UpdatedAt: created,
	}

	require.NoError(t, db.Create(&payment).Error)

	stripePayment := model.Payment{
		PublicID:  "pay_demo_stripe_001",
		OrderID:   order.ID,
		Provider:  "STRIPE",
		Status:    "PENDING",
		Amount:    decimal.NewFromInt(7700),
		Currency:  "usd",
		CreatedAt: created,
		UpdatedAt: created,
	}

	require.NoError(t, db.Create(&stripePayment).Error)

	alipayOrder := model.Order{
		PublicID:          "ord_demo_alipay_1001",
		CustomerReference: "customer_demo_alipay",
		Status:            "PAYMENT_PENDING",
		CreatedAt:         created,
		UpdatedAt:         created,
	}

	require.NoError(t, db.Create(&alipayOrder).Error)

	alipayPayment := model.Payment{
		PublicID:  "pay_demo_alipay_001",
		OrderID:   alipayOrder.ID,
		Provider:  "ALIPAY",
		Status:    "PENDING",
		Amount:    decimal.NewFromInt(129900),
		Currency:  "CNY",
		CreatedAt: created,
		UpdatedAt: created,
	}

	require.NoError(t, db.Create(&alipayPayment).Error)
}
