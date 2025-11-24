package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"order-service-wbtech/internal/api"
	"order-service-wbtech/internal/cache"
	"order-service-wbtech/internal/config"
	"order-service-wbtech/internal/kafka"
	"order-service-wbtech/internal/service"
	"order-service-wbtech/internal/storage"
)

func main() {
	// cfg
	cfg := config.LoadConfig()
	log.Printf("cfg = %+v", cfg)

	// cache
	orderCache, err := cache.New()
	if err != nil {
		log.Fatalf("failed to initialize cache: %v", err)
	}

	// PostgreSQL
	pg, err := storage.NewPostgres(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer pg.Pool.Close()

	// service
	srvc := service.New(orderCache, pg)

	if err := srvc.RestoreCache(context.Background()); err != nil {
		fmt.Printf("Failed to load cache from DB: %v", err)
	}
	fmt.Println("Cache loaded from DB")

	// HTTP
	srv := api.New(srvc)
	httpServer := &http.Server{
		Addr:    ":" + cfg.HTTPPort,
		Handler: srv.Router(),
	}

	var wg sync.WaitGroup

	// Starting HTTP server
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Printf("HTTP api started on %s", cfg.HTTPPort)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("HTTP api error: %v", err)
		}
	}()

	// Starting Kafka consumer
	wg.Add(1)
	ctxKafka, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		defer wg.Done()
		brokers := strings.Split(cfg.Kafkacfg.Brokers, ",")
		kafka.StartConsumerWithWorkerPool(
			ctxKafka,
			brokers,
			cfg.Kafkacfg.Topic,
			cfg.Kafkacfg.GroupID,
			srvc,
			10,
		)
	}()

	// Waiting for SIGINT/SIGTERM signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	log.Println("Shutdown signal received")

	cancel()

	// Graceful shutdown HTTP
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP api shutdown error: %v", err)
	}

	// Waiting for goroutines to finish
	wg.Wait()
	log.Println("Service stopped cleanly")
}
