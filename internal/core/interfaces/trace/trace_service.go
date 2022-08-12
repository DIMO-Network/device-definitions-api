//go:generate mockgen -source trace_service.go -destination mocks/trace_service_mock.go -package mocks
package interfaces

import (
	"go.opentelemetry.io/otel/trace"
)

type ITraceService interface {
	AddSpanTags(span trace.Span, tags map[string]string)
	AddSpanEvents(span trace.Span, name string, events map[string]string)
	AddSpanError(span trace.Span, err error)
	FailSpan(span trace.Span, msg string)
}
