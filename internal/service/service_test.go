package service

import (
	"context"
	"testing"

	"order-service-wbtech/internal/mocks"
	"order-service-wbtech/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateOrder(t *testing.T) {
	ctx := context.Background()
	dbMock := &mocks.Storage{}
	cacheMock := &mocks.Cache{}

	order := &model.Order{OrderUID: "123"}

	dbMock.On("SaveOrder", mock.Anything, order).Return(nil)
	cacheMock.On("Set", order).Return()

	svc := New(cacheMock, dbMock)

	err := svc.CreateOrder(ctx, order)
	assert.NoError(t, err)

	dbMock.AssertCalled(t, "SaveOrder", mock.Anything, order)
	cacheMock.AssertCalled(t, "Set", order)
}

func TestGetOrder_FromCache(t *testing.T) {
	ctx := context.Background()
	dbMock := &mocks.Storage{}
	cacheMock := &mocks.Cache{}

	order := &model.Order{OrderUID: "123"}

	cacheMock.On("Get", "123").Return(order, true)

	svc := New(cacheMock, dbMock)

	got, err := svc.GetOrder(ctx, "123")
	assert.NoError(t, err)
	assert.Equal(t, order, got)

	dbMock.AssertNotCalled(t, "GetOrder", mock.Anything, mock.Anything)
}

func TestGetOrder_FromDB(t *testing.T) {
	ctx := context.Background()
	dbMock := &mocks.Storage{}
	cacheMock := &mocks.Cache{}

	order := &model.Order{OrderUID: "123"}

	cacheMock.On("Get", "123").Return(nil, false)
	dbMock.On("GetOrder", mock.Anything, "123").Return(order, nil)
	cacheMock.On("Set", order).Return()

	svc := New(cacheMock, dbMock)

	got, err := svc.GetOrder(ctx, "123")
	assert.NoError(t, err)
	assert.Equal(t, order, got)

	dbMock.AssertCalled(t, "GetOrder", mock.Anything, "123")
	cacheMock.AssertCalled(t, "Set", order)
}
