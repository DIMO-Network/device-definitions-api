// Code generated by MockGen. DO NOT EDIT.
// Source: vin_decoding_service.go

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	models "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	models0 "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	gomock "github.com/golang/mock/gomock"
)

// MockVINDecodingService is a mock of VINDecodingService interface.
type MockVINDecodingService struct {
	ctrl     *gomock.Controller
	recorder *MockVINDecodingServiceMockRecorder
}

// MockVINDecodingServiceMockRecorder is the mock recorder for MockVINDecodingService.
type MockVINDecodingServiceMockRecorder struct {
	mock *MockVINDecodingService
}

// NewMockVINDecodingService creates a new mock instance.
func NewMockVINDecodingService(ctrl *gomock.Controller) *MockVINDecodingService {
	mock := &MockVINDecodingService{ctrl: ctrl}
	mock.recorder = &MockVINDecodingServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockVINDecodingService) EXPECT() *MockVINDecodingServiceMockRecorder {
	return m.recorder
}

// GetVIN mocks base method.
func (m *MockVINDecodingService) GetVIN(vin string, dt *models0.DeviceType) (*models.VINDecodingInfoData, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetVIN", vin, dt)
	ret0, _ := ret[0].(*models.VINDecodingInfoData)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetVIN indicates an expected call of GetVIN.
func (mr *MockVINDecodingServiceMockRecorder) GetVIN(vin, dt interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetVIN", reflect.TypeOf((*MockVINDecodingService)(nil).GetVIN), vin, dt)
}
