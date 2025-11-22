package cache

import (
	"testing"

	"order-service-wbtech/internal/model"
)

func TestCacheSetGet(t *testing.T) {
	c, err := New()
	if err != nil {
		t.Fatalf("failed to create cache: %v", err)
	}

	order := &model.Order{OrderUID: "123"}
	c.Set(order)

	got, ok := c.Get("123")
	if !ok {
		t.Fatal("expected order to be found in cache")
	}
	if got.OrderUID != "123" {
		t.Errorf("expected OrderUID '123', got '%s'", got.OrderUID)
	}

	_, ok = c.Get("456")
	if ok {
		t.Error("expected order '456' to be missing")
	}
}
