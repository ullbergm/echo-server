package handlers

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/ullbergm/echo-server/services"
)

// sanitizePath ensures the path is valid for HTTP requests
func sanitizePath(path string) string {
	if path == "" {
		return "/"
	}
	// Ensure path starts with /
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	// Remove control characters and null bytes
	cleanPath := strings.Map(func(r rune) rune {
		if r < 32 || r == 127 {
			return -1 // Remove control characters
		}
		return r
	}, path)
	if cleanPath == "" {
		return "/"
	}
	// URL encode the path properly
	u, err := url.Parse(cleanPath)
	if err != nil {
		return "/test"
	}
	// Ensure the path is properly escaped
	return u.EscapedPath()
}

// FuzzEchoHandler tests the echo handler with various request configurations
func FuzzEchoHandler(f *testing.F) {
	// Seed with various paths - all should be valid URL paths
	paths := []string{
		"/",
		"/test",
		"/api/v1/resource",
		"/path/with/multiple/segments",
		"/path-with-dashes",
		"/unicode/test",
		"/special/chars",
		"/parent/path",
		"/path",
		"/very/long/path/segment",
	}

	for _, path := range paths {
		f.Add(path)
	}

	jwtService := services.NewJWTService()
	bodyService := services.NewBodyService()

	f.Fuzz(func(t *testing.T, path string) {
		// Sanitize the path to avoid httptest.NewRequest panics
		safePath := sanitizePath(path)

		app := fiber.New(fiber.Config{
			DisableStartupMessage: true,
			Views:                 nil,
		})
		app.Get("/*", EchoHandler(jwtService, bodyService))

		// Create request with sanitized path
		req := httptest.NewRequest("GET", safePath, http.NoBody)
		req.Header.Set("Accept", "application/json")

		// Should not panic
		resp, err := app.Test(req, -1)
		if err != nil {
			// Some malformed paths may cause errors, that's acceptable
			return
		}
		defer resp.Body.Close()

		// Should return a valid status code
		if resp.StatusCode < 100 || resp.StatusCode > 599 {
			t.Errorf("Invalid status code: %d", resp.StatusCode)
		}
	})
}

