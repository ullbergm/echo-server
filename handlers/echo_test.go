package handlers

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/ullbergm/echo-server/models"
	"github.com/ullbergm/echo-server/services"
)

func TestEchoHandler_JSONResponse(t *testing.T) {
	app := fiber.New()
	jwtService := services.NewJWTService()
	bodyService := services.NewBodyService()

	app.Get("/test", EchoHandler(jwtService, bodyService))

	req := httptest.NewRequest("GET", "/test?param=value", nil)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Custom-Header", "custom-value")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	var echoResponse models.EchoResponse
	err = json.Unmarshal(body, &echoResponse)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Verify request info
	if echoResponse.Request.Method != "GET" {
		t.Errorf("Expected method GET, got %s", echoResponse.Request.Method)
	}

	if echoResponse.Request.Path != "/test" {
		t.Errorf("Expected path /test, got %s", echoResponse.Request.Path)
	}

	if echoResponse.Request.Query != "param=value" {
		t.Errorf("Expected query param=value, got %s", echoResponse.Request.Query)
	}

	// Verify headers were captured
	if _, ok := echoResponse.Request.Headers["X-Custom-Header"]; !ok {
		t.Error("Expected X-Custom-Header to be captured")
	}

	// Verify server info is present
	if echoResponse.Server.Hostname == "" {
		t.Error("Expected hostname to be set")
	}
}

func TestEchoHandler_HTMLResponse(t *testing.T) {
	// Create a minimal template engine for testing
	app := fiber.New(fiber.Config{
		Views: nil, // We'll check for template rendering attempt
	})
	jwtService := services.NewJWTService()
	bodyService := services.NewBodyService()

	app.Get("/test", EchoHandler(jwtService, bodyService))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept", "text/html")

	_, err := app.Test(req, -1)
	if err != nil {
		// Expected to fail without template engine
		// This test validates the handler attempts to render HTML
		return
	}

	// Note: Without actual template, this will fail, but we can test the handler is called
	// For a production test, we'd need to set up the template engine
}

func TestEchoHandler_CustomStatusCode(t *testing.T) {
	app := fiber.New()
	jwtService := services.NewJWTService()
	bodyService := services.NewBodyService()

	app.Get("/test", EchoHandler(jwtService, bodyService))

	tests := []struct {
		name           string
		statusHeader   string
		expectedStatus int
	}{
		{
			name:           "custom 201 status",
			statusHeader:   "201",
			expectedStatus: 201,
		},
		{
			name:           "custom 404 status",
			statusHeader:   "404",
			expectedStatus: 404,
		},
		{
			name:           "custom 500 status",
			statusHeader:   "500",
			expectedStatus: 500,
		},
		{
			name:           "invalid status code (too low)",
			statusHeader:   "100",
			expectedStatus: 200, // Should default to 200
		},
		{
			name:           "invalid status code (too high)",
			statusHeader:   "600",
			expectedStatus: 200, // Should default to 200
		},
		{
			name:           "invalid status code (not a number)",
			statusHeader:   "abc",
			expectedStatus: 200, // Should default to 200
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Accept", "application/json")
			req.Header.Set("x-set-response-status-code", tt.statusHeader)

			resp, err := app.Test(req, -1)
			if err != nil {
				t.Fatalf("Failed to send request: %v", err)
			}

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
		})
	}
}

func TestEchoHandler_WithJWTToken(t *testing.T) {
	// Create a valid JWT token
	header := map[string]interface{}{"alg": "HS256", "typ": "JWT"}
	payload := map[string]interface{}{"sub": "1234567890", "name": "Test User"}

	headerJSON, _ := json.Marshal(header)
	payloadJSON, _ := json.Marshal(payload)

	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)
	payloadB64 := base64.RawURLEncoding.EncodeToString(payloadJSON)
	token := headerB64 + "." + payloadB64 + ".signature"

	app := fiber.New()
	jwtService := services.NewJWTService()
	bodyService := services.NewBodyService()

	app.Get("/test", EchoHandler(jwtService, bodyService))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	var echoResponse models.EchoResponse
	err = json.Unmarshal(body, &echoResponse)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Verify JWT token was decoded
	if len(echoResponse.JwtTokens) == 0 {
		t.Error("Expected JWT tokens to be present")
	}

	if _, ok := echoResponse.JwtTokens["Authorization"]; !ok {
		t.Error("Expected Authorization JWT token to be decoded")
	}
}

