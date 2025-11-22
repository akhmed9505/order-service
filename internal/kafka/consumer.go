package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/segmentio/kafka-go"

	"order-service-wbtech/internal/model"
	"order-service-wbtech/internal/service"
	"order-service-wbtech/internal/validator"
)

func StartConsumer(ctx context.Context, brokers []string, topic, groupID string, svc *service.Service) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:          brokers,
		GroupID:          groupID,
		Topic:            topic,
		MinBytes:         1,
		MaxBytes:         10e6,
		ReadBatchTimeout: 10 * time.Second,
		CommitInterval:   0,
	})
	defer r.Close()

	log.Printf("kafka consumer started, topic=%s, group=%s", topic, groupID)

	for {
		m, err := r.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			log.Printf("kafka: fetch message error: %v. Retrying in 1s...", err)
			time.Sleep(time.Second)
			continue
		}

		var order model.Order
		if err := json.Unmarshal(m.Value, &order); err != nil {
			log.Printf("kafka: [BAD MESSAGE] Invalid message JSON (offset %d): %v. Committing offset to skip.", m.Offset, err)
			if commitErr := r.CommitMessages(ctx, m); commitErr != nil {
				log.Printf("kafka: commit failed for bad message (offset %d): %v", m.Offset, commitErr)
			}
			continue
		}

		if err := validator.ValidateOrder(&order); err != nil {
			log.Printf("kafka: [INVALID DATA] Order validation failed (offset %d): %v. Committing offset to skip.", m.Offset, err)
			if commitErr := r.CommitMessages(ctx, m); commitErr != nil {
				log.Printf("kafka: commit failed for invalid data (offset %d): %v", m.Offset, commitErr)
			}
			continue
		}

		if err := svc.CreateOrder(ctx, &order); err != nil {
			log.Printf("failed to process order %s (offset %d): %v. Will retry.", order.OrderUID, m.Offset, err)
			continue
		}

		log.Printf("Order processed and committed: %s (offset %d)", order.OrderUID, m.Offset)
		if err := r.CommitMessages(ctx, m); err != nil {
			log.Printf("kafka: WARNING! Commit failed for %s (offset %d): %v. Data is in DB, but offset not updated. May lead to reprocessing.", order.OrderUID, m.Offset, err)
		}
	}
}
