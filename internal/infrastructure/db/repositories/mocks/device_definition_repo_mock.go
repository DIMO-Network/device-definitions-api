// Code generated by MockGen. DO NOT EDIT.
// Source: device_definition_repo.go

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	models "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	gomock "github.com/golang/mock/gomock"
	null "github.com/volatiletech/null/v8"
)

// MockDeviceDefinitionRepository is a mock of DeviceDefinitionRepository interface.
type MockDeviceDefinitionRepository struct {
	ctrl     *gomock.Controller
	recorder *MockDeviceDefinitionRepositoryMockRecorder
}

// MockDeviceDefinitionRepositoryMockRecorder is the mock recorder for MockDeviceDefinitionRepository.
type MockDeviceDefinitionRepositoryMockRecorder struct {
	mock *MockDeviceDefinitionRepository
}

// NewMockDeviceDefinitionRepository creates a new mock instance.
func NewMockDeviceDefinitionRepository(ctrl *gomock.Controller) *MockDeviceDefinitionRepository {
	mock := &MockDeviceDefinitionRepository{ctrl: ctrl}
	mock.recorder = &MockDeviceDefinitionRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockDeviceDefinitionRepository) EXPECT() *MockDeviceDefinitionRepositoryMockRecorder {
	return m.recorder
}

