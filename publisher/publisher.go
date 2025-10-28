package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/descooly/order-service-wb/internal/model"

	"github.com/nats-io/stan.go"
)

func main() {
	sc, err := stan.Connect("test-cluster", "publisher", stan.NatsURL("nats://localhost:4222"))
	if err != nil {
		log.Printf("Connection error: %v", err)
	}
	defer sc.Close()

	order := model.OrderStruct{
		OrderUID:          "TEST123",
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

	data, err := json.Marshal(order)
	if err != nil {
		log.Println(err)
	}

	err = sc.Publish("orders", data)
	if err != nil {
		log.Println(err)
	}
	log.Println("Message successfully sent")
}
