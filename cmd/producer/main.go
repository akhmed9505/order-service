package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"order-service-wbtech/internal/config"
	"order-service-wbtech/internal/model"

	"github.com/joho/godotenv"
	"github.com/segmentio/kafka-go"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not loaded, relying on environment variables: %v", err)
	}

	cfg := config.LoadConfig()

	brokers := []string{os.Getenv("KAFKA_BROKERS")}
	if brokers[0] == "" {
		brokers = []string{"localhost:9094"}
	}
	topic := cfg.Kafkacfg.Topic

	w := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}
	defer func() {
		if err := w.Close(); err != nil {
			log.Printf("Failed to close Kafka writer: %v", err)
		}
	}()

	log.Printf("Kafka producer started, brokers=%v, topic=%s", brokers, topic)

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		order := generateTestOrder()

		b, err := json.Marshal(order)
		if err != nil {
			log.Printf("failed to marshal order: %v", err)
			continue
		}

		msg := kafka.Message{
			Key:   []byte(order.OrderUID),
			Value: b,
			Time:  time.Now(),
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err = w.WriteMessages(ctx, msg)
		cancel()

		if err != nil {
			log.Printf("failed to write message order_uid=%s: %v", order.OrderUID, err)
			continue
		}

		log.Printf("message produced order_uid=%s", order.OrderUID)
	}
}

// generateTestOrder creates a sample order for the producer
func generateTestOrder() *model.Order {
	ts := time.Now()
	return &model.Order{
		OrderUID:          fmt.Sprintf("order-%d", ts.UnixNano()),
		TrackNumber:       "PTRACK",
		Entry:             "WBIL",
		Locale:            "en",
		CustomerID:        "cust-prod",
		DateCreated:       ts,
		InternalSignature: "",
		DeliveryService:   "meest",
		ShardKey:          "9",
		SMID:              99,
		OOFShard:          "1",

		Delivery: model.Delivery{
			Name:    "Producer User",
			Phone:   "+10000000000",
			Zip:     "1",
			City:    "City",
			Address: "Addr",
			Region:  "Region",
			Email:   "p@example.com",
		},
		Payment: model.Payment{
			Transaction:  "txn-prod",
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       42,
			PaymentDt:    ts.Unix(),
			Bank:         "bank",
			DeliveryCost: 0,
			GoodsTotal:   42,
			CustomFee:    0,
		},
		Items: []model.Item{
			{
				ChrtID:      1,
				TrackNumber: "TRK123456",
				Price:       42,
				Rid:         "rid1",
				Name:        "ItemP",
				Sale:        0,
				Size:        "M",
				TotalPrice:  42,
				NmID:        1,
				Brand:       "brand",
				Status:      202,
			},
		},
	}
}
