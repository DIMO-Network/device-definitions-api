// Code generated by MockGen. DO NOT EDIT.
// Source: vincario_api_service.go

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockVincarioAPIService is a mock of VincarioAPIService interface.
type MockVincarioAPIService struct {
	ctrl     *gomock.Controller
	recorder *MockVincarioAPIServiceMockRecorder
}

// MockVincarioAPIServiceMockRecorder is the mock recorder for MockVincarioAPIService.
type MockVincarioAPIServiceMockRecorder struct {
	mock *MockVincarioAPIService
}

// NewMockVincarioAPIService creates a new mock instance.
func NewMockVincarioAPIService(ctrl *gomock.Controller) *MockVincarioAPIService {
	mock := &MockVincarioAPIService{ctrl: ctrl}
	mock.recorder = &MockVincarioAPIServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockVincarioAPIService) EXPECT() *MockVincarioAPIServiceMockRecorder {
	return m.recorder
}

// DecodeVIN mocks base method.
func (m *MockVincarioAPIService) DecodeVIN(vin string) (any, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DecodeVIN", vin)
	ret0, _ := ret[0].(any)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DecodeVIN indicates an expected call of DecodeVIN.
func (mr *MockVincarioAPIServiceMockRecorder) DecodeVIN(vin interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DecodeVIN", reflect.TypeOf((*MockVincarioAPIService)(nil).DecodeVIN), vin)
}
