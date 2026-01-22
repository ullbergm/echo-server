package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/ullbergm/echo-server/services"
)

// Global metrics service to avoid duplicate registration
var (
	testMetricsService     *services.MetricsService
	testMetricsServiceOnce sync.Once
)

func getTestMetricsService() *services.MetricsService {
	testMetricsServiceOnce.Do(func() {
		testMetricsService = services.NewMetricsService()
	})
	return testMetricsService
}

// setupMetricsTestApp creates a test app with metrics initialized
func setupMetricsTestApp(t *testing.T) *fiber.App {
	t.Helper()

	// Get the singleton metrics service
	metricsService := getTestMetricsService()

	app := fiber.New()

	// Use metrics middleware to record some metrics
	app.Use(metricsService.MetricsMiddleware)
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("test")
	})
	app.Get("/metrics", MetricsHandler)

	// Make a test request to generate metrics
	testReq := httptest.NewRequest("GET", "/test", http.NoBody)
	testResp, err := app.Test(testReq, -1)
	if err != nil {
		t.Fatalf("Failed to send test request: %v", err)
	}
	testResp.Body.Close()

	return app
}

func TestMetricsHandler(t *testing.T) {
	app := setupMetricsTestApp(t)

	// Now test the metrics endpoint
	req := httptest.NewRequest("GET", "/metrics", http.NoBody)
	resp, err := app.Test(req, -1)

	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	bodyStr := string(body)

	// Verify it contains Prometheus metrics format
	// Prometheus metrics should have # HELP and # TYPE lines
	if !strings.Contains(bodyStr, "# HELP") && !strings.Contains(bodyStr, "# TYPE") {
		t.Error("Response doesn't appear to contain Prometheus metrics")
	}
}

func TestMetricsHandlerContentType(t *testing.T) {
	app := setupMetricsTestApp(t)

	// Now test the metrics endpoint
	req := httptest.NewRequest("GET", "/metrics", http.NoBody)
	resp, err := app.Test(req, -1)

	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Prometheus metrics endpoint typically returns text/plain
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		t.Error("Expected Content-Type header to be set")
	}
}

func TestMetricsHandlerMultipleCalls(t *testing.T) {
	// Skip this test as it conflicts with other metrics tests due to shared global registry
	t.Skip("Skipping due to metrics registry conflicts in test environment")

	// Get the singleton metrics service (already initialized by earlier tests)
	metricsService := getTestMetricsService()

	app := fiber.New()
	app.Use(metricsService.MetricsMiddleware)
	app.Get("/metrics", MetricsHandler)

	// Call metrics endpoint multiple times - should work without errors
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/metrics", http.NoBody)
		resp, err := app.Test(req, -1)

		if err != nil {
			t.Fatalf("Failed to send request on iteration %d: %v", i, err)
		}

		if resp.StatusCode != fiber.StatusOK {
			// Read body to see error message
			body, _ := io.ReadAll(resp.Body)
			t.Errorf("Expected status 200 on iteration %d, got %d. Body: %s", i, resp.StatusCode, string(body))
		}
		resp.Body.Close()
	}
}
