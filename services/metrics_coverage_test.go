package services

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

// TestMetricsMiddleware_Success tests successful metrics recording
func TestMetricsMiddleware_Success(t *testing.T) {
	app := fiber.New()
	app.Use(getTestMetricsService().MetricsMiddleware)
	app.Get("/test-success", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test-success", http.NoBody)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

// TestMetricsMiddleware_HTTPS tests metrics recording for HTTPS
func TestMetricsMiddleware_HTTPS(t *testing.T) {
	app := fiber.New()
	app.Use(getTestMetricsService().MetricsMiddleware)
	app.Get("/secure-test", func(c *fiber.Ctx) error {
		return c.SendString("secure")
	})

	req := httptest.NewRequest("GET", "/secure-test", http.NoBody)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

// TestMetricsMiddleware_WithError tests metrics when handler returns error
func TestMetricsMiddleware_WithError(t *testing.T) {
	app := fiber.New()
	app.Use(getTestMetricsService().MetricsMiddleware)
	app.Get("/error-test", func(c *fiber.Ctx) error {
		return fiber.NewError(500, "internal error")
	})

	req := httptest.NewRequest("GET", "/error-test", http.NoBody)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Metrics should still be recorded even on error
	// The middleware should propagate the error
	if resp.StatusCode == 200 {
		t.Error("Expected error status code")
	}
}

// TestMetricsMiddleware_DifferentMethods tests metrics for different HTTP methods
func TestMetricsMiddleware_DifferentMethods(t *testing.T) {
	app := fiber.New()
	app.Use(getTestMetricsService().MetricsMiddleware)
	app.Post("/post-test", func(c *fiber.Ctx) error {
		return c.SendString("posted")
	})
	app.Put("/put-test", func(c *fiber.Ctx) error {
		return c.SendString("updated")
	})

	// Test POST
	req := httptest.NewRequest("POST", "/post-test", http.NoBody)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Failed to send POST request: %v", err)
	}
	resp.Body.Close()

	// Test PUT
	req = httptest.NewRequest("PUT", "/put-test", http.NoBody)
	resp, err = app.Test(req, -1)
	if err != nil {
		t.Fatalf("Failed to send PUT request: %v", err)
	}
	resp.Body.Close()
}

// TestMetricsMiddleware_DifferentPaths tests metrics for different paths
func TestMetricsMiddleware_DifferentPaths(t *testing.T) {
	app := fiber.New()
	app.Use(getTestMetricsService().MetricsMiddleware)
	app.Get("/metrics-path1", func(c *fiber.Ctx) error {
		return c.SendString("path1")
	})
	app.Get("/metrics-path2", func(c *fiber.Ctx) error {
		return c.SendString("path2")
	})

	// Test path1
	req := httptest.NewRequest("GET", "/metrics-path1", http.NoBody)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Failed to send request to path1: %v", err)
	}
	resp.Body.Close()

	// Test path2
	req = httptest.NewRequest("GET", "/metrics-path2", http.NoBody)
	resp, err = app.Test(req, -1)
	if err != nil {
		t.Fatalf("Failed to send request to path2: %v", err)
	}
	resp.Body.Close()
}
