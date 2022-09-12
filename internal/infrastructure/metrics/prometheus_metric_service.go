//go:generate mockgen -source prometheus_metric_service.go -destination mocks/prometheus_metric_service_mock.go -package mocks

package metrics

import (
	"fmt"
	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"regexp"
	"strings"
)

type PrometheusMetricService interface {
	Success(label string)
	InternalError(label string)
	BadRequestError(label string)
}

type prometheusMetricService struct {
	svc      string
	settings *config.Settings
}

func NewMetricService(metricPrefix string, settings *config.Settings) PrometheusMetricService {
	match, _ := regexp.MatchString("^[a-zA-Z_:][a-zA-Z0-9_:]*", metricPrefix)
	if !match || strings.Contains(metricPrefix, "-") { // go has issue with not matching hyphen in regex
		panic("invalid metric name prefix")
	}
	return &prometheusMetricService{svc: metricPrefix, settings: settings}
}

func (d *prometheusMetricService) Success(label string) {
	c := promauto.NewCounterVec(prometheus.CounterOpts{
		Name: fmt.Sprintf("%s_request_success_ops_total", d.svc),
		Help: "Total successful",
	}, []string{"method"})

	defer c.With(prometheus.Labels{"method": label}).Inc()
}

func (d *prometheusMetricService) InternalError(label string) {
	c := promauto.NewCounterVec(prometheus.CounterOpts{
		Name: fmt.Sprintf("%s_api_request_error_ops_total", d.svc),
		Help: "Total error",
	}, []string{"method"})

	defer c.With(prometheus.Labels{"method": label}).Inc()
}

func (d *prometheusMetricService) BadRequestError(label string) {
	c := promauto.NewCounterVec(prometheus.CounterOpts{
		Name: fmt.Sprintf("%s_api_request_bad_ops_total", d.svc),
		Help: "Total error",
	}, []string{"method"})

	defer c.With(prometheus.Labels{"method": label}).Inc()
}
