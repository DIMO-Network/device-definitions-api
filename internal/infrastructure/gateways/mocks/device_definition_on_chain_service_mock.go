// Code generated by MockGen. DO NOT EDIT.
// Source: device_definition_on_chain_service.go
//
// Generated by this command:
//
//	mockgen -source device_definition_on_chain_service.go -destination mocks/device_definition_on_chain_service_mock.go -package mocks
//

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	big "math/big"
	reflect "reflect"

	contracts "github.com/DIMO-Network/device-definitions-api/internal/contracts"
	models "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	gateways "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"
	db "github.com/DIMO-Network/shared/db"
	types "github.com/volatiletech/sqlboiler/v4/types"
	gomock "go.uber.org/mock/gomock"
)

// MockDeviceDefinitionOnChainService is a mock of ddOnChainSvc interface.
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

// Create mocks base method.
func (m *MockDeviceDefinitionOnChainService) Create(ctx context.Context, mk models.DeviceMake, dd models.DeviceDefinition) (*string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", ctx, mk, dd)
	ret0, _ := ret[0].(*string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Create indicates an expected call of Create.
func (mr *MockDeviceDefinitionOnChainServiceMockRecorder) Create(ctx, mk, dd any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockDeviceDefinitionOnChainService)(nil).Create), ctx, mk, dd)
}

// Delete mocks base method.
func (m *MockDeviceDefinitionOnChainService) Delete(ctx context.Context, manufacturerName, id string) (*string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", ctx, manufacturerName, id)
	ret0, _ := ret[0].(*string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Delete indicates an expected call of Delete.
func (mr *MockDeviceDefinitionOnChainServiceMockRecorder) Delete(ctx, manufacturerName, id any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockDeviceDefinitionOnChainService)(nil).Delete), ctx, manufacturerName, id)
}

// GetDefinitionByID mocks base method.
func (m *MockDeviceDefinitionOnChainService) GetDefinitionByID(ctx context.Context, ID string, reader *db.DB) (*gateways.DeviceDefinitionTablelandModel, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDefinitionByID", ctx, ID, reader)
	ret0, _ := ret[0].(*gateways.DeviceDefinitionTablelandModel)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetDefinitionByID indicates an expected call of GetDefinitionByID.
func (mr *MockDeviceDefinitionOnChainServiceMockRecorder) GetDefinitionByID(ctx, ID, reader any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDefinitionByID", reflect.TypeOf((*MockDeviceDefinitionOnChainService)(nil).GetDefinitionByID), ctx, ID, reader)
}

// GetDefinitionTableland mocks base method.
func (m *MockDeviceDefinitionOnChainService) GetDefinitionTableland(ctx context.Context, manufacturerID *big.Int, ID string) (*gateways.DeviceDefinitionTablelandModel, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDefinitionTableland", ctx, manufacturerID, ID)
	ret0, _ := ret[0].(*gateways.DeviceDefinitionTablelandModel)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetDefinitionTableland indicates an expected call of GetDefinitionTableland.
func (mr *MockDeviceDefinitionOnChainServiceMockRecorder) GetDefinitionTableland(ctx, manufacturerID, ID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDefinitionTableland", reflect.TypeOf((*MockDeviceDefinitionOnChainService)(nil).GetDefinitionTableland), ctx, manufacturerID, ID)
}

// GetDeviceDefinitionByID mocks base method.
func (m *MockDeviceDefinitionOnChainService) GetDeviceDefinitionByID(ctx context.Context, manufacturerID *big.Int, ID string) (*models.DeviceDefinition, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDeviceDefinitionByID", ctx, manufacturerID, ID)
	ret0, _ := ret[0].(*models.DeviceDefinition)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetDeviceDefinitionByID indicates an expected call of GetDeviceDefinitionByID.
func (mr *MockDeviceDefinitionOnChainServiceMockRecorder) GetDeviceDefinitionByID(ctx, manufacturerID, ID any) *gomock.Call {
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
func (mr *MockDeviceDefinitionOnChainServiceMockRecorder) GetDeviceDefinitions(ctx, manufacturerID, ID, model, year, pageIndex, pageSize any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDeviceDefinitions", reflect.TypeOf((*MockDeviceDefinitionOnChainService)(nil).GetDeviceDefinitions), ctx, manufacturerID, ID, model, year, pageIndex, pageSize)
}

// Update mocks base method.
func (m *MockDeviceDefinitionOnChainService) Update(ctx context.Context, manufacturerName string, input contracts.DeviceDefinitionUpdateInput) (*string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Update", ctx, manufacturerName, input)
	ret0, _ := ret[0].(*string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Update indicates an expected call of Update.
func (mr *MockDeviceDefinitionOnChainServiceMockRecorder) Update(ctx, manufacturerName, input any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockDeviceDefinitionOnChainService)(nil).Update), ctx, manufacturerName, input)
}
