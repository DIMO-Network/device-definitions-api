// Code generated by MockGen. DO NOT EDIT.
// Source: fuel_api_service.go
//
// Generated by this command:
//
//	mockgen -source fuel_api_service.go -destination mocks/fuel_api_service_mock.go -package mocks
//
// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	gateways "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"
	gomock "go.uber.org/mock/gomock"
)

// MockFuelAPIService is a mock of FuelAPIService interface.
type MockFuelAPIService struct {
	ctrl     *gomock.Controller
	recorder *MockFuelAPIServiceMockRecorder
}

// MockFuelAPIServiceMockRecorder is the mock recorder for MockFuelAPIService.
type MockFuelAPIServiceMockRecorder struct {
	mock *MockFuelAPIService
}

// NewMockFuelAPIService creates a new mock instance.
func NewMockFuelAPIService(ctrl *gomock.Controller) *MockFuelAPIService {
	mock := &MockFuelAPIService{ctrl: ctrl}
	mock.recorder = &MockFuelAPIServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockFuelAPIService) EXPECT() *MockFuelAPIServiceMockRecorder {
	return m.recorder
}

// FetchDeviceImages mocks base method.
func (m *MockFuelAPIService) FetchDeviceImages(mk, mdl string, yr, prodID, prodFormat int) (gateways.FuelDeviceImages, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FetchDeviceImages", mk, mdl, yr, prodID, prodFormat)
	ret0, _ := ret[0].(gateways.FuelDeviceImages)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FetchDeviceImages indicates an expected call of FetchDeviceImages.
func (mr *MockFuelAPIServiceMockRecorder) FetchDeviceImages(mk, mdl, yr, prodID, prodFormat any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FetchDeviceImages", reflect.TypeOf((*MockFuelAPIService)(nil).FetchDeviceImages), mk, mdl, yr, prodID, prodFormat)
}
