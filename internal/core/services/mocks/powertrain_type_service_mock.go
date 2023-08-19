// Code generated by MockGen. DO NOT EDIT.
// Source: powertrain_type_service.go

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	null "github.com/volatiletech/null/v8"
)

// MockPowerTrainTypeService is a mock of PowerTrainTypeService interface.
type MockPowerTrainTypeService struct {
	ctrl     *gomock.Controller
	recorder *MockPowerTrainTypeServiceMockRecorder
}

// MockPowerTrainTypeServiceMockRecorder is the mock recorder for MockPowerTrainTypeService.
type MockPowerTrainTypeServiceMockRecorder struct {
	mock *MockPowerTrainTypeService
}

// NewMockPowerTrainTypeService creates a new mock instance.
func NewMockPowerTrainTypeService(ctrl *gomock.Controller) *MockPowerTrainTypeService {
	mock := &MockPowerTrainTypeService{ctrl: ctrl}
	mock.recorder = &MockPowerTrainTypeServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockPowerTrainTypeService) EXPECT() *MockPowerTrainTypeServiceMockRecorder {
	return m.recorder
}

// ResolvePowerTrainType mocks base method.
func (m *MockPowerTrainTypeService) ResolvePowerTrainType(ctx context.Context, makeSlug, modelSlug string, definitionID *string, drivlyData, vincarioData null.JSON) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ResolvePowerTrainType", ctx, makeSlug, modelSlug, definitionID, drivlyData, vincarioData)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ResolvePowerTrainType indicates an expected call of ResolvePowerTrainType.
func (mr *MockPowerTrainTypeServiceMockRecorder) ResolvePowerTrainType(ctx, makeSlug, modelSlug, definitionID, drivlyData, vincarioData interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ResolvePowerTrainType", reflect.TypeOf((*MockPowerTrainTypeService)(nil).ResolvePowerTrainType), ctx, makeSlug, modelSlug, definitionID, drivlyData, vincarioData)
}