// CreateOrUpdate mocks base method.
func (m *MockDeviceDefinitionRepository) CreateOrUpdate(ctx context.Context, dd *models.DeviceDefinition, deviceStyles []*models.DeviceStyle, deviceIntegrations []*models.DeviceIntegration) (*models.DeviceDefinition, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateOrUpdate", ctx, dd, deviceStyles, deviceIntegrations)
	ret0, _ := ret[0].(*models.DeviceDefinition)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateOrUpdate indicates an expected call of CreateOrUpdate.
func (mr *MockDeviceDefinitionRepositoryMockRecorder) CreateOrUpdate(ctx, dd, deviceStyles, deviceIntegrations interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateOrUpdate", reflect.TypeOf((*MockDeviceDefinitionRepository)(nil).CreateOrUpdate), ctx, dd, deviceStyles, deviceIntegrations)
}

// FetchDeviceCompatibility mocks base method.
func (m *MockDeviceDefinitionRepository) FetchDeviceCompatibility(ctx context.Context, makeID, integrationID, region, cursor string, size int64) (models.DeviceDefinitionSlice, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FetchDeviceCompatibility", ctx, makeID, integrationID, region, cursor, size)
	ret0, _ := ret[0].(models.DeviceDefinitionSlice)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FetchDeviceCompatibility indicates an expected call of FetchDeviceCompatibility.
func (mr *MockDeviceDefinitionRepositoryMockRecorder) FetchDeviceCompatibility(ctx, makeID, integrationID, region, cursor, size interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FetchDeviceCompatibility", reflect.TypeOf((*MockDeviceDefinitionRepository)(nil).FetchDeviceCompatibility), ctx, makeID, integrationID, region, cursor, size)
}

// GetAll mocks base method.
func (m *MockDeviceDefinitionRepository) GetAll(ctx context.Context) ([]*models.DeviceDefinition, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAll", ctx)
	ret0, _ := ret[0].([]*models.DeviceDefinition)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAll indicates an expected call of GetAll.
func (mr *MockDeviceDefinitionRepositoryMockRecorder) GetAll(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAll", reflect.TypeOf((*MockDeviceDefinitionRepository)(nil).GetAll), ctx)
}

// GetAllDevicesMMY mocks base method.
func (m *MockDeviceDefinitionRepository) GetAllDevicesMMY(ctx context.Context) ([]*models.DeviceDefinition, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAllDevicesMMY", ctx)
	ret0, _ := ret[0].([]*models.DeviceDefinition)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAllDevicesMMY indicates an expected call of GetAllDevicesMMY.
func (mr *MockDeviceDefinitionRepositoryMockRecorder) GetAllDevicesMMY(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAllDevicesMMY", reflect.TypeOf((*MockDeviceDefinitionRepository)(nil).GetAllDevicesMMY), ctx)
}

// GetByID mocks base method.
func (m *MockDeviceDefinitionRepository) GetByID(ctx context.Context, id string) (*models.DeviceDefinition, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetByID", ctx, id)
	ret0, _ := ret[0].(*models.DeviceDefinition)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetByID indicates an expected call of GetByID.
func (mr *MockDeviceDefinitionRepositoryMockRecorder) GetByID(ctx, id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByID", reflect.TypeOf((*MockDeviceDefinitionRepository)(nil).GetByID), ctx, id)
}

// GetByMakeModelAndYears mocks base method.
func (m *MockDeviceDefinitionRepository) GetByMakeModelAndYears(ctx context.Context, make, model string, year int, loadIntegrations bool) (*models.DeviceDefinition, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetByMakeModelAndYears", ctx, make, model, year, loadIntegrations)
	ret0, _ := ret[0].(*models.DeviceDefinition)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetByMakeModelAndYears indicates an expected call of GetByMakeModelAndYears.
func (mr *MockDeviceDefinitionRepositoryMockRecorder) GetByMakeModelAndYears(ctx, make, model, year, loadIntegrations interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByMakeModelAndYears", reflect.TypeOf((*MockDeviceDefinitionRepository)(nil).GetByMakeModelAndYears), ctx, make, model, year, loadIntegrations)
}

// GetBySlugAndYear mocks base method.
func (m *MockDeviceDefinitionRepository) GetBySlugAndYear(ctx context.Context, slug string, year int, loadIntegrations bool) (*models.DeviceDefinition, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBySlugAndYear", ctx, slug, year, loadIntegrations)
	ret0, _ := ret[0].(*models.DeviceDefinition)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetBySlugAndYear indicates an expected call of GetBySlugAndYear.
func (mr *MockDeviceDefinitionRepositoryMockRecorder) GetBySlugAndYear(ctx, slug, year, loadIntegrations interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBySlugAndYear", reflect.TypeOf((*MockDeviceDefinitionRepository)(nil).GetBySlugAndYear), ctx, slug, year, loadIntegrations)
}

// GetOrCreate mocks base method.
func (m *MockDeviceDefinitionRepository) GetOrCreate(ctx context.Context, source, extID, makeOrID, model string, year int, deviceTypeID string, metaData null.JSON, verified bool, hardwareTemplateID string) (*models.DeviceDefinition, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetOrCreate", ctx, source, extID, makeOrID, model, year, deviceTypeID, metaData, verified, hardwareTemplateID)
	ret0, _ := ret[0].(*models.DeviceDefinition)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetOrCreate indicates an expected call of GetOrCreate.
func (mr *MockDeviceDefinitionRepositoryMockRecorder) GetOrCreate(ctx, source, extID, makeOrID, model, year, deviceTypeID, metaData, verified, hardwareTemplateID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetOrCreate", reflect.TypeOf((*MockDeviceDefinitionRepository)(nil).GetOrCreate), ctx, source, extID, makeOrID, model, year, deviceTypeID, metaData, verified, hardwareTemplateID)
}

// GetWithIntegrations mocks base method.
func (m *MockDeviceDefinitionRepository) GetWithIntegrations(ctx context.Context, id string) (*models.DeviceDefinition, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetWithIntegrations", ctx, id)
	ret0, _ := ret[0].(*models.DeviceDefinition)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetWithIntegrations indicates an expected call of GetWithIntegrations.
func (mr *MockDeviceDefinitionRepositoryMockRecorder) GetWithIntegrations(ctx, id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetWithIntegrations", reflect.TypeOf((*MockDeviceDefinitionRepository)(nil).GetWithIntegrations), ctx, id)
}
