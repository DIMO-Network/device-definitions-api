package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const serviceName = "device_definitions_api_"

var (
	Success = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: serviceName + "request_success_ops_total",
		Help: "Total execution",
	}, []string{"method"})

	InternalError = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: serviceName + "request_error_ops_total",
		Help: "Total execution",
	}, []string{"method"})

	BadRequestError = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: serviceName + "request_bad_ops_total",
		Help: "Total execution",
	}, []string{"method"})

	ConflictRequestError = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: serviceName + "request_conflict_ops_total",
		Help: "Total execution",
	}, []string{"method"})

	NotFoundRequestError = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: serviceName + "request_notfound_ops_total",
		Help: "Total execution",
	}, []string{"method"})

	GRPCRequestCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: serviceName + "grpc_request_count",
			Help: "The total number of requests served by the GRPC Server",
		},
		[]string{"method", "status"},
	)

	GRPCResponseTime = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    serviceName + "grpc_response_time",
			Help:    "The response time distribution of the GRPC Server",
			Buckets: []float64{0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"method", "status"},
	)

	GRPCPanicsCount = promauto.NewCounter(prometheus.CounterOpts{
		Name: serviceName + "panics_total",
		Help: "Total Panics recovered",
	})

	HTTPRequestCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: serviceName + "http_request_count",
			Help: "The total number of requests served by the Http Server",
		},
		[]string{"method", "path", "status"},
	)

	HTTPResponseTime = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    serviceName + "http_response_time",
			Help:    "The response time distribution of the Http Server",
			Buckets: []float64{0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"method", "path", "status"},
	)
)
