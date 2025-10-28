package cache

import (
	"testing"
	"time"

	"github.com/descooly/order-service-wb/internal/model"
)

func TestSetAndGet(t *testing.T) {
	c := New()

	order := model.OrderStruct{
		OrderUID:          "testy_test123",
		TrackNumber:       "2222221111111",
		Entry:             "WBIL",
		Locale:            "en",
		InternalSignature: "",
		CustomerId:        "test",
		DeliveryService:   "meest",
		Shardkey:          "9",
		SmId:              99,
		DateCreated:       time.Date(2021, time.November, 26, 6, 22, 19, 0, time.UTC),
		OofShard:          "1",
		Delivery: model.Delivery{
			Name:    "Test Testov",
			Phone:   "+9720000000",
			Zip:     "2639809",
			City:    "Kiryat Mozkin",
			Address: "Ploshad Mira 15",
			Region:  "Kraiot",
			Email:   "test@gmail.com",
		},
		Payment: model.Payment{
			Transaction:  "b563feb7b2b84b6test",
			RequestId:    "",
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       1817,
			PaymentDt:    1637907727,
			Bank:         "alpha",
			DeliveryCost: 1500,
			GoodsTotal:   317,
			CustomFee:    0,
		},
		Items: []model.Item{
			{
				ChrtId:      9934930,
				TrackNumber: "WBILMTESTTRACK",
				Price:       453,
				Rid:         "ab4219087a764ae0btest",
				Name:        "Mascaras",
				Sale:        30,
				Size:        "0",
				TotalPrice:  317,
				NmId:        2389212,
				Brand:       "Vivienne Sabo",
				Status:      202,
			},
		},
	}

	c.Set(order)

	got, ok := c.Get("testy_test123")
	if !ok {
		t.Fatal("Expected order to be found")
	}

	if got.OrderUID != "testy_test123" {
		t.Errorf("got OrderUID: %v, want: %v", got.OrderUID, "testy_test123")
	}
}

func TestGetNotFound(t *testing.T) {
	c := New()
	_, ok := c.Get("non-existent-order123")
	if ok {
		t.Error("expected order not to be found")
	}
}
