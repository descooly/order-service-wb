package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/descooly/order-service-wb/internal/cache"
	database "github.com/descooly/order-service-wb/internal/db"
	"github.com/descooly/order-service-wb/internal/httpserver"
	"github.com/descooly/order-service-wb/internal/nats"
	"github.com/descooly/order-service-wb/internal/service"
)

func main() {
	//if err := godotenv.Load(); err != nil {
	//	log.Println("No .env file found, using system environment")
	//} else {
	//	log.Println("Loaded .env file")
	//}

	db, err := database.ConnectDB()
	if err != nil {
		log.Fatal("Failed to connect to DB:", err)
	}
	defer db.Close()
	log.Println("Connected to DB")

	cache := cache.New()
	existingOrders, err := database.LoadOrders(db)
	if err != nil {
		log.Printf("Failed to load orders from DB: %v", err)
	} else {
		for _, order := range existingOrders {
			cache.Set(order)
		}
		log.Printf("Cache initialized with %d orders", cache.Len())
	}

	orderService := service.New(db, cache)

	httpServer := httpserver.NewServer(orderService)
	httpServer.Start()

	natsSubscriber, err := nats.NewSubscriber(orderService)
	if err != nil {
		log.Fatal("Failed to start NATS subscriber:", err)
	}
	defer natsSubscriber.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-signalChan
		log.Println("Shutdown signal received")
		cancel()
	}()

	natsSubscriber.Wait(ctx)

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP shutdown error: %v", err)
	}

	log.Println("Shutdown complete")
}
