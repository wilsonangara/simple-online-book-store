// Code generated by MockGen. DO NOT EDIT.
// Source: order.go

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	models "github.com/wilsonangara/simple-online-book-store/storage/models"
)

// MockOrderStorage is a mock of OrderStorage interface.
type MockOrderStorage struct {
	ctrl     *gomock.Controller
	recorder *MockOrderStorageMockRecorder
}

// MockOrderStorageMockRecorder is the mock recorder for MockOrderStorage.
type MockOrderStorageMockRecorder struct {
	mock *MockOrderStorage
}

// NewMockOrderStorage creates a new mock instance.
func NewMockOrderStorage(ctrl *gomock.Controller) *MockOrderStorage {
	mock := &MockOrderStorage{ctrl: ctrl}
	mock.recorder = &MockOrderStorageMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockOrderStorage) EXPECT() *MockOrderStorageMockRecorder {
	return m.recorder
}

// Create mocks base method.
func (m *MockOrderStorage) Create(arg0 context.Context, arg1 *models.Order, arg2 []*models.OrderItem) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// Create indicates an expected call of Create.
func (mr *MockOrderStorageMockRecorder) Create(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockOrderStorage)(nil).Create), arg0, arg1, arg2)
}

// GetOrderHistory mocks base method.
func (m *MockOrderStorage) GetOrderHistory(arg0 context.Context, arg1 int64) ([]*models.OrderHistory, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetOrderHistory", arg0, arg1)
	ret0, _ := ret[0].([]*models.OrderHistory)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetOrderHistory indicates an expected call of GetOrderHistory.
func (mr *MockOrderStorageMockRecorder) GetOrderHistory(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetOrderHistory", reflect.TypeOf((*MockOrderStorage)(nil).GetOrderHistory), arg0, arg1)
}
