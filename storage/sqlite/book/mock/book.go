// Code generated by MockGen. DO NOT EDIT.
// Source: book.go

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	models "github.com/wilsonangara/simple-online-book-store/storage/models"
)

// MockBookStorage is a mock of BookStorage interface.
type MockBookStorage struct {
	ctrl     *gomock.Controller
	recorder *MockBookStorageMockRecorder
}

// MockBookStorageMockRecorder is the mock recorder for MockBookStorage.
type MockBookStorageMockRecorder struct {
	mock *MockBookStorage
}

// NewMockBookStorage creates a new mock instance.
func NewMockBookStorage(ctrl *gomock.Controller) *MockBookStorage {
	mock := &MockBookStorage{ctrl: ctrl}
	mock.recorder = &MockBookStorageMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockBookStorage) EXPECT() *MockBookStorageMockRecorder {
	return m.recorder
}

// GetBooks mocks base method.
func (m *MockBookStorage) GetBooks(arg0 context.Context) ([]*models.Book, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBooks", arg0)
	ret0, _ := ret[0].([]*models.Book)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetBooks indicates an expected call of GetBooks.
func (mr *MockBookStorageMockRecorder) GetBooks(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBooks", reflect.TypeOf((*MockBookStorage)(nil).GetBooks), arg0)
}
