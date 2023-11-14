// Code generated by MockGen. DO NOT EDIT.
// Source: elastic_search_api_service.go

// Package mocks is a generated GoMock package.
package mocks

import (
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/elastic"
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockElasticSearchService is a mock of ElasticSearchService interface.
type MockElasticSearchService struct {
	ctrl     *gomock.Controller
	recorder *MockElasticSearchServiceMockRecorder
}

// MockElasticSearchServiceMockRecorder is the mock recorder for MockElasticSearchService.
type MockElasticSearchServiceMockRecorder struct {
	mock *MockElasticSearchService
}

// NewMockElasticSearchService creates a new mock instance.
func NewMockElasticSearchService(ctrl *gomock.Controller) *MockElasticSearchService {
	mock := &MockElasticSearchService{ctrl: ctrl}
	mock.recorder = &MockElasticSearchServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockElasticSearchService) EXPECT() *MockElasticSearchServiceMockRecorder {
	return m.recorder
}

// AddSourceEngineToMetaEngine mocks base method.
func (m *MockElasticSearchService) AddSourceEngineToMetaEngine(sourceName, metaName string) (*elastic.EngineDetail, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddSourceEngineToMetaEngine", sourceName, metaName)
	ret0, _ := ret[0].(*elastic.EngineDetail)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AddSourceEngineToMetaEngine indicates an expected call of AddSourceEngineToMetaEngine.
func (mr *MockElasticSearchServiceMockRecorder) AddSourceEngineToMetaEngine(sourceName, metaName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddSourceEngineToMetaEngine", reflect.TypeOf((*MockElasticSearchService)(nil).AddSourceEngineToMetaEngine), sourceName, metaName)
}

// CreateDocuments mocks base method.
func (m *MockElasticSearchService) CreateDocuments(docs []elastic.DeviceDefinitionSearchDoc, engineName string) ([]elastic.CreateDocsResp, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateDocuments", docs, engineName)
	ret0, _ := ret[0].([]elastic.CreateDocsResp)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateDocuments indicates an expected call of CreateDocuments.
func (mr *MockElasticSearchServiceMockRecorder) CreateDocuments(docs, engineName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateDocuments", reflect.TypeOf((*MockElasticSearchService)(nil).CreateDocuments), docs, engineName)
}

// CreateDocumentsBatched mocks base method.
func (m *MockElasticSearchService) CreateDocumentsBatched(docs []elastic.DeviceDefinitionSearchDoc, engineName string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateDocumentsBatched", docs, engineName)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateDocumentsBatched indicates an expected call of CreateDocumentsBatched.
func (mr *MockElasticSearchServiceMockRecorder) CreateDocumentsBatched(docs, engineName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateDocumentsBatched", reflect.TypeOf((*MockElasticSearchService)(nil).CreateDocumentsBatched), docs, engineName)
}

// CreateEngine mocks base method.
func (m *MockElasticSearchService) CreateEngine(name string, metaSource *string) (*elastic.EngineDetail, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateEngine", name, metaSource)
	ret0, _ := ret[0].(*elastic.EngineDetail)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateEngine indicates an expected call of CreateEngine.
func (mr *MockElasticSearchServiceMockRecorder) CreateEngine(name, metaSource interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateEngine", reflect.TypeOf((*MockElasticSearchService)(nil).CreateEngine), name, metaSource)
}

// DeleteEngine mocks base method.
func (m *MockElasticSearchService) DeleteEngine(name string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteEngine", name)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteEngine indicates an expected call of DeleteEngine.
func (mr *MockElasticSearchServiceMockRecorder) DeleteEngine(name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteEngine", reflect.TypeOf((*MockElasticSearchService)(nil).DeleteEngine), name)
}

// GetEngines mocks base method.
func (m *MockElasticSearchService) GetEngines() (*elastic.GetEnginesResp, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetEngines")
	ret0, _ := ret[0].(*elastic.GetEnginesResp)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetEngines indicates an expected call of GetEngines.
func (mr *MockElasticSearchServiceMockRecorder) GetEngines() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetEngines", reflect.TypeOf((*MockElasticSearchService)(nil).GetEngines))
}

// GetMetaEngineName mocks base method.
func (m *MockElasticSearchService) GetMetaEngineName() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetMetaEngineName")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetMetaEngineName indicates an expected call of GetMetaEngineName.
func (mr *MockElasticSearchServiceMockRecorder) GetMetaEngineName() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetMetaEngineName", reflect.TypeOf((*MockElasticSearchService)(nil).GetMetaEngineName))
}

// LoadDeviceDefinitions mocks base method.
func (m *MockElasticSearchService) LoadDeviceDefinitions() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LoadDeviceDefinitions")
	ret0, _ := ret[0].(error)
	return ret0
}

// LoadDeviceDefinitions indicates an expected call of LoadDeviceDefinitions.
func (mr *MockElasticSearchServiceMockRecorder) LoadDeviceDefinitions() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LoadDeviceDefinitions", reflect.TypeOf((*MockElasticSearchService)(nil).LoadDeviceDefinitions))
}

// RemoveSourceEngine mocks base method.
func (m *MockElasticSearchService) RemoveSourceEngine(sourceName, metaName string) (*elastic.EngineDetail, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RemoveSourceEngine", sourceName, metaName)
	ret0, _ := ret[0].(*elastic.EngineDetail)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// RemoveSourceEngine indicates an expected call of RemoveSourceEngine.
func (mr *MockElasticSearchServiceMockRecorder) RemoveSourceEngine(sourceName, metaName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemoveSourceEngine", reflect.TypeOf((*MockElasticSearchService)(nil).RemoveSourceEngine), sourceName, metaName)
}

// UpdateSearchSettingsForDeviceDefs mocks base method.
func (m *MockElasticSearchService) UpdateSearchSettingsForDeviceDefs(engineName string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateSearchSettingsForDeviceDefs", engineName)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateSearchSettingsForDeviceDefs indicates an expected call of UpdateSearchSettingsForDeviceDefs.
func (mr *MockElasticSearchServiceMockRecorder) UpdateSearchSettingsForDeviceDefs(engineName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateSearchSettingsForDeviceDefs", reflect.TypeOf((*MockElasticSearchService)(nil).UpdateSearchSettingsForDeviceDefs), engineName)
}
