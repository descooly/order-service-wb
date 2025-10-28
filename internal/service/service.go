package service

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/descooly/order-service-wb/internal/cache"
	database "github.com/descooly/order-service-wb/internal/db"
	"github.com/descooly/order-service-wb/internal/model"

	"github.com/nats-io/stan.go"
)

type OrderService struct {
	Db    *sql.DB
	Cache *cache.OrderCache
}

func New(db *sql.DB, cache *cache.OrderCache) *OrderService {
	return &OrderService{
		Db:    db,
		Cache: cache,
	}
}

func (s *OrderService) HandleNATSMessage(msg *stan.Msg) {
	data := msg.Data
	if len(data) == 0 {
		log.Println("Empty message")
		return
	}

	var order model.OrderStruct
	if err := json.Unmarshal(data, &order); err != nil {
		log.Printf("Failed to parse JSON: %v, raw message: \n%v", err, data)
		return
	}
	log.Println("Received order")

	if order.OrderUID == "" {
		log.Println("Missing OrderUID in received message")
		return
	}

	if err := database.InsertOrder(s.Db, &order); err != nil {
		log.Printf("DB error: %v", err)
		return
	}

	s.Cache.Set(order)
	log.Println("Saved order to cache")

	msg.Ack()
}

func (s *OrderService) HTTPHandler(w http.ResponseWriter, r *http.Request) {
	uid := r.URL.Query().Get("order_uid")
	if uid == "" {
		http.Error(w, "Missing order_uid parameter", http.StatusBadRequest)
		return
	}
	if wo, exists := s.Cache.Get(uid); exists {
		data, err := json.Marshal(wo)
		if err != nil {
			log.Printf("JSON Marshal error for order %s: %v", uid, err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		var prettyJson bytes.Buffer
		if err := json.Indent(&prettyJson, data, "", "  "); err != nil {
			log.Printf("JSON Indent error for order %s: %v", uid, err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(prettyJson.Bytes())

	} else {
		http.Error(w, "Order not found", http.StatusNotFound)
	}
}
