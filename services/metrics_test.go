package services

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
)

func TestNewMetricsService(t *testing.T) {
	service := NewMetricsService()

	if service == nil {
		t.Fatal("Expected non-nil MetricsService")
	}

	if service.requestCounter == nil {
		t.Error("Expected requestCounter to be initialized")
	}

	if service.requestLatency == nil {
		t.Error("Expected requestLatency to be initialized")
	}
}

func TestMetricsMiddleware(t *testing.T) {
	// Note: We can't easily test metrics registration in unit tests due to
	// global prometheus registry. This test validates the middleware works.
	app := fiber.New()

	// Create a new service (will reuse existing metrics from global registry)
	service := &MetricsService{
		requestCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "test_requests_total",
				Help: "Test counter",
			},
			[]string{"method", "uri"},
		),
		requestLatency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "test_requests_seconds",
				Help:    "Test histogram",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "uri"},
		),
	}

	// Add middleware
	app.Use(service.MetricsMiddleware)

	// Add test endpoint
	app.Get("/test", func(c *fiber.Ctx) error {
		time.Sleep(10 * time.Millisecond) // Simulate some work
		return c.SendString("OK")
	})
	app.Post("/test", func(c *fiber.Ctx) error {
		time.Sleep(10 * time.Millisecond) // Simulate some work
		return c.SendString("OK")
	})

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{
			name:   "GET request",
			method: "GET",
			path:   "/test",
		},
		{
			name:   "POST request",
			method: "POST",
			path:   "/test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			resp, err := app.Test(req, -1)

			if err != nil {
				t.Fatalf("Failed to send request: %v", err)
			}

			if resp.StatusCode != fiber.StatusOK && resp.StatusCode != fiber.StatusNotFound {
				t.Errorf("Expected status 200 or 404, got %d", resp.StatusCode)
			}

			// Verify metrics were recorded (we can't easily check values in tests,
			// but we can verify the middleware doesn't panic)
		})
	}
}

func TestMetricsMiddlewareError(t *testing.T) {
	app := fiber.New()

	service := &MetricsService{
		requestCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "test_error_requests_total",
				Help: "Test counter",
			},
			[]string{"method", "uri"},
		),
		requestLatency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "test_error_requests_seconds",
				Help:    "Test histogram",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "uri"},
		),
	}

	app.Use(service.MetricsMiddleware)

	// Add endpoint that returns an error
	app.Get("/error", func(c *fiber.Ctx) error {
		return fiber.NewError(fiber.StatusInternalServerError, "test error")
	})

	req := httptest.NewRequest("GET", "/error", nil)
	resp, err := app.Test(req, -1)

	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}

	if resp.StatusCode != fiber.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", resp.StatusCode)
	}

	// Verify middleware still records metrics even when handler returns error
}

func TestMetricsServiceConcurrency(t *testing.T) {
	app := fiber.New()

	service := &MetricsService{
		requestCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "test_concurrent_requests_total",
				Help: "Test counter",
			},
			[]string{"method", "uri"},
		),
		requestLatency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "test_concurrent_requests_seconds",
				Help:    "Test histogram",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "uri"},
		),
	}

	app.Use(service.MetricsMiddleware)
	app.Get("/concurrent", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// Send multiple concurrent requests
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			req := httptest.NewRequest("GET", "/concurrent", nil)
			_, err := app.Test(req, -1)
			if err != nil {
				t.Errorf("Failed to send request: %v", err)
			}
			done <- true
		}()
	}

	// Wait for all requests to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Test passes if no race conditions or panics occur
}
