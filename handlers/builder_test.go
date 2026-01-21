package handlers

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v3"
)

func TestBuilderHandler(t *testing.T) {
	// Create a minimal template engine for testing
	engine := html.New("../templates", ".html")

	app := fiber.New(fiber.Config{
		Views: engine,
	})

	app.Get("/builder", BuilderHandler())

	req := httptest.NewRequest("GET", "/builder", nil)
	resp, err := app.Test(req, -1)

	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Verify content type is HTML
	contentType := resp.Header.Get("Content-Type")
	if contentType != "text/html; charset=utf-8" {
		t.Errorf("Expected content type text/html; charset=utf-8, got %s", contentType)
	}
}

func TestBuilderHandler_RendersTemplate(t *testing.T) {
	// Create template engine
	engine := html.New("../templates", ".html")

	app := fiber.New(fiber.Config{
		Views: engine,
	})

	app.Get("/builder", BuilderHandler())

	req := httptest.NewRequest("GET", "/builder", nil)
	resp, err := app.Test(req, -1)

	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}