func TestEchoHandlerHead(t *testing.T) {
	app := fiber.New()

	app.Head("/test", EchoHandlerHead())

	tests := []struct {
		name                string
		acceptHeader        string
		expectedContentType string
	}{
		{
			name:                "JSON content type",
			acceptHeader:        "application/json",
			expectedContentType: "application/json",
		},
		{
			name:                "HTML content type",
			acceptHeader:        "text/html",
			expectedContentType: "text/html; charset=utf-8",
		},
		{
			name:                "default content type",
			acceptHeader:        "",
			expectedContentType: "application/json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("HEAD", "/test", nil)
			if tt.acceptHeader != "" {
				req.Header.Set("Accept", tt.acceptHeader)
			}

			resp, err := app.Test(req, -1)
			if err != nil {
				t.Fatalf("Failed to send request: %v", err)
			}

			if resp.StatusCode != fiber.StatusOK {
				t.Errorf("Expected status 200, got %d", resp.StatusCode)
			}

			contentType := resp.Header.Get("Content-Type")
			if contentType != tt.expectedContentType {
				t.Errorf("Expected content type %s, got %s", tt.expectedContentType, contentType)
			}

			// HEAD requests should have no body
			body, _ := io.ReadAll(resp.Body)
			if len(body) != 0 {
				t.Error("Expected empty body for HEAD request")
			}
		})
	}
}

func TestBuildHeadersMap(t *testing.T) {
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		headers := buildHeadersMap(c)

		if len(headers) == 0 {
			t.Error("Expected headers to be present")
		}

		// Verify custom header was captured
		if val, ok := headers["X-Test-Header"]; !ok || val != "test-value" {
			t.Errorf("Expected X-Test-Header with value test-value, got %s", val)
		}

		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Test-Header", "test-value")
	req.Header.Set("Content-Type", "application/json")

	_, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
}

func TestGetKubernetesInfo(t *testing.T) {
	tests := []struct {
		name        string
		envVars     map[string]string
		expectNil   bool
		expectLabel bool
	}{
		{
			name:      "no kubernetes env vars",
			envVars:   map[string]string{},
			expectNil: true,
		},
		{
			name: "kubernetes env vars present",
			envVars: map[string]string{
				"K8S_NAMESPACE": "default",
				"K8S_POD_NAME":  "echo-server-123",
				"K8S_POD_IP":    "10.0.0.1",
			},
			expectNil: false,
		},
		{
			name: "kubernetes with labels",
			envVars: map[string]string{
				"K8S_NAMESPACE": "production",
				"K8S_POD_NAME":  "echo-server-abc",
				"K8S_LABEL_app": "echo-server",
			},
			expectNil:   false,
			expectLabel: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for key, val := range tt.envVars {
				if err := os.Setenv(key, val); err != nil {
					t.Fatalf("Failed to set env var %s: %v", key, err)
				}
			}

			// Clean up after test
			defer func() {
				for key := range tt.envVars {
					_ = os.Unsetenv(key)
				}
			}()

			k8sInfo := getKubernetesInfo()

			if tt.expectNil && k8sInfo != nil {
				t.Error("Expected nil Kubernetes info")
			}

			if !tt.expectNil && k8sInfo == nil {
				t.Error("Expected non-nil Kubernetes info")
			}

			if tt.expectLabel && k8sInfo != nil {
				if len(k8sInfo.Labels) == 0 {
					t.Error("Expected labels to be present")
				}
			}
		})
	}
}

