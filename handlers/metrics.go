package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	dto "github.com/prometheus/client_model/go"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

// filteredGatherer filters metrics to only include application-specific ones
type filteredGatherer struct {
	gatherer prometheus.Gatherer
}

// Gather implements prometheus.Gatherer and filters to only application metrics
func (fg *filteredGatherer) Gather() ([]*dto.MetricFamily, error) {
	metrics, err := fg.gatherer.Gather()
	if err != nil {
		return nil, err
	}

	// Filter to only include our application metrics
	filtered := make([]*dto.MetricFamily, 0)
	for _, m := range metrics {
		if m.GetName() == "echo_requests_total" || m.GetName() == "http_server_requests_seconds" {
			filtered = append(filtered, m)
		}
	}

	return filtered, nil
}

// MetricsHandler handles Prometheus metrics requests
// Only exposes application-specific metrics (echo_requests_total and http_server_requests_seconds)
func MetricsHandler(c *fiber.Ctx) error {
	// Use filtered gatherer to show only our application metrics
	handler := fasthttpadaptor.NewFastHTTPHandler(promhttp.HandlerFor(
		&filteredGatherer{gatherer: prometheus.DefaultGatherer},
		promhttp.HandlerOpts{},
	))
	handler(c.Context())
	return nil
}
