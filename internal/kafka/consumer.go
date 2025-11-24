package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"

	"order-service-wbtech/internal/model"
	"order-service-wbtech/internal/service"
	"order-service-wbtech/internal/validator"
)

// StartConsumerWithWorkerPool — main consumer startup with worker pool
func StartConsumerWithWorkerPool(
	ctx context.Context,
	brokers []string,
	topic string,
	groupID string,
	svc *service.Service,
	workerCount int,
) {
	if workerCount <= 0 {
		workerCount = 5
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:          brokers,
		GroupID:          groupID,
		Topic:            topic,
		MinBytes:         1,
		MaxBytes:         10e6,
		ReadBatchTimeout: 5 * time.Second,
		CommitInterval:   0,
		Logger:           kafka.LoggerFunc(log.Printf),
		ErrorLogger:      kafka.LoggerFunc(log.Printf),
	})

	jobs := make(chan kafka.Message, workerCount*2)
	var wg sync.WaitGroup

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			worker(ctx, workerID, jobs, svc, reader)
		}(i)
	}

	log.Printf("Kafka consumer started | topic=%s group=%s workers=%d", topic, groupID, workerCount)

	go func() {
		defer close(jobs)
		for {
			m, err := reader.FetchMessage(ctx)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					log.Println("Kafka consumer context cancelled, shutting down...")
					return
				}
				log.Printf("kafka: fetch message error: %v. Retrying...", err)
				time.Sleep(time.Second)
				continue
			}

			select {
			case jobs <- m:
			case <-ctx.Done():
				return
			}
		}
	}()

	<-ctx.Done()
	log.Println("Shutting down Kafka consumer...")

	if err := reader.Close(); err != nil {
		log.Printf("Error closing reader: %v", err)
	}
	wg.Wait()
	log.Println("Kafka consumer stopped gracefully")
}

func worker(ctx context.Context, workerID int, jobs <-chan kafka.Message, svc *service.Service, reader *kafka.Reader) {
	for msg := range jobs {
		if err := processMessage(ctx, msg, svc); err != nil {
			log.Printf("[worker-%d] Failed to process message (offset=%d): %v → will retry later", workerID, msg.Offset, err)
			continue
		}

		if err := reader.CommitMessages(ctx, msg); err != nil {
			log.Printf("[worker-%d] Commit failed for offset %d: %v", workerID, msg.Offset, err)
		} else {
			log.Printf("[worker-%d] Successfully processed and committed offset=%d", workerID, msg.Offset)
		}
	}
}

// processMessage processes a single message with all business logic
func processMessage(ctx context.Context, m kafka.Message, svc *service.Service) error {
	log.Printf("Processing message offset=%d partition=%d", m.Offset, m.Partition)

	var order model.Order
	if err := json.Unmarshal(m.Value, &order); err != nil {
		log.Printf("Invalid JSON (offset %d): %v → skipping (no retry)", m.Offset, err)
		return nil
	}

	if err := validator.ValidateOrder(&order); err != nil {
		log.Printf("Validation failed for order %s (offset %d): %v → skipping", order.OrderUID, m.Offset, err)
		return nil
	}

	if err := svc.CreateOrder(ctx, &order); err != nil {
		return err
	}

	return nil
}
