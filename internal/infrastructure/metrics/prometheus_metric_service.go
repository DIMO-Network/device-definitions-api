//go:generate mockgen -source prometheus_metric_service.go -destination mocks/prometheus_metric_service_mock.go -package mocks

package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type PrometheusMetricService interface {
	Success()
	InternalError()
	BadRequestError()
}

type prometheusMetricService struct {
}

func NewMetricService() PrometheusMetricService {
	return &prometheusMetricService{}
}

func (d *prometheusMetricService) Success() {
	c := promauto.NewCounter(prometheus.CounterOpts{
		Name: "device_definitions_api_request_success_ops_total",
		Help: "Total successful",
	})

	defer c.Inc()
}

func (d *prometheusMetricService) InternalError() {
	c := promauto.NewCounter(prometheus.CounterOpts{
		Name: "device_definitions_api_request_error_ops_total",
		Help: "Total error",
	})

	defer c.Inc()
}

func (d *prometheusMetricService) BadRequestError() {
	c := promauto.NewCounter(prometheus.CounterOpts{
		Name: "device_definitions_api_request_bad_ops_total",
		Help: "Total error",
	})

	defer c.Inc()
}
