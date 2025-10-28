package nats

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/descooly/order-service-wb/internal/service"

	"github.com/nats-io/stan.go"
)

const (
	subject     = "orders"
	clusterID   = "test-cluster"
	clientID    = "order-processor-1"
	durableName = "order-cache-durable"
)

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

type Subscriber struct {
	sc  stan.Conn
	sub stan.Subscription
}

func NewSubscriber(svc *service.OrderService) (*Subscriber, error) {
	natsHost := getEnv("NATS_HOST", "localhost")
	natsURL := fmt.Sprintf("nats://%s:4222", natsHost)

	sc, err := stan.Connect(
		clusterID,
		clientID,
		stan.NatsURL(natsURL),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS Streaming cluster %q at %s: %w", clusterID, natsURL, err)
	}

	log.Printf("Connected to NATS Streaming cluster %q", clusterID)

	sub, err := sc.Subscribe(subject, svc.HandleNATSMessage, stan.DurableName(durableName))
	if err != nil {
		sc.Close()
		return nil, fmt.Errorf("failed to subscribe to subject %q: %w", subject, err)
	}

	log.Printf("Subscribed to subject %q with durable name %q", subject, durableName)

	return &Subscriber{
		sc:  sc,
		sub: sub,
	}, nil
}

func (s *Subscriber) Close() {
	if s.sub != nil {
		s.sub.Unsubscribe()
	}
	if s.sc != nil {
		s.sc.Close()
	}
}

func (s *Subscriber) Wait(ctx context.Context) {
	<-ctx.Done()
	log.Println("NATS subscriber received shutdown signal")
	s.Close()
}
