package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"project/internal"
	my_cache "project/internal/cache"
	database "project/internal/db"

	_ "github.com/lib/pq"
	"github.com/nats-io/stan.go"
)

const subject = "orders"

type OrderService struct {
	db    *sql.DB
	cache *my_cache.OrderCache
}

func (s *OrderService) HandleMsg(msg *stan.Msg) {
	data := msg.Data
	if len(data) == 0 {
		log.Println("Empty message")
		return
	}

	var order internal.OrderStruct
	if err := json.Unmarshal(data, &order); err != nil {
		log.Printf("Failed to parse JSON: %v, raw message: \n%v", err, data)
		return
	}
	log.Println("Received order")
	if order.OrderUID == "" {
		log.Println("Missing OrderUID in received message")
		return
	}

	if err := database.InsertOrder(s.db, &order); err != nil {
		log.Printf("DB error: %v", err)
		return
	}

	s.cache.Set(order)
	log.Println("Saved order to cache")

	msg.Ack()

}

func (o *OrderService) HTTPHandler(w http.ResponseWriter, r *http.Request) {
	uid := r.URL.Query().Get("order_uid")
	if uid == "" {
		http.Error(w, "Missing order_uid parameter", http.StatusBadRequest)
		return
	}
	if wo, exists := o.cache.Get(uid); exists {
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

func main() {
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}
	db, err := database.ConnectDB(dbHost)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	log.Println("Connected to DB")

	existingOrders, err := database.LoadOrders(db)
	if err != nil {
		log.Println(err)
		return
	}
	service := &OrderService{db: db, cache: my_cache.New()}
	for _, elem := range existingOrders {
		service.cache.Set(elem)
	}
	log.Printf("Cache initialized with %d orders", service.cache.Len())

	go func() {
		log.Println("HTTP Server starting on :8080")
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/" {
				http.ServeFile(w, r, "static/index.html")
				return
			}
			http.NotFound(w, r)
		})

		http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))

		http.HandleFunc("/order", service.HTTPHandler)

		log.Fatal(http.ListenAndServe(":8080", nil))
	}()

	natsHost := os.Getenv("NATS_HOST")
	if natsHost == "" {
		natsHost = "localhost"
	}
	sc, err := stan.Connect(
		"test-cluster",
		"order-processor-1",
		stan.NatsURL(fmt.Sprintf("nats://%s:4222", natsHost)),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer sc.Close()
	log.Println("Connected to test-cluster")

	sub, err := sc.Subscribe(subject, service.HandleMsg, stan.DurableName("order-cache-durable"))
	if err != nil {
		log.Println(err)
		return
	}
	log.Printf("Subscribed to channel")

	defer sub.Unsubscribe()

	select {}
}