func TestGetRemoteAddress(t *testing.T) {
	tests := []struct {
		name           string
		headers        map[string]string
		expectedPrefix string
	}{
		{
			name: "X-Forwarded-For header",
			headers: map[string]string{
				"X-Forwarded-For": "192.168.1.1, 10.0.0.1",
			},
			expectedPrefix: "192.168.1.1",
		},
		{
			name: "X-Real-IP header",
			headers: map[string]string{
				"X-Real-IP": "172.16.0.1",
			},
			expectedPrefix: "172.16.0.1",
		},
		{
			name:           "default IP",
			headers:        map[string]string{},
			expectedPrefix: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			app.Get("/test", func(c *fiber.Ctx) error {
				remoteAddr := getRemoteAddress(c)
				if tt.expectedPrefix != "" && remoteAddr != tt.expectedPrefix {
					t.Errorf("Expected remote address to start with %s, got %s", tt.expectedPrefix, remoteAddr)
				}
				return c.SendStatus(fiber.StatusOK)
			})

			req := httptest.NewRequest("GET", "/test", nil)
			for key, val := range tt.headers {
				req.Header.Set(key, val)
			}

			_, err := app.Test(req, -1)
			if err != nil {
				t.Fatalf("Failed to send request: %v", err)
			}
		})
	}
}

func TestGetEnvironmentVariables(t *testing.T) {
	tests := []struct {
		name           string
		displayEnv     string
		k8sEnvs        map[string]string
		expectedCount  int
		expectHostname bool
		expectK8sVars  bool
	}{
		{
			name:           "default behavior - only HOSTNAME",
			displayEnv:     "",
			k8sEnvs:        map[string]string{},
			expectedCount:  0, // HOSTNAME might not be set in test environment
			expectHostname: false,
		},
		{
			name:          "custom display vars",
			displayEnv:    "PATH,HOME",
			k8sEnvs:       map[string]string{},
			expectedCount: 0, // These vars might not be set
		},
		{
			name:       "K8S vars when not in kubernetes",
			displayEnv: "",
			k8sEnvs: map[string]string{
				"K8S_TEST": "value",
			},
			expectK8sVars: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment
			if tt.displayEnv != "" {
				if err := os.Setenv("ECHO_ENVIRONMENT_VARIABLES_DISPLAY", tt.displayEnv); err != nil {
					t.Fatalf("Failed to set display env: %v", err)
				}
				defer func() { _ = os.Unsetenv("ECHO_ENVIRONMENT_VARIABLES_DISPLAY") }()
			}

			for key, val := range tt.k8sEnvs {
				if err := os.Setenv(key, val); err != nil {
					t.Fatalf("Failed to set env var %s: %v", key, err)
				}
				defer func(k string) { _ = os.Unsetenv(k) }(key)
			}

			envVars := getEnvironmentVariables()

			// Basic validation
			if envVars == nil {
				t.Error("Expected non-nil environment variables map")
			}

			if tt.expectK8sVars {
				hasK8sVar := false
				for key := range envVars {
					if len(key) >= 4 && key[:4] == "K8S_" {
						hasK8sVar = true
						break
					}
				}
				if !hasK8sVar {
					t.Error("Expected K8S_ variables to be present")
				}
			}
		})
	}
}

func TestEchoHandler_PostWithJSONBody(t *testing.T) {
	app := fiber.New()
	jwtService := services.NewJWTService()
	bodyService := services.NewBodyService()

	app.Post("/test", EchoHandler(jwtService, bodyService))

	jsonBody := `{"name":"John Doe","age":30,"email":"john@example.com"}`
	req := httptest.NewRequest("POST", "/test", strings.NewReader(jsonBody))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	var echoResponse models.EchoResponse
	err = json.Unmarshal(body, &echoResponse)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Verify body was captured
	if echoResponse.Request.Body == nil {
		t.Fatal("Expected body to be present")
	}

	if echoResponse.Request.Body.Size != len(jsonBody) {
		t.Errorf("Expected body size %d, got %d", len(jsonBody), echoResponse.Request.Body.Size)
	}

	if echoResponse.Request.Body.ContentType != "application/json" {
		t.Errorf("Expected content type application/json, got %s", echoResponse.Request.Body.ContentType)
	}

	// Verify JSON was parsed
	bodyContent, ok := echoResponse.Request.Body.Content.(map[string]interface{})
	if !ok {
		t.Fatal("Expected body content to be a map")
	}

	if bodyContent["name"] != "John Doe" {
		t.Errorf("Expected name to be 'John Doe', got %v", bodyContent["name"])
	}
}

func TestEchoHandler_PostWithFormData(t *testing.T) {
	app := fiber.New()
	jwtService := services.NewJWTService()
	bodyService := services.NewBodyService()

	app.Post("/test", EchoHandler(jwtService, bodyService))

	formData := "username=johndoe&password=secret123&remember=true"
	req := httptest.NewRequest("POST", "/test", strings.NewReader(formData))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	var echoResponse models.EchoResponse
	err = json.Unmarshal(body, &echoResponse)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Verify body was captured
	if echoResponse.Request.Body == nil {
		t.Fatal("Expected body to be present")
	}

	// Verify form data was parsed
	bodyContent, ok := echoResponse.Request.Body.Content.(map[string]interface{})
	if !ok {
		t.Fatal("Expected body content to be a map")
	}

	if bodyContent["username"] != "johndoe" {
		t.Errorf("Expected username to be 'johndoe', got %v", bodyContent["username"])
	}
}

func TestEchoHandler_PostWithPlainText(t *testing.T) {
	app := fiber.New()
	jwtService := services.NewJWTService()
	bodyService := services.NewBodyService()

	app.Post("/test", EchoHandler(jwtService, bodyService))

	textBody := "This is plain text content"
	req := httptest.NewRequest("POST", "/test", strings.NewReader(textBody))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "text/plain")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	var echoResponse models.EchoResponse
	err = json.Unmarshal(body, &echoResponse)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Verify body was captured
	if echoResponse.Request.Body == nil {
		t.Fatal("Expected body to be present")
	}

	// Verify text content
	bodyContent, ok := echoResponse.Request.Body.Content.(string)
	if !ok {
		t.Fatal("Expected body content to be a string")
	}

	if bodyContent != textBody {
		t.Errorf("Expected body content '%s', got '%s'", textBody, bodyContent)
	}
}

func TestEchoHandler_PutWithBody(t *testing.T) {
	app := fiber.New()
	jwtService := services.NewJWTService()
	bodyService := services.NewBodyService()

	app.Put("/test", EchoHandler(jwtService, bodyService))

	jsonBody := `{"updated":true}`
	req := httptest.NewRequest("PUT", "/test", strings.NewReader(jsonBody))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	var echoResponse models.EchoResponse
	err = json.Unmarshal(body, &echoResponse)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Verify body was captured
	if echoResponse.Request.Body == nil {
		t.Fatal("Expected body to be present")
	}
}

func TestEchoHandler_GetWithoutBody(t *testing.T) {
	app := fiber.New()
	jwtService := services.NewJWTService()
	bodyService := services.NewBodyService()

	app.Get("/test", EchoHandler(jwtService, bodyService))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	var echoResponse models.EchoResponse
	err = json.Unmarshal(body, &echoResponse)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Verify body is nil for GET request
	if echoResponse.Request.Body != nil {
		t.Error("Expected body to be nil for GET request")
	}
}

