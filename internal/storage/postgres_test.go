package storage_test

import (
	"context"
	"testing"

	"order-service-wbtech/internal/mocks"
	"order-service-wbtech/internal/model"

	"github.com/stretchr/testify/mock"
)

func TestServiceCallsStorage(t *testing.T) {
	dbMock := &mocks.Storage{}
	order := &model.Order{OrderUID: "123"}

	dbMock.On("SaveOrder", mock.Anything, order).Return(nil)

	err := dbMock.SaveOrder(context.Background(), order)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	dbMock.AssertCalled(t, "SaveOrder", mock.Anything, order)
}
