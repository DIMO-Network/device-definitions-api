// Code generated by MockGen. DO NOT EDIT.
// Source: identity_api.go
//
// Generated by this command:
//
//	mockgen -source identity_api.go -destination mocks/identity_api_mock.go -package mocks
//

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	models "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	gomock "go.uber.org/mock/gomock"
)

// MockIdentityAPI is a mock of IdentityAPI interface.
type MockIdentityAPI struct {
	ctrl     *gomock.Controller
	recorder *MockIdentityAPIMockRecorder
	isgomock struct{}
}

// MockIdentityAPIMockRecorder is the mock recorder for MockIdentityAPI.
type MockIdentityAPIMockRecorder struct {
	mock *MockIdentityAPI
}

// NewMockIdentityAPI creates a new mock instance.
func NewMockIdentityAPI(ctrl *gomock.Controller) *MockIdentityAPI {
	mock := &MockIdentityAPI{ctrl: ctrl}
	mock.recorder = &MockIdentityAPIMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockIdentityAPI) EXPECT() *MockIdentityAPIMockRecorder {
	return m.recorder
}

// GetManufacturer mocks base method.
func (m *MockIdentityAPI) GetManufacturer(slug string) (*models.Manufacturer, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetManufacturer", slug)
	ret0, _ := ret[0].(*models.Manufacturer)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetManufacturer indicates an expected call of GetManufacturer.
func (mr *MockIdentityAPIMockRecorder) GetManufacturer(slug any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetManufacturer", reflect.TypeOf((*MockIdentityAPI)(nil).GetManufacturer), slug)
}

// GetManufacturers mocks base method.
func (m *MockIdentityAPI) GetManufacturers() ([]models.Manufacturer, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetManufacturers")
	ret0, _ := ret[0].([]models.Manufacturer)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetManufacturers indicates an expected call of GetManufacturers.
func (mr *MockIdentityAPIMockRecorder) GetManufacturers() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetManufacturers", reflect.TypeOf((*MockIdentityAPI)(nil).GetManufacturers))
}
