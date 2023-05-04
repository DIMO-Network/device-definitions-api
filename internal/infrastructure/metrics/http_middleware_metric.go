package metrics

import (
	"strconv"

	"time"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
)

func HTTPMetricsPrometheusMiddleware(c *fiber.Ctx) error {
	start := time.Now()
	method := c.Route().Method

	err := c.Next()
	status := fiber.StatusInternalServerError
	if err != nil {
		if e, ok := err.(*fiber.Error); ok {
			// Get correct error code from fiber.Error type
			status = e.Code
		}

		if _, ok := err.(*exceptions.ValidationError); ok {
			status = fiber.StatusBadRequest
		}

		if _, ok := err.(*exceptions.NotFoundError); ok {
			status = fiber.StatusNotFound
		}

		if _, ok := err.(*exceptions.ConflictError); ok {
			status = fiber.StatusConflict
		}
	} else {
		status = c.Response().StatusCode()
	}

	path := c.Route().Name
	statusCode := strconv.Itoa(status)

	HTTPRequestCount.WithLabelValues(method, path, statusCode).Inc()

	defer func() {
		HTTPResponseTime.With(prometheus.Labels{
			"method": method,
			"path":   path,
			"status": statusCode,
		}).Observe(time.Since(start).Seconds())
	}()

	return err
}
