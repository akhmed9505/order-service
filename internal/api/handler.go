package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"order-service-wbtech/internal/model"
)

type Service interface {
	GetOrder(ctx context.Context, orderUID string) (*model.Order, error)
}

type Server struct {
	service Service
}

func New(s Service) *Server {
	return &Server{
		service: s,
	}
}

func (s *Server) Router() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/", http.FileServer(http.Dir("frontend")))
	mux.HandleFunc("/order/", s.handleGetOrder)

	return mux
}

func (s *Server) handleGetOrder(w http.ResponseWriter, r *http.Request) {
	orderUID := strings.TrimPrefix(r.URL.Path, "/order/")
	orderUID = strings.Trim(orderUID, "/")

	if orderUID == "" {
		http.Error(w, "order_uid is required", http.StatusBadRequest)
		return
	}

	log.Printf("HTTP GET /order/%s", orderUID)

	order, err := s.service.GetOrder(r.Context(), orderUID)
	if err != nil {
		log.Printf("GetOrder error: %v", err)
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(order); err != nil {
		log.Printf("json encode error: %v", err)
	}
}
