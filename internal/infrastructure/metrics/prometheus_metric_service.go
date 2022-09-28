package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	Success = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "request_success_ops_total",
		Help: "Total execution",
	}, []string{"method"})

	InternalError = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "request_error_ops_total",
		Help: "Total execution",
	}, []string{"method"})

	BadRequestError = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "request_bad_ops_total",
		Help: "Total execution",
	}, []string{"method"})

	ConflictRequestError = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "request_conflict_ops_total",
		Help: "Total execution",
	}, []string{"method"})

	NotFoundRequestError = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "request_notfound_ops_total",
		Help: "Total execution",
	}, []string{"method"})
)
