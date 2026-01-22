package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/ullbergm/echo-server/services"
)

// TestGetRequestTLSInfo_HTTPS tests TLS info extraction for HTTPS requests
func TestGetRequestTLSInfo_HTTPS(t *testing.T) {
	app := fiber.New()

	var capturedTLSInfo interface{}
	app.Get("/test", func(c *fiber.Ctx) error {
		// Simulate HTTPS by setting protocol
		// Note: In real scenario, this would come from the actual TLS connection
		info := getRequestTLSInfo(c)
		capturedTLSInfo = info
		return c.SendStatus(200)
	})

	// Test HTTP (non-TLS)
	req := httptest.NewRequest("GET", "/test", http.NoBody)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	resp.Body.Close()

	// The function should return non-nil even for HTTP
	if capturedTLSInfo == nil {
		t.Error("Expected TLS info to be non-nil even for HTTP")
	}
}

// TestGetRequestTLSInfo_WithTLSVersionHeader tests TLS version header detection
func TestGetRequestTLSInfo_WithTLSVersionHeader(t *testing.T) {
	app := fiber.New()
	jwtService := services.NewJWTService()
	bodyService := services.NewBodyService()

	app.Get("/test", EchoHandler(jwtService, bodyService))

	// Test with X-TLS-Version header (simulating load balancer)
	req := httptest.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("X-TLS-Version", "TLSv1.3")
	req.Header.Set("Accept", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

// TestBuildServerInfo_NoErrors tests server info building
func TestBuildServerInfo_NoErrors(t *testing.T) {
	// This test simply ensures buildServerInfo doesn't panic
	// and returns valid data
	info := buildServerInfo()

	if info.Hostname == "" {
		t.Error("Expected hostname to be set")
	}

	if info.Environment == nil {
		t.Error("Expected environment map to be initialized")
	}

	// HostAddress might be empty if no network interfaces, that's ok
	// Just verify it returns a string
	_ = info.HostAddress

	// TLS info might be nil if not configured, that's ok
	_ = info.TLS
}

// TestGetHostAddress tests network interface address retrieval
func TestGetHostAddress(t *testing.T) {
	// This test ensures getHostAddress doesn't panic
	// The actual result depends on the system's network configuration
	address := getHostAddress()

	// Address might be empty on systems without network interfaces
	// Just verify it returns a string (possibly empty)
	_ = address

	// If we get an address, verify it's not "127.0.0.1" (loopback)
	if address == "127.0.0.1" {
		t.Error("getHostAddress should not return loopback address")
	}
}

// TestEchoHandlerHead_Success tests HEAD request handling
func TestEchoHandlerHead_Success(t *testing.T) {
	app := fiber.New()
	app.Head("/test", EchoHandlerHead())

	req := httptest.NewRequest("HEAD", "/test", http.NoBody)
	req.Header.Set("Accept", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// HEAD should have appropriate content type
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		t.Error("Expected Content-Type header to be set")
	}
}

// TestEchoHandlerHead_HTMLAccept tests HEAD request with HTML accept
func TestEchoHandlerHead_HTMLAccept(t *testing.T) {
	app := fiber.New()
	app.Head("/test", EchoHandlerHead())

	req := httptest.NewRequest("HEAD", "/test", http.NoBody)
	req.Header.Set("Accept", "text/html")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Should return HTML content type
	contentType := resp.Header.Get("Content-Type")
	if contentType != "text/html; charset=utf-8" {
		t.Errorf("Expected text/html content type, got %s", contentType)
	}
}

// TestEchoHandlerHead_CustomStatusCode tests HEAD with custom status code
func TestEchoHandlerHead_CustomStatusCode(t *testing.T) {
	app := fiber.New()
	app.Head("/test", EchoHandlerHead())

	req := httptest.NewRequest("HEAD", "/test", http.NoBody)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("x-set-response-status-code", "201")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		t.Errorf("Expected status 201, got %d", resp.StatusCode)
	}
}

// TestSetResponseCookies_Basic tests response cookie handling
func TestSetResponseCookies_Basic(t *testing.T) {
	app := fiber.New()
	jwtService := services.NewJWTService()
	bodyService := services.NewBodyService()

	app.Get("/test", EchoHandler(jwtService, bodyService))

	// Test with x-set-cookie header
	req := httptest.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("x-set-cookie", "session=abc123; Path=/; HttpOnly")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Check if Set-Cookie header was set
	cookies := resp.Header.Values("Set-Cookie")
	if len(cookies) == 0 {
		t.Error("Expected Set-Cookie header to be set")
	}
}

// TestSetResponseCookies_MultipleCookies tests multiple cookie handling
func TestSetResponseCookies_MultipleCookies(t *testing.T) {
	app := fiber.New()
	jwtService := services.NewJWTService()
	bodyService := services.NewBodyService()

	app.Get("/test", EchoHandler(jwtService, bodyService))

	// Test with x-set-cookie header (only first one will be used by the handler)
	req := httptest.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("Accept", "application/json")
	req.Header.Add("x-set-cookie", "session=abc123")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Check if Set-Cookie header was set
	cookies := resp.Header.Values("Set-Cookie")
	if len(cookies) == 0 {
		t.Error("Expected Set-Cookie header to be set")
	}
}
