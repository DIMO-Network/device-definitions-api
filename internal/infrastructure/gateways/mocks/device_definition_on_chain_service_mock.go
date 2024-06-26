// Code generated by MockGen. DO NOT EDIT.
// Source: device_definition_on_chain_service.go

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	models "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	gomock "go.uber.org/mock/gomock"
	types "github.com/volatiletech/sqlboiler/v4/types"
)

// MockDeviceDefinitionOnChainService is a mock of DeviceDefinitionOnChainService interface.
type MockDeviceDefinitionOnChainService struct {
	ctrl     *gomock.Controller
	recorder *MockDeviceDefinitionOnChainServiceMockRecorder
}

// MockDeviceDefinitionOnChainServiceMockRecorder is the mock recorder for MockDeviceDefinitionOnChainService.
type MockDeviceDefinitionOnChainServiceMockRecorder struct {
	mock *MockDeviceDefinitionOnChainService
}

// NewMockDeviceDefinitionOnChainService creates a new mock instance.
func NewMockDeviceDefinitionOnChainService(ctrl *gomock.Controller) *MockDeviceDefinitionOnChainService {
	mock := &MockDeviceDefinitionOnChainService{ctrl: ctrl}
	mock.recorder = &MockDeviceDefinitionOnChainServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockDeviceDefinitionOnChainService) EXPECT() *MockDeviceDefinitionOnChainServiceMockRecorder {
	return m.recorder
}

// CreateOrUpdate mocks base method.
func (m *MockDeviceDefinitionOnChainService) CreateOrUpdate(ctx context.Context, make models.DeviceMake, dd models.DeviceDefinition) (*string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateOrUpdate", ctx, make, dd)
	ret0, _ := ret[0].(*string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateOrUpdate indicates an expected call of CreateOrUpdate.
func (mr *MockDeviceDefinitionOnChainServiceMockRecorder) CreateOrUpdate(ctx, make, dd interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateOrUpdate", reflect.TypeOf((*MockDeviceDefinitionOnChainService)(nil).CreateOrUpdate), ctx, make, dd)
}

// GetDeviceDefinitionByID mocks base method.
func (m *MockDeviceDefinitionOnChainService) GetDeviceDefinitionByID(ctx context.Context, manufacturerID types.NullDecimal, ID string) (*models.DeviceDefinition, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDeviceDefinitionByID", ctx, manufacturerID, ID)
	ret0, _ := ret[0].(*models.DeviceDefinition)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetDeviceDefinitionByID indicates an expected call of GetDeviceDefinitionByID.
func (mr *MockDeviceDefinitionOnChainServiceMockRecorder) GetDeviceDefinitionByID(ctx, manufacturerID, ID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDeviceDefinitionByID", reflect.TypeOf((*MockDeviceDefinitionOnChainService)(nil).GetDeviceDefinitionByID), ctx, manufacturerID, ID)
}

// GetDeviceDefinitions mocks base method.
func (m *MockDeviceDefinitionOnChainService) GetDeviceDefinitions(ctx context.Context, manufacturerID types.NullDecimal, ID, model string, year int, pageIndex, pageSize int32) ([]*models.DeviceDefinition, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDeviceDefinitions", ctx, manufacturerID, ID, model, year, pageIndex, pageSize)
	ret0, _ := ret[0].([]*models.DeviceDefinition)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetDeviceDefinitions indicates an expected call of GetDeviceDefinitions.
func (mr *MockDeviceDefinitionOnChainServiceMockRecorder) GetDeviceDefinitions(ctx, manufacturerID, ID, model, year, pageIndex, pageSize interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDeviceDefinitions", reflect.TypeOf((*MockDeviceDefinitionOnChainService)(nil).GetDeviceDefinitions), ctx, manufacturerID, ID, model, year, pageIndex, pageSize)
}
