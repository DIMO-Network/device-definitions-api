// Code generated by MockGen. DO NOT EDIT.
// Source: drivly_api_service.go

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	gateways "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"
	gomock "github.com/golang/mock/gomock"
)

// MockDrivlyAPIService is a mock of DrivlyAPIService interface.
type MockDrivlyAPIService struct {
	ctrl     *gomock.Controller
	recorder *MockDrivlyAPIServiceMockRecorder
}

// MockDrivlyAPIServiceMockRecorder is the mock recorder for MockDrivlyAPIService.
type MockDrivlyAPIServiceMockRecorder struct {
	mock *MockDrivlyAPIService
}

// NewMockDrivlyAPIService creates a new mock instance.
func NewMockDrivlyAPIService(ctrl *gomock.Controller) *MockDrivlyAPIService {
	mock := &MockDrivlyAPIService{ctrl: ctrl}
	mock.recorder = &MockDrivlyAPIServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockDrivlyAPIService) EXPECT() *MockDrivlyAPIServiceMockRecorder {
	return m.recorder
}

// GetVINInfo mocks base method.
func (m *MockDrivlyAPIService) GetVINInfo(vin string) (*gateways.DrivlyVINResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetVINInfo", vin)
	ret0, _ := ret[0].(*gateways.DrivlyVINResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetVINInfo indicates an expected call of GetVINInfo.
func (mr *MockDrivlyAPIServiceMockRecorder) GetVINInfo(vin interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetVINInfo", reflect.TypeOf((*MockDrivlyAPIService)(nil).GetVINInfo), vin)
}
