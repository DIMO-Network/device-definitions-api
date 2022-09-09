//go:generate mockgen -source prometheus_metric_service.go -destination mocks/prometheus_metric_service_mock.go -package mocks

package metrics

import (
	"fmt"

	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type PrometheusMetricService interface {
	Success()
	InternalError()
	BadRequestError()
}

type prometheusMetricService struct {
	svc      string
	settings *config.Settings
}

func NewMetricService(serviceName string, settings *config.Settings) PrometheusMetricService {
	return &prometheusMetricService{svc: serviceName, settings: settings}
}

func (d *prometheusMetricService) Success() {
	if d.settings.Environment == "prod" {
		c := promauto.NewCounter(prometheus.CounterOpts{
			Name: fmt.Sprintf("%s_request_success_ops_total", d.svc),
			Help: "Total successful",
		})

		defer c.Inc()
	}
}

func (d *prometheusMetricService) InternalError() {
	if d.settings.Environment == "prod" {
		c := promauto.NewCounter(prometheus.CounterOpts{
			Name: fmt.Sprintf("%s_api_request_error_ops_total", d.svc),
			Help: "Total error",
		})

		defer c.Inc()
	}

}

func (d *prometheusMetricService) BadRequestError() {
	if d.settings.Environment == "prod" {
		c := promauto.NewCounter(prometheus.CounterOpts{
			Name: fmt.Sprintf("%s_api_request_bad_ops_total", d.svc),
			Help: "Total error",
		})

		defer c.Inc()
	}
}
