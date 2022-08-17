// Code generated by MockGen. DO NOT EDIT.
// Source: trace_service.go

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	trace "go.opentelemetry.io/otel/trace"
)

// MockTraceService is a mock of TraceService interface.
type MockTraceService struct {
	ctrl     *gomock.Controller
	recorder *MockTraceServiceMockRecorder
}

// MockTraceServiceMockRecorder is the mock recorder for MockTraceService.
type MockTraceServiceMockRecorder struct {
	mock *MockTraceService
}

// NewMockTraceService creates a new mock instance.
func NewMockTraceService(ctrl *gomock.Controller) *MockTraceService {
	mock := &MockTraceService{ctrl: ctrl}
	mock.recorder = &MockTraceServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockTraceService) EXPECT() *MockTraceServiceMockRecorder {
	return m.recorder
}

// AddSpanError mocks base method.
func (m *MockTraceService) AddSpanError(span trace.Span, err error) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "AddSpanError", span, err)
}

// AddSpanError indicates an expected call of AddSpanError.
func (mr *MockTraceServiceMockRecorder) AddSpanError(span, err interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddSpanError", reflect.TypeOf((*MockTraceService)(nil).AddSpanError), span, err)
}

// AddSpanEvents mocks base method.
func (m *MockTraceService) AddSpanEvents(span trace.Span, name string, events map[string]string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "AddSpanEvents", span, name, events)
}

// AddSpanEvents indicates an expected call of AddSpanEvents.
func (mr *MockTraceServiceMockRecorder) AddSpanEvents(span, name, events interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddSpanEvents", reflect.TypeOf((*MockTraceService)(nil).AddSpanEvents), span, name, events)
}

// AddSpanTags mocks base method.
func (m *MockTraceService) AddSpanTags(span trace.Span, tags map[string]string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "AddSpanTags", span, tags)
}

// AddSpanTags indicates an expected call of AddSpanTags.
func (mr *MockTraceServiceMockRecorder) AddSpanTags(span, tags interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddSpanTags", reflect.TypeOf((*MockTraceService)(nil).AddSpanTags), span, tags)
}

// FailSpan mocks base method.
func (m *MockTraceService) FailSpan(span trace.Span, msg string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "FailSpan", span, msg)
}

// FailSpan indicates an expected call of FailSpan.
func (mr *MockTraceServiceMockRecorder) FailSpan(span, msg interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FailSpan", reflect.TypeOf((*MockTraceService)(nil).FailSpan), span, msg)
}
