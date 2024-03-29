// Code generated by MockGen. DO NOT EDIT.
// Source: smart_car_api_service.go
//
// Generated by this command:
//
//	mockgen -source smart_car_api_service.go -destination mocks/smart_car_api_service_mock.go -package mocks
//
// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	gateways "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"
	gomock "go.uber.org/mock/gomock"
)

// MockSmartCarService is a mock of SmartCarService interface.
type MockSmartCarService struct {
	ctrl     *gomock.Controller
	recorder *MockSmartCarServiceMockRecorder
}

// MockSmartCarServiceMockRecorder is the mock recorder for MockSmartCarService.
type MockSmartCarServiceMockRecorder struct {
	mock *MockSmartCarService
}

// NewMockSmartCarService creates a new mock instance.
func NewMockSmartCarService(ctrl *gomock.Controller) *MockSmartCarService {
	mock := &MockSmartCarService{ctrl: ctrl}
	mock.recorder = &MockSmartCarServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSmartCarService) EXPECT() *MockSmartCarServiceMockRecorder {
	return m.recorder
}

// GetOrCreateSmartCarIntegration mocks base method.
func (m *MockSmartCarService) GetOrCreateSmartCarIntegration(ctx context.Context) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetOrCreateSmartCarIntegration", ctx)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetOrCreateSmartCarIntegration indicates an expected call of GetOrCreateSmartCarIntegration.
func (mr *MockSmartCarServiceMockRecorder) GetOrCreateSmartCarIntegration(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetOrCreateSmartCarIntegration", reflect.TypeOf((*MockSmartCarService)(nil).GetOrCreateSmartCarIntegration), ctx)
}

// GetSmartCarVehicleData mocks base method.
func (m *MockSmartCarService) GetSmartCarVehicleData() (*gateways.SmartCarCompatibilityData, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetSmartCarVehicleData")
	ret0, _ := ret[0].(*gateways.SmartCarCompatibilityData)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetSmartCarVehicleData indicates an expected call of GetSmartCarVehicleData.
func (mr *MockSmartCarServiceMockRecorder) GetSmartCarVehicleData() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSmartCarVehicleData", reflect.TypeOf((*MockSmartCarService)(nil).GetSmartCarVehicleData))
}
