// Code generated by MockGen. DO NOT EDIT.
// Source: device_integration_repo.go
//
// Generated by this command:
//
//	mockgen -source device_integration_repo.go -destination mocks/device_integration_repo_mock.go -package mocks
//
// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	models "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	gomock "go.uber.org/mock/gomock"
)

// MockDeviceIntegrationRepository is a mock of DeviceIntegrationRepository interface.
type MockDeviceIntegrationRepository struct {
	ctrl     *gomock.Controller
	recorder *MockDeviceIntegrationRepositoryMockRecorder
}

// MockDeviceIntegrationRepositoryMockRecorder is the mock recorder for MockDeviceIntegrationRepository.
type MockDeviceIntegrationRepositoryMockRecorder struct {
	mock *MockDeviceIntegrationRepository
}

// NewMockDeviceIntegrationRepository creates a new mock instance.
func NewMockDeviceIntegrationRepository(ctrl *gomock.Controller) *MockDeviceIntegrationRepository {
	mock := &MockDeviceIntegrationRepository{ctrl: ctrl}
	mock.recorder = &MockDeviceIntegrationRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockDeviceIntegrationRepository) EXPECT() *MockDeviceIntegrationRepositoryMockRecorder {
	return m.recorder
}

// Create mocks base method.
func (m *MockDeviceIntegrationRepository) Create(ctx context.Context, deviceDefinitionID, integrationID, region string, features []map[string]any) (*models.DeviceIntegration, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", ctx, deviceDefinitionID, integrationID, region, features)
	ret0, _ := ret[0].(*models.DeviceIntegration)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Create indicates an expected call of Create.
func (mr *MockDeviceIntegrationRepositoryMockRecorder) Create(ctx, deviceDefinitionID, integrationID, region, features any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockDeviceIntegrationRepository)(nil).Create), ctx, deviceDefinitionID, integrationID, region, features)
}
