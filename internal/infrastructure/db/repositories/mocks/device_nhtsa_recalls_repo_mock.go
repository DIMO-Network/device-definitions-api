// Code generated by MockGen. DO NOT EDIT.
// Source: device_nhtsa_recalls_repo.go

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	models "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	gomock "github.com/golang/mock/gomock"
	null "github.com/volatiletech/null/v8"
)

// MockDeviceNHTSARecallsRepository is a mock of DeviceNHTSARecallsRepository interface.
type MockDeviceNHTSARecallsRepository struct {
	ctrl     *gomock.Controller
	recorder *MockDeviceNHTSARecallsRepositoryMockRecorder
}

// MockDeviceNHTSARecallsRepositoryMockRecorder is the mock recorder for MockDeviceNHTSARecallsRepository.
type MockDeviceNHTSARecallsRepositoryMockRecorder struct {
	mock *MockDeviceNHTSARecallsRepository
}

// NewMockDeviceNHTSARecallsRepository creates a new mock instance.
func NewMockDeviceNHTSARecallsRepository(ctrl *gomock.Controller) *MockDeviceNHTSARecallsRepository {
	mock := &MockDeviceNHTSARecallsRepository{ctrl: ctrl}
	mock.recorder = &MockDeviceNHTSARecallsRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockDeviceNHTSARecallsRepository) EXPECT() *MockDeviceNHTSARecallsRepositoryMockRecorder {
	return m.recorder
}

// Create mocks base method.
func (m *MockDeviceNHTSARecallsRepository) Create(ctx context.Context, deviceDefinitionID null.String, data []string, metadata null.JSON, hash []byte) (*models.DeviceNhtsaRecall, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", ctx, deviceDefinitionID, data, metadata, hash)
	ret0, _ := ret[0].(*models.DeviceNhtsaRecall)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Create indicates an expected call of Create.
func (mr *MockDeviceNHTSARecallsRepositoryMockRecorder) Create(ctx, deviceDefinitionID, data, metadata, hash interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockDeviceNHTSARecallsRepository)(nil).Create), ctx, deviceDefinitionID, data, metadata, hash)
}

// GetLastDataRecordID mocks base method.
func (m *MockDeviceNHTSARecallsRepository) GetLastDataRecordID(ctx context.Context) (*null.Int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetLastDataRecordID", ctx)
	ret0, _ := ret[0].(*null.Int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetLastDataRecordID indicates an expected call of GetLastDataRecordID.
func (mr *MockDeviceNHTSARecallsRepositoryMockRecorder) GetLastDataRecordID(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetLastDataRecordID", reflect.TypeOf((*MockDeviceNHTSARecallsRepository)(nil).GetLastDataRecordID), ctx)
}

// MatchDeviceDefinition mocks base method.
func (m *MockDeviceNHTSARecallsRepository) MatchDeviceDefinition(ctx context.Context, matchingVersion string) (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MatchDeviceDefinition", ctx, matchingVersion)
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// MatchDeviceDefinition indicates an expected call of MatchDeviceDefinition.
func (mr *MockDeviceNHTSARecallsRepositoryMockRecorder) MatchDeviceDefinition(ctx, matchingVersion interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MatchDeviceDefinition", reflect.TypeOf((*MockDeviceNHTSARecallsRepository)(nil).MatchDeviceDefinition), ctx, matchingVersion)
}