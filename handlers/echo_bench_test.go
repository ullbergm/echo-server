package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/ullbergm/echo-server/services"
)

func BenchmarkEchoHandler_JSON(b *testing.B) {
	app := fiber.New()
	jwtService := services.NewJWTService()
	bodyService := services.NewBodyService()
	app.Get("/test", EchoHandler(jwtService, bodyService))

	req := httptest.NewRequest("GET", "/test?param=value", http.NoBody)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Custom-Header", "test-value")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		resp, _ := app.Test(req, -1)
		_ = resp.Body.Close()
	}
}

func BenchmarkEchoHandler_HTML(b *testing.B) {
	app := fiber.New()
	jwtService := services.NewJWTService()
	bodyService := services.NewBodyService()
	app.Get("/test", EchoHandler(jwtService, bodyService))

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("Accept", "text/html")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		resp, _ := app.Test(req, -1)
		_ = resp.Body.Close()
	}
}

func BenchmarkEchoHandler_WithJWT(b *testing.B) {
	app := fiber.New()
	jwtService := services.NewJWTService()
	bodyService := services.NewBodyService()
	app.Get("/test", EchoHandler(jwtService, bodyService))

	// Valid JWT token (header.payload.signature)
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		resp, _ := app.Test(req, -1)
		_ = resp.Body.Close()
	}
}

func BenchmarkEchoHandler_WithBody(b *testing.B) {
	app := fiber.New()
	jwtService := services.NewJWTService()
	bodyService := services.NewBodyService()
	app.Post("/test", EchoHandler(jwtService, bodyService))

	body := `{"name":"John Doe","email":"john@example.com","age":30}`
	req := httptest.NewRequest("POST", "/test", strings.NewReader(body))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		resp, _ := app.Test(req, -1)
		_ = resp.Body.Close()
	}
}

func BenchmarkEchoHandler_CustomStatus(b *testing.B) {
	app := fiber.New()
	jwtService := services.NewJWTService()
	bodyService := services.NewBodyService()
	app.Get("/test", EchoHandler(jwtService, bodyService))

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("x-set-response-status-code", "201")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		resp, _ := app.Test(req, -1)
		_ = resp.Body.Close()
	}
}
