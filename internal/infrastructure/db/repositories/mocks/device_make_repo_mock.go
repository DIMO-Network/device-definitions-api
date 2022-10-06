// Code generated by MockGen. DO NOT EDIT.
// Source: device_make_repo.go

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	models "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	gomock "github.com/golang/mock/gomock"
)

// MockDeviceMakeRepository is a mock of DeviceMakeRepository interface.
type MockDeviceMakeRepository struct {
	ctrl     *gomock.Controller
	recorder *MockDeviceMakeRepositoryMockRecorder
}

// MockDeviceMakeRepositoryMockRecorder is the mock recorder for MockDeviceMakeRepository.
type MockDeviceMakeRepositoryMockRecorder struct {
	mock *MockDeviceMakeRepository
}

// NewMockDeviceMakeRepository creates a new mock instance.
func NewMockDeviceMakeRepository(ctrl *gomock.Controller) *MockDeviceMakeRepository {
	mock := &MockDeviceMakeRepository{ctrl: ctrl}
	mock.recorder = &MockDeviceMakeRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockDeviceMakeRepository) EXPECT() *MockDeviceMakeRepositoryMockRecorder {
	return m.recorder
}

// GetAll mocks base method.
func (m *MockDeviceMakeRepository) GetAll(ctx context.Context) ([]*models.DeviceMake, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAll", ctx)
	ret0, _ := ret[0].([]*models.DeviceMake)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAll indicates an expected call of GetAll.
func (mr *MockDeviceMakeRepositoryMockRecorder) GetAll(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAll", reflect.TypeOf((*MockDeviceMakeRepository)(nil).GetAll), ctx)
}

// GetOrCreate mocks base method.
func (m *MockDeviceMakeRepository) GetOrCreate(ctx context.Context, makeName, logURL string) (*models.DeviceMake, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetOrCreate", ctx, makeName, logURL)
	ret0, _ := ret[0].(*models.DeviceMake)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetOrCreate indicates an expected call of GetOrCreate.
func (mr *MockDeviceMakeRepositoryMockRecorder) GetOrCreate(ctx, makeName, logURL interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetOrCreate", reflect.TypeOf((*MockDeviceMakeRepository)(nil).GetOrCreate), ctx, makeName, logURL)
}
