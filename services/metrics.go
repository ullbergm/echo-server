package services

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// MetricsService handles Prometheus metrics
type MetricsService struct {
	requestCounter *prometheus.CounterVec
	requestLatency *prometheus.HistogramVec
}

// NewMetricsService creates a new metrics service
func NewMetricsService() *MetricsService {
	return &MetricsService{
		requestCounter: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "echo_requests_total",
				Help: "Total number of echo requests",
			},
			[]string{"method", "uri"},
		),
		requestLatency: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_server_requests_seconds",
				Help:    "HTTP request latency in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "uri"},
		),
	}
}

// MetricsMiddleware records metrics for each request
func (m *MetricsService) MetricsMiddleware(c *fiber.Ctx) error {
	start := time.Now()

	// Continue to next handler
	err := c.Next()

	// Record metrics
	duration := time.Since(start).Seconds()
	method := c.Method()
	uri := c.Path()

	m.requestCounter.WithLabelValues(method, uri).Inc()
	m.requestLatency.WithLabelValues(method, uri).Observe(duration)

	return err
}