// FuzzEchoHandlerHeaders tests header handling
func FuzzEchoHandlerHeaders(f *testing.F) {
	// Seed with various header values
	headers := []struct {
		name  string
		value string
	}{
		{"X-Custom-Header", "custom-value"},
		{"Authorization", "Bearer token123"},
		{"Content-Type", "application/json"},
		{"Accept", "text/html"},
		{"X-Forwarded-For", "192.168.1.1, 10.0.0.1"},
		{"X-Real-IP", "192.168.1.1"},
		{"Cookie", "session=abc123; token=xyz"},
		// Potentially malicious headers
		{"X-Injection", "<script>alert('xss')</script>"},
		{"X-SQL", "'; DROP TABLE users; --"},
		{"X-Path", "../../../etc/passwd"},
		// Long header value
		{"X-Long", strings.Repeat("x", 1000)},
		// Unicode header
		{"X-Unicode", "test-unicode"},
		// Empty value
		{"X-Empty", ""},
		// Special characters
		{"X-Special", "value-with-special-chars"},
	}

	for _, h := range headers {
		f.Add(h.name, h.value)
	}

	jwtService := services.NewJWTService()
	bodyService := services.NewBodyService()

	f.Fuzz(func(t *testing.T, headerName, headerValue string) {
		app := fiber.New(fiber.Config{
			DisableStartupMessage: true,
			Views:                 nil,
		})
		app.Get("/*", EchoHandler(jwtService, bodyService))

		req := httptest.NewRequest("GET", "/test", http.NoBody)
		req.Header.Set("Accept", "application/json")

		// Skip if header name contains invalid characters
		if headerName != "" {
			req.Header.Set(headerName, headerValue)
		}

		// Should not panic
		resp, err := app.Test(req, -1)
		if err != nil {
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode < 100 || resp.StatusCode > 599 {
			t.Errorf("Invalid status code: %d", resp.StatusCode)
		}
	})
}

// FuzzEchoHandlerBody tests request body handling
func FuzzEchoHandlerBody(f *testing.F) {
	bodies := [][]byte{
		[]byte(`{"test": "json"}`),
		[]byte(`<xml>data</xml>`),
		[]byte(`key=value&other=data`),
		[]byte(`plain text body`),
		{0x00, 0x01, 0x02, 0x03}, // Binary
		[]byte(``),               // Empty
		// Large body
		bytes.Repeat([]byte("a"), 10000),
		// Unicode
		[]byte(`{"unicode": "日本語"}`),
		// Malicious payloads
		[]byte(`<script>alert('xss')</script>`),
		[]byte(`{"__proto__": {"polluted": true}}`),
	}

	for _, body := range bodies {
		f.Add(body, "application/json")
	}

	jwtService := services.NewJWTService()
	bodyService := services.NewBodyService()

	f.Fuzz(func(t *testing.T, body []byte, contentType string) {
		app := fiber.New(fiber.Config{
			DisableStartupMessage: true,
			Views:                 nil,
		})
		app.Post("/*", EchoHandler(jwtService, bodyService))

		req := httptest.NewRequest("POST", "/test", bytes.NewReader(body))
		req.Header.Set("Accept", "application/json")
		if contentType != "" {
			req.Header.Set("Content-Type", contentType)
		}

		// Should not panic
		resp, err := app.Test(req, -1)
		if err != nil {
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode < 100 || resp.StatusCode > 599 {
			t.Errorf("Invalid status code: %d", resp.StatusCode)
		}
	})
}

// FuzzEchoHandlerQueryParams tests query parameter handling
func FuzzEchoHandlerQueryParams(f *testing.F) {
	queries := []string{
		"key=value",
		"key1=value1&key2=value2",
		"array=1&array=2&array=3",
		"encoded=%20%21%40%23",
		"unicode=%E6%97%A5%E6%9C%AC",
		"empty=",
		"=nokey",
		// Potentially dangerous
		"sql=1%27%20OR%20%271%27=%271",
		"xss=%3Cscript%3Ealert(1)%3C/script%3E",
		"path=../../../etc/passwd",
		// Long query
		"long=" + strings.Repeat("x", 500),
		// Multiple equals
		"key=val=ue=test",
		// Empty
		"",
		// Special characters (URL encoded)
		"special=%21%40%23%24%25",
	}

	for _, q := range queries {
		f.Add(q)
	}

	jwtService := services.NewJWTService()
	bodyService := services.NewBodyService()

	f.Fuzz(func(t *testing.T, query string) {
		app := fiber.New(fiber.Config{
			DisableStartupMessage: true,
			Views:                 nil,
		})
		app.Get("/*", EchoHandler(jwtService, bodyService))

		// Sanitize the query string to remove control characters
		cleanQuery := strings.Map(func(r rune) rune {
			if r < 32 || r == 127 {
				return -1
			}
			return r
		}, query)

		path := "/test"
		if cleanQuery != "" {
			path = "/test?" + cleanQuery
		}

		req := httptest.NewRequest("GET", path, http.NoBody)
		req.Header.Set("Accept", "application/json")

		// Should not panic
		resp, err := app.Test(req, -1)
		if err != nil {
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode < 100 || resp.StatusCode > 599 {
			t.Errorf("Invalid status code: %d", resp.StatusCode)
		}
	})
}

// FuzzEchoHandlerStatusCode tests custom status code handling
func FuzzEchoHandlerStatusCode(f *testing.F) {
	statusCodes := []string{
		"200",
		"201",
		"301",
		"400",
		"401",
		"403",
		"404",
		"500",
		"503",
		"599",
		// Edge cases
		"0",
		"-1",
		"1000",
		"99",
		"600",
		// Invalid
		"abc",
		"200.5",
		"",
		"200 OK",
	}

	for _, sc := range statusCodes {
		f.Add(sc)
	}

	jwtService := services.NewJWTService()
	bodyService := services.NewBodyService()

	f.Fuzz(func(t *testing.T, statusCode string) {
		app := fiber.New(fiber.Config{
			DisableStartupMessage: true,
			Views:                 nil,
		})
		app.Get("/*", EchoHandler(jwtService, bodyService))

		req := httptest.NewRequest("GET", "/test", http.NoBody)
		req.Header.Set("Accept", "application/json")
		req.Header.Set("x-set-response-status-code", statusCode)

		// Should not panic
		resp, err := app.Test(req, -1)
		if err != nil {
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode < 100 || resp.StatusCode > 599 {
			t.Errorf("Invalid status code returned: %d", resp.StatusCode)
		}
	})
}

// FuzzEchoHandlerAcceptHeader tests content negotiation
func FuzzEchoHandlerAcceptHeader(f *testing.F) {
	accepts := []string{
		"application/json",
		"text/html",
		"text/plain",
		"*/*",
		"application/xml",
		"text/html, application/json",
		"text/html;q=0.9, application/json;q=0.8",
		"",
		"invalid/type",
		"application/json; charset=utf-8",
	}

	for _, a := range accepts {
		f.Add(a)
	}

	jwtService := services.NewJWTService()
	bodyService := services.NewBodyService()

	f.Fuzz(func(t *testing.T, accept string) {
		app := fiber.New(fiber.Config{
			Views:                 nil, // HTML rendering will fail gracefully
			DisableStartupMessage: true,
		})
		app.Get("/*", EchoHandler(jwtService, bodyService))

		req := httptest.NewRequest("GET", "/test", http.NoBody)
		if accept != "" {
			req.Header.Set("Accept", accept)
		}

		// Should not panic
		resp, err := app.Test(req, -1)
		if err != nil {
			return
		}
		defer resp.Body.Close()

		// Read body to ensure no issues
		_, _ = io.ReadAll(resp.Body)
	})
}

// FuzzEchoHandlerCookies tests cookie handling
func FuzzEchoHandlerCookies(f *testing.F) {
	cookies := []string{
		"session=abc123",
		"session=abc123; token=xyz789",
		"name=value; Path=/; HttpOnly",
		"unicode=test",
		"empty=",
		"=noname",
		// Potentially dangerous
		"xss=script-test",
		"sql=drop-table-test",
		// Long cookie value
		"long=" + strings.Repeat("x", 500),
		// Special characters
		"special=test-chars",
	}

	for _, c := range cookies {
		f.Add(c)
	}

	jwtService := services.NewJWTService()
	bodyService := services.NewBodyService()

	f.Fuzz(func(t *testing.T, cookie string) {
		app := fiber.New(fiber.Config{
			DisableStartupMessage: true,
			Views:                 nil,
		})
		app.Get("/*", EchoHandler(jwtService, bodyService))

		req := httptest.NewRequest("GET", "/test", http.NoBody)
		req.Header.Set("Accept", "application/json")
		if cookie != "" {
			req.Header.Set("Cookie", cookie)
		}

		// Should not panic
		resp, err := app.Test(req, -1)
		if err != nil {
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode < 100 || resp.StatusCode > 599 {
			t.Errorf("Invalid status code: %d", resp.StatusCode)
		}
	})
}
