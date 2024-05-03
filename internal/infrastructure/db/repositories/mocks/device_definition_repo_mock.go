// Code generated by MockGen. DO NOT EDIT.
// Source: device_definition_repo.go
//
// Generated by this command:
//
//	mockgen -source device_definition_repo.go -destination mocks/device_definition_repo_mock.go -package mocks
//

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	sql "database/sql"
	reflect "reflect"

	models "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	repositories "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/repositories"
	null "github.com/volatiletech/null/v8"
	gomock "go.uber.org/mock/gomock"
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
func (mr *MockDeviceDefinitionRepositoryMockRecorder) CreateOrUpdate(ctx, dd, deviceStyles, deviceIntegrations any) *gomock.Call {
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
func (mr *MockDeviceDefinitionRepositoryMockRecorder) FetchDeviceCompatibility(ctx, makeID, integrationID, region, cursor, size any) *gomock.Call {
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
func (mr *MockDeviceDefinitionRepositoryMockRecorder) GetAll(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAll", reflect.TypeOf((*MockDeviceDefinitionRepository)(nil).GetAll), ctx)
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
func (mr *MockDeviceDefinitionRepositoryMockRecorder) GetByID(ctx, id any) *gomock.Call {
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
func (mr *MockDeviceDefinitionRepositoryMockRecorder) GetByMakeModelAndYears(ctx, make, model, year, loadIntegrations any) *gomock.Call {
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
func (mr *MockDeviceDefinitionRepositoryMockRecorder) GetBySlugAndYear(ctx, slug, year, loadIntegrations any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBySlugAndYear", reflect.TypeOf((*MockDeviceDefinitionRepository)(nil).GetBySlugAndYear), ctx, slug, year, loadIntegrations)
}

// GetBySlugName mocks base method.
func (m *MockDeviceDefinitionRepository) GetBySlugName(ctx context.Context, slug string, loadIntegrations bool) (*models.DeviceDefinition, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBySlugName", ctx, slug, loadIntegrations)
	ret0, _ := ret[0].(*models.DeviceDefinition)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetBySlugName indicates an expected call of GetBySlugName.
func (mr *MockDeviceDefinitionRepositoryMockRecorder) GetBySlugName(ctx, slug, loadIntegrations any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBySlugName", reflect.TypeOf((*MockDeviceDefinitionRepository)(nil).GetBySlugName), ctx, slug, loadIntegrations)
}

// GetDevicesByMakeYearRange mocks base method.
func (m *MockDeviceDefinitionRepository) GetDevicesByMakeYearRange(ctx context.Context, makeID string, yearStart, yearEnd int32) ([]*models.DeviceDefinition, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDevicesByMakeYearRange", ctx, makeID, yearStart, yearEnd)
	ret0, _ := ret[0].([]*models.DeviceDefinition)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetDevicesByMakeYearRange indicates an expected call of GetDevicesByMakeYearRange.
func (mr *MockDeviceDefinitionRepositoryMockRecorder) GetDevicesByMakeYearRange(ctx, makeID, yearStart, yearEnd any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDevicesByMakeYearRange", reflect.TypeOf((*MockDeviceDefinitionRepository)(nil).GetDevicesByMakeYearRange), ctx, makeID, yearStart, yearEnd)
}

// GetDevicesMMY mocks base method.
func (m *MockDeviceDefinitionRepository) GetDevicesMMY(ctx context.Context) ([]*repositories.DeviceMMYJoinQueryOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDevicesMMY", ctx)
	ret0, _ := ret[0].([]*repositories.DeviceMMYJoinQueryOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetDevicesMMY indicates an expected call of GetDevicesMMY.
func (mr *MockDeviceDefinitionRepositoryMockRecorder) GetDevicesMMY(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDevicesMMY", reflect.TypeOf((*MockDeviceDefinitionRepository)(nil).GetDevicesMMY), ctx)
}

// GetOrCreate mocks base method.
func (m *MockDeviceDefinitionRepository) GetOrCreate(ctx context.Context, tx *sql.Tx, source, extID, makeOrID, model string, year int, deviceTypeID string, metaData null.JSON, verified bool, hardwareTemplateID *string) (*models.DeviceDefinition, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetOrCreate", ctx, tx, source, extID, makeOrID, model, year, deviceTypeID, metaData, verified, hardwareTemplateID)
	ret0, _ := ret[0].(*models.DeviceDefinition)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetOrCreate indicates an expected call of GetOrCreate.
func (mr *MockDeviceDefinitionRepositoryMockRecorder) GetOrCreate(ctx, tx, source, extID, makeOrID, model, year, deviceTypeID, metaData, verified, hardwareTemplateID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetOrCreate", reflect.TypeOf((*MockDeviceDefinitionRepository)(nil).GetOrCreate), ctx, tx, source, extID, makeOrID, model, year, deviceTypeID, metaData, verified, hardwareTemplateID)
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
func (mr *MockDeviceDefinitionRepositoryMockRecorder) GetWithIntegrations(ctx, id any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetWithIntegrations", reflect.TypeOf((*MockDeviceDefinitionRepository)(nil).GetWithIntegrations), ctx, id)
}
