package handlers

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestMetricsHandler(t *testing.T) {
	app := fiber.New()

	app.Get("/metrics", MetricsHandler)

	req := httptest.NewRequest("GET", "/metrics", nil)
	resp, err := app.Test(req, -1)

	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}

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
	app := fiber.New()

	app.Get("/metrics", MetricsHandler)

	req := httptest.NewRequest("GET", "/metrics", nil)
	resp, err := app.Test(req, -1)

	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}

	// Prometheus metrics endpoint typically returns text/plain
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		t.Error("Expected Content-Type header to be set")
	}
}

func TestMetricsHandlerMultipleCalls(t *testing.T) {
	app := fiber.New()

	app.Get("/metrics", MetricsHandler)

	// Call metrics endpoint multiple times
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/metrics", nil)
		resp, err := app.Test(req, -1)

		if err != nil {
			t.Fatalf("Failed to send request on iteration %d: %v", i, err)
		}

		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("Expected status 200 on iteration %d, got %d", i, resp.StatusCode)
		}
	}
}
