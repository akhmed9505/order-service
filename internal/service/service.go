package service

import (
	"context"
	"fmt"

	"order-service-wbtech/internal/model"
)

type Storage interface {
	SaveOrder(ctx context.Context, order *model.Order) error
	GetOrder(ctx context.Context, orderUID string) (*model.Order, error)
	LoadOrders(ctx context.Context) ([]*model.Order, error)
}

type Cache interface {
	Get(orderUID string) (*model.Order, bool)
	Set(order *model.Order)
}

type Service struct {
	cache Cache
	db    Storage
}

func New(c Cache, d Storage) *Service {
	return &Service{
		cache: c,
		db:    d,
	}
}

func (s *Service) CreateOrder(ctx context.Context, order *model.Order) error {
	if err := s.db.SaveOrder(ctx, order); err != nil {
		return fmt.Errorf("failed to save order to db: %w", err)
	}

	if order != nil && order.OrderUID != "" {
		s.cache.Set(order)
	}

	return nil
}

func (s *Service) GetOrder(ctx context.Context, orderUID string) (*model.Order, error) {
	order, ok := s.cache.Get(orderUID)
	if ok {
		return order, nil
	}

	order, err := s.db.GetOrder(ctx, orderUID)
	if err != nil {
		return nil, err
	}

	s.cache.Set(order)
	return order, nil
}

func (s *Service) RestoreCache(ctx context.Context) error {
	fmt.Println("Restoring cache from Database...")

	orders, err := s.db.LoadOrders(ctx)
	if err != nil {
		return err
	}

	for _, v := range orders {
		s.cache.Set(v)
	}

	return nil
}
