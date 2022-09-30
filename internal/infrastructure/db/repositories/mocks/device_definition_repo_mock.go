// Code generated by MockGen. DO NOT EDIT.
// Source: internal/infrastructure/db/repositories/device_definition_repo.go

// Package mock_repositories is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	models "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	repositories "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/repositories"
	gomock "github.com/golang/mock/gomock"
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

// FetchCompatibilityByMakeID mocks base method.
func (m *MockDeviceDefinitionRepository) FetchCompatibilityByMakeID(ctx context.Context, makeID string) ([]*repositories.DeviceCompatibilityModel, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FetchCompatibilityByMakeID", ctx, makeID)
	ret0, _ := ret[0].([]*repositories.DeviceCompatibilityModel)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FetchCompatibilityByMakeID indicates an expected call of FetchCompatibilityByMakeID.
func (mr *MockDeviceDefinitionRepositoryMockRecorder) FetchCompatibilityByMakeID(ctx, makeID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FetchCompatibilityByMakeID", reflect.TypeOf((*MockDeviceDefinitionRepository)(nil).FetchCompatibilityByMakeID), ctx, makeID)
}

// GetAll mocks base method.
func (m *MockDeviceDefinitionRepository) GetAll(ctx context.Context, verified bool) ([]*models.DeviceDefinition, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAll", ctx, verified)
	ret0, _ := ret[0].([]*models.DeviceDefinition)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAll indicates an expected call of GetAll.
func (mr *MockDeviceDefinitionRepositoryMockRecorder) GetAll(ctx, verified interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAll", reflect.TypeOf((*MockDeviceDefinitionRepository)(nil).GetAll), ctx, verified)
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

// GetOrCreate mocks base method.
func (m *MockDeviceDefinitionRepository) GetOrCreate(ctx context.Context, source, make, model string, year int) (*models.DeviceDefinition, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetOrCreate", ctx, source, make, model, year)
	ret0, _ := ret[0].(*models.DeviceDefinition)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetOrCreate indicates an expected call of GetOrCreate.
func (mr *MockDeviceDefinitionRepositoryMockRecorder) GetOrCreate(ctx, source, make, model, year interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetOrCreate", reflect.TypeOf((*MockDeviceDefinitionRepository)(nil).GetOrCreate), ctx, source, make, model, year)
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