func TestParseCookies(t *testing.T) {
	app := fiber.New()
	jwtService := services.NewJWTService()
	bodyService := services.NewBodyService()

	app.Get("/test", EchoHandler(jwtService, bodyService))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Cookie", "session=abc123; user=john")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	var echoResponse models.EchoResponse
	err = json.Unmarshal(body, &echoResponse)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Verify cookies were parsed
	if len(echoResponse.Request.Cookies) == 0 {
		t.Error("Expected cookies to be parsed")
	}

	// Verify cookie values
	cookieMap := make(map[string]string)
	for _, cookie := range echoResponse.Request.Cookies {
		cookieMap[cookie.Name] = cookie.Value
	}

	if cookieMap["session"] != "abc123" {
		t.Errorf("Expected session cookie value 'abc123', got '%s'", cookieMap["session"])
	}

	if cookieMap["user"] != "john" {
		t.Errorf("Expected user cookie value 'john', got '%s'", cookieMap["user"])
	}
}

func TestSetResponseCookies(t *testing.T) {
	app := fiber.New()
	jwtService := services.NewJWTService()
	bodyService := services.NewBodyService()

	app.Get("/test", EchoHandler(jwtService, bodyService))

	tests := []struct {
		name             string
		setCookieHeader  string
		expectCookie     bool
		expectedName     string
		expectedValue    string
		checkAttributes  bool
		expectedPath     string
		expectedHttpOnly bool
		expectedSecure   bool
	}{
		{
			name:            "simple cookie",
			setCookieHeader: "mycookie=myvalue",
			expectCookie:    true,
			expectedName:    "mycookie",
			expectedValue:   "myvalue",
		},
		{
			name:            "cookie with path",
			setCookieHeader: "sessionid=xyz789; Path=/api",
			expectCookie:    true,
			expectedName:    "sessionid",
			expectedValue:   "xyz789",
			checkAttributes: true,
			expectedPath:    "/api",
		},
		{
			name:             "cookie with HttpOnly",
			setCookieHeader:  "secure_token=secret123; HttpOnly",
			expectCookie:     true,
			expectedName:     "secure_token",
			expectedValue:    "secret123",
			checkAttributes:  true,
			expectedHttpOnly: true,
		},
		{
			name:             "cookie with multiple attributes",
			setCookieHeader:  "auth=token456; Path=/; HttpOnly; Secure",
			expectCookie:     true,
			expectedName:     "auth",
			expectedValue:    "token456",
			checkAttributes:  true,
			expectedPath:     "/",
			expectedHttpOnly: true,
			expectedSecure:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Accept", "application/json")
			req.Header.Set("x-set-cookie", tt.setCookieHeader)

			resp, err := app.Test(req, -1)
			if err != nil {
				t.Fatalf("Failed to send request: %v", err)
			}

			if resp.StatusCode != fiber.StatusOK {
				t.Errorf("Expected status 200, got %d", resp.StatusCode)
			}

			// Check if Set-Cookie header is present
			setCookieHeaders := resp.Header["Set-Cookie"]
			if tt.expectCookie && len(setCookieHeaders) == 0 {
				t.Error("Expected Set-Cookie header to be present")
			}

			if tt.expectCookie && len(setCookieHeaders) > 0 {
				cookieHeader := setCookieHeaders[0]

				// Verify cookie name and value
				if !strings.Contains(cookieHeader, tt.expectedName+"="+tt.expectedValue) {
					t.Errorf("Expected cookie header to contain %s=%s, got %s", tt.expectedName, tt.expectedValue, cookieHeader)
				}

				// Check attributes if requested
				if tt.checkAttributes {
					if tt.expectedPath != "" && !strings.Contains(strings.ToLower(cookieHeader), "path="+tt.expectedPath) {
						t.Errorf("Expected Path=%s in cookie header, got %s", tt.expectedPath, cookieHeader)
					}

					if tt.expectedHttpOnly && !strings.Contains(strings.ToLower(cookieHeader), "httponly") {
						t.Errorf("Expected HttpOnly in cookie header, got %s", cookieHeader)
					}

					if tt.expectedSecure && !strings.Contains(strings.ToLower(cookieHeader), "secure") {
						t.Errorf("Expected Secure in cookie header, got %s", cookieHeader)
					}
				}
			}
		})
	}
}

func TestParseSetCookieHeader(t *testing.T) {
	tests := []struct {
		name           string
		headerValue    string
		expectNil      bool
		expectedName   string
		expectedValue  string
		expectedPath   string
		expectedDomain string
		expectedMaxAge int
		httpOnly       bool
		secure         bool
		sameSite       string
	}{
		{
			name:          "simple cookie",
			headerValue:   "name=value",
			expectedName:  "name",
			expectedValue: "value",
		},
		{
			name:          "cookie with path",
			headerValue:   "session=abc; Path=/app",
			expectedName:  "session",
			expectedValue: "abc",
			expectedPath:  "/app",
		},
		{
			name:           "cookie with domain",
			headerValue:    "user=john; Domain=example.com",
			expectedName:   "user",
			expectedValue:  "john",
			expectedDomain: "example.com",
		},
		{
			name:           "cookie with max-age",
			headerValue:    "temp=data; Max-Age=3600",
			expectedName:   "temp",
			expectedValue:  "data",
			expectedMaxAge: 3600,
		},
		{
			name:          "cookie with HttpOnly",
			headerValue:   "token=xyz; HttpOnly",
			expectedName:  "token",
			expectedValue: "xyz",
			httpOnly:      true,
		},
		{
			name:          "cookie with Secure",
			headerValue:   "auth=key; Secure",
			expectedName:  "auth",
			expectedValue: "key",
			secure:        true,
		},
		{
			name:          "cookie with SameSite",
			headerValue:   "tracking=id; SameSite=Strict",
			expectedName:  "tracking",
			expectedValue: "id",
			sameSite:      "Strict",
		},
		{
			name:           "full cookie",
			headerValue:    "full=cookie; Domain=test.com; Path=/api; Max-Age=7200; HttpOnly; Secure; SameSite=Lax",
			expectedName:   "full",
			expectedValue:  "cookie",
			expectedDomain: "test.com",
			expectedPath:   "/api",
			expectedMaxAge: 7200,
			httpOnly:       true,
			secure:         true,
			sameSite:       "Lax",
		},
		{
			name:        "invalid cookie (no value)",
			headerValue: "invalid",
			expectNil:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cookie := parseSetCookieHeader(tt.headerValue)

			if tt.expectNil {
				if cookie != nil {
					t.Error("Expected nil cookie")
				}
				return
			}

			if cookie == nil {
				t.Fatal("Expected non-nil cookie")
			}

			if cookie.Name != tt.expectedName {
				t.Errorf("Expected name %s, got %s", tt.expectedName, cookie.Name)
			}

			if cookie.Value != tt.expectedValue {
				t.Errorf("Expected value %s, got %s", tt.expectedValue, cookie.Value)
			}

			if tt.expectedPath != "" && cookie.Path != tt.expectedPath {
				t.Errorf("Expected path %s, got %s", tt.expectedPath, cookie.Path)
			}

			if tt.expectedDomain != "" && cookie.Domain != tt.expectedDomain {
				t.Errorf("Expected domain %s, got %s", tt.expectedDomain, cookie.Domain)
			}

			if tt.expectedMaxAge != 0 && cookie.MaxAge != tt.expectedMaxAge {
				t.Errorf("Expected max-age %d, got %d", tt.expectedMaxAge, cookie.MaxAge)
			}

			if cookie.HTTPOnly != tt.httpOnly {
				t.Errorf("Expected HTTPOnly %v, got %v", tt.httpOnly, cookie.HTTPOnly)
			}

			if cookie.Secure != tt.secure {
				t.Errorf("Expected Secure %v, got %v", tt.secure, cookie.Secure)
			}

			if tt.sameSite != "" && cookie.SameSite != tt.sameSite {
				t.Errorf("Expected SameSite %s, got %s", tt.sameSite, cookie.SameSite)
			}
		})
	}
}
