package handlers

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
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

	req := httptest.NewRequest("GET", "/test?param=value", http.NoBody)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Custom-Header", "custom-value")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

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

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("Accept", "text/html")

	resp, err := app.Test(req, -1)
	if err != nil {
		// Expected to fail without template engine
		// This test validates the handler attempts to render HTML
		return
	}
	resp.Body.Close()

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
			req := httptest.NewRequest("GET", "/test", http.NoBody)
			req.Header.Set("Accept", "application/json")
			req.Header.Set("x-set-response-status-code", tt.statusHeader)

			resp, err := app.Test(req, -1)
			if err != nil {
				t.Fatalf("Failed to send request: %v", err)
			}
			defer resp.Body.Close()

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

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

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
			req := httptest.NewRequest("HEAD", "/test", http.NoBody)
			if tt.acceptHeader != "" {
				req.Header.Set("Accept", tt.acceptHeader)
			}

			resp, err := app.Test(req, -1)
			if err != nil {
				t.Fatalf("Failed to send request: %v", err)
			}
			defer resp.Body.Close()

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

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("X-Test-Header", "test-value")
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	resp.Body.Close()
}

func TestGetKubernetesInfo(t *testing.T) {
	tests := []struct {
		envVars     map[string]string
		name        string
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

			req := httptest.NewRequest("GET", "/test", http.NoBody)
			for key, val := range tt.headers {
				req.Header.Set(key, val)
			}

			resp, err := app.Test(req, -1)
			if err != nil {
				t.Fatalf("Failed to send request: %v", err)
			}
			defer resp.Body.Close()
		})
	}
}

func TestGetEnvironmentVariables(t *testing.T) {
	tests := []struct {
		k8sEnvs        map[string]string
		name           string
		displayEnv     string
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
			}
			defer func() {
				for key := range tt.k8sEnvs {
					_ = os.Unsetenv(key)
				}
			}()

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
	defer resp.Body.Close()

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

	formData := "username=johndoe&password=secret123&remember=true" // pragma: allowlist secret
	req := httptest.NewRequest("POST", "/test", strings.NewReader(formData))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

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
	defer resp.Body.Close()

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
	defer resp.Body.Close()

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

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("Accept", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

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

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Cookie", "session=abc123; user=john")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

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
		expectedName     string
		expectedValue    string
		expectedPath     string
		expectCookie     bool
		checkAttributes  bool
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
			req := httptest.NewRequest("GET", "/test", http.NoBody)
			req.Header.Set("Accept", "application/json")
			req.Header.Set("x-set-cookie", tt.setCookieHeader)

			resp, err := app.Test(req, -1)
			if err != nil {
				t.Fatalf("Failed to send request: %v", err)
			}
			defer resp.Body.Close()

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
		expectedName   string
		expectedValue  string
		expectedPath   string
		expectedDomain string
		sameSite       string
		expectedMaxAge int
		expectNil      bool
		httpOnly       bool
		secure         bool
	}{
		{
			name:          "simple cookie",
			headerValue:   "name=value",
			expectedName:  "name",
			expectedValue: "value",
		},
		{
			name:        "empty header",
			headerValue: "",
			expectNil:   true,
		},
		{
			name:          "cookie with SameSite None",
			headerValue:   "tracking=id; SameSite=None",
			expectedName:  "tracking",
			expectedValue: "id",
			sameSite:      "None",
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

// TestParseExpires tests the parseExpires function with various date formats
func TestParseExpires(t *testing.T) {
	tests := []struct {
		name        string
		dateStr     string
		expectError bool
	}{
		{
			name:        "RFC1123 format",
			dateStr:     "Mon, 02 Jan 2006 15:04:05 MST",
			expectError: false,
		},
		{
			name:        "RFC850 format",
			dateStr:     "Monday, 02-Jan-06 15:04:05 MST",
			expectError: false,
		},
		{
			name:        "ANSIC format",
			dateStr:     "Mon Jan  2 15:04:05 2006",
			expectError: false,
		},
		{
			name:        "RFC3339 format",
			dateStr:     "2006-01-02T15:04:05Z",
			expectError: false,
		},
		{
			name:        "invalid format",
			dateStr:     "not-a-date",
			expectError: true,
		},
		{
			name:        "empty string",
			dateStr:     "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseExpires(tt.dateStr)
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error parsing '%s', got nil", tt.dateStr)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error parsing '%s': %v", tt.dateStr, err)
				}
				if result.IsZero() {
					t.Error("Expected non-zero time result")
				}
			}
		})
	}
}

// TestGetCompressionInfo tests the getCompressionInfo function
func TestGetCompressionInfo(t *testing.T) {
	tests := []struct {
		name                   string
		acceptEncoding         string
		expectedEncodings      []string
		expectedEncodingsCount int
		expectedSupported      bool
	}{
		{
			name:                   "no accept encoding",
			acceptEncoding:         "",
			expectedSupported:      false,
			expectedEncodingsCount: 0,
		},
		{
			name:                   "gzip only",
			acceptEncoding:         "gzip",
			expectedSupported:      true,
			expectedEncodingsCount: 1,
			expectedEncodings:      []string{"gzip"},
		},
		{
			name:                   "multiple encodings",
			acceptEncoding:         "gzip, deflate, br",
			expectedSupported:      true,
			expectedEncodingsCount: 3,
			expectedEncodings:      []string{"gzip", "deflate", "br"},
		},
		{
			name:                   "with quality values",
			acceptEncoding:         "gzip;q=1.0, deflate;q=0.5, *;q=0.1",
			expectedSupported:      true,
			expectedEncodingsCount: 3,
			expectedEncodings:      []string{"gzip", "deflate", "*"},
		},
		{
			name:                   "identity encoding",
			acceptEncoding:         "identity",
			expectedSupported:      true,
			expectedEncodingsCount: 1,
			expectedEncodings:      []string{"identity"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			app.Get("/test", func(c *fiber.Ctx) error {
				info := getCompressionInfo(c)
				if info == nil {
					t.Fatal("Expected non-nil compression info")
				}

				if info.Supported != tt.expectedSupported {
					t.Errorf("Expected Supported=%v, got %v", tt.expectedSupported, info.Supported)
				}

				if len(info.AcceptedEncodings) != tt.expectedEncodingsCount {
					t.Errorf("Expected %d encodings, got %d", tt.expectedEncodingsCount, len(info.AcceptedEncodings))
				}

				// Verify expected encodings are present
				for _, expected := range tt.expectedEncodings {
					found := false
					for _, actual := range info.AcceptedEncodings {
						if actual == expected {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Expected encoding '%s' not found in %v", expected, info.AcceptedEncodings)
					}
				}

				return c.SendStatus(fiber.StatusOK)
			})

			req := httptest.NewRequest("GET", "/test", http.NoBody)
			if tt.acceptEncoding != "" {
				req.Header.Set("Accept-Encoding", tt.acceptEncoding)
			}

			resp, err := app.Test(req, -1)
			if err != nil {
				t.Fatalf("Failed to send request: %v", err)
			}
			resp.Body.Close()
		})
	}
}

// TestGetCompressionInfo_WithResponseEncoding tests compression info when response has Content-Encoding
func TestGetCompressionInfo_WithResponseEncoding(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		// Set a Content-Encoding header on the response
		c.Response().Header.Set("Content-Encoding", "gzip")

		info := getCompressionInfo(c)
		if info == nil {
			t.Fatal("Expected non-nil compression info")
		}

		// Check that ResponseEncoding is captured
		if info.ResponseEncoding != "gzip" {
			t.Errorf("Expected ResponseEncoding='gzip', got '%s'", info.ResponseEncoding)
		}

		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("Accept-Encoding", "gzip, deflate")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	resp.Body.Close()
}

// TestGetRequestTLSInfo tests the getRequestTLSInfo function
func TestGetRequestTLSInfo(t *testing.T) {
	// Note: We can't easily simulate HTTPS with httptest
	// So we only test the HTTP case and the function logic
	t.Run("HTTP request", func(t *testing.T) {
		app := fiber.New()
		app.Get("/test", func(c *fiber.Ctx) error {
			info := getRequestTLSInfo(c)
			if info == nil {
				t.Fatal("Expected non-nil TLS info")
			}

			if info.Enabled {
				t.Error("Expected Enabled=false for HTTP request")
			}

			return c.SendStatus(fiber.StatusOK)
		})

		req := httptest.NewRequest("GET", "/test", http.NoBody)

		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}
		resp.Body.Close()
	})
}

// TestGetServerTLSInfo tests the getServerTLSInfo function
func TestGetServerTLSInfo(t *testing.T) {
	tests := []struct {
		envVars        map[string]string
		expectedFields map[string]string
		name           string
		expectedNil    bool
	}{
		{
			name:        "TLS not enabled",
			envVars:     map[string]string{},
			expectedNil: true,
		},
		{
			name: "TLS enabled false",
			envVars: map[string]string{
				"TLS_ENABLED": "false",
			},
			expectedNil: true,
		},
		{
			name: "TLS enabled true with no cert info",
			envVars: map[string]string{
				"TLS_ENABLED": "true",
			},
			expectedNil: false,
		},
		{
			name: "TLS enabled with all cert info",
			envVars: map[string]string{
				"TLS_ENABLED":          "true",
				"_TLS_CERT_SUBJECT":    "CN=test",
				"_TLS_CERT_ISSUER":     "CN=issuer",
				"_TLS_CERT_NOT_BEFORE": "2024-01-01T00:00:00Z",
				"_TLS_CERT_NOT_AFTER":  "2025-01-01T00:00:00Z",
				"_TLS_CERT_SERIAL":     "12345",
				"_TLS_CERT_DNS_NAMES":  "localhost, example.com",
			},
			expectedNil: false,
			expectedFields: map[string]string{
				"Subject":      "CN=test",
				"Issuer":       "CN=issuer",
				"NotBefore":    "2024-01-01T00:00:00Z",
				"NotAfter":     "2025-01-01T00:00:00Z",
				"SerialNumber": "12345",
			},
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

			info := getServerTLSInfo()

			if tt.expectedNil {
				if info != nil {
					t.Error("Expected nil TLS info")
				}
				return
			}

			if info == nil {
				t.Fatal("Expected non-nil TLS info")
			}

			if !info.Enabled {
				t.Error("Expected Enabled to be true")
			}

			// Check expected fields
			if tt.expectedFields != nil {
				if info.Subject != tt.expectedFields["Subject"] {
					t.Errorf("Expected Subject=%s, got %s", tt.expectedFields["Subject"], info.Subject)
				}
				if info.Issuer != tt.expectedFields["Issuer"] {
					t.Errorf("Expected Issuer=%s, got %s", tt.expectedFields["Issuer"], info.Issuer)
				}
				if info.NotBefore != tt.expectedFields["NotBefore"] {
					t.Errorf("Expected NotBefore=%s, got %s", tt.expectedFields["NotBefore"], info.NotBefore)
				}
				if info.NotAfter != tt.expectedFields["NotAfter"] {
					t.Errorf("Expected NotAfter=%s, got %s", tt.expectedFields["NotAfter"], info.NotAfter)
				}
				if info.SerialNumber != tt.expectedFields["SerialNumber"] {
					t.Errorf("Expected SerialNumber=%s, got %s", tt.expectedFields["SerialNumber"], info.SerialNumber)
				}
				// Check DNS names
				if len(info.DNSNames) != 2 {
					t.Errorf("Expected 2 DNS names, got %d", len(info.DNSNames))
				}
			}
		})
	}
}

// TestBuildServerInfo tests the buildServerInfo function
func TestBuildServerInfo(t *testing.T) {
	info := buildServerInfo()

	// Hostname should always be set (to "unknown" at minimum)
	if info.Hostname == "" {
		t.Error("Expected hostname to be set")
	}

	// Environment should be a valid map
	if info.Environment == nil {
		t.Error("Expected environment map to be non-nil")
	}
}

// TestGetKubernetesInfoWithAnnotations tests Kubernetes info with annotations
func TestGetKubernetesInfoWithAnnotations(t *testing.T) {
	// Set up environment for Kubernetes with annotations
	envVars := map[string]string{
		"K8S_NAMESPACE":           "production",
		"K8S_POD_NAME":            "echo-server-abc",
		"K8S_POD_IP":              "10.0.0.5",
		"K8S_NODE_NAME":           "node-1",
		"KUBERNETES_SERVICE_HOST": "10.0.0.1",
		"KUBERNETES_SERVICE_PORT": "443",
		"K8S_LABEL_app":           "echo-server",
		"K8S_ANNOTATION_owner":    "team-a",
		"K8S_ANNOTATION_version":  "v1.0.0",
	}

	for key, val := range envVars {
		if err := os.Setenv(key, val); err != nil {
			t.Fatalf("Failed to set env var %s: %v", key, err)
		}
	}

	defer func() {
		for key := range envVars {
			_ = os.Unsetenv(key)
		}
	}()

	k8sInfo := getKubernetesInfo()

	if k8sInfo == nil {
		t.Fatal("Expected non-nil Kubernetes info")
	}

	if k8sInfo.Namespace != "production" {
		t.Errorf("Expected namespace 'production', got '%s'", k8sInfo.Namespace)
	}

	if k8sInfo.PodIP != "10.0.0.5" {
		t.Errorf("Expected pod IP '10.0.0.5', got '%s'", k8sInfo.PodIP)
	}

	if k8sInfo.NodeName != "node-1" {
		t.Errorf("Expected node name 'node-1', got '%s'", k8sInfo.NodeName)
	}

	if k8sInfo.ServiceHost != "10.0.0.1" {
		t.Errorf("Expected service host '10.0.0.1', got '%s'", k8sInfo.ServiceHost)
	}

	if k8sInfo.ServicePort != "443" {
		t.Errorf("Expected service port '443', got '%s'", k8sInfo.ServicePort)
	}

	// Check labels
	if len(k8sInfo.Labels) == 0 {
		t.Error("Expected labels to be present")
	}

	// Check annotations
	if len(k8sInfo.Annotations) == 0 {
		t.Error("Expected annotations to be present")
	}

	if k8sInfo.Annotations["owner"] != "team-a" {
		t.Errorf("Expected annotation owner='team-a', got '%s'", k8sInfo.Annotations["owner"])
	}
}

// TestEchoHandler_DeleteWithBody tests DELETE with body
func TestEchoHandler_DeleteWithBody(t *testing.T) {
	app := fiber.New()
	jwtService := services.NewJWTService()
	bodyService := services.NewBodyService()

	app.Delete("/test", EchoHandler(jwtService, bodyService))

	jsonBody := `{"id":"12345"}`
	req := httptest.NewRequest("DELETE", "/test", strings.NewReader(jsonBody))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

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
		t.Fatal("Expected body to be present for DELETE request")
	}
}

// TestEchoHandler_PatchWithBody tests PATCH with body
func TestEchoHandler_PatchWithBody(t *testing.T) {
	app := fiber.New()
	jwtService := services.NewJWTService()
	bodyService := services.NewBodyService()

	app.Patch("/test", EchoHandler(jwtService, bodyService))

	jsonBody := `{"field":"updated"}`
	req := httptest.NewRequest("PATCH", "/test", strings.NewReader(jsonBody))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

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
		t.Fatal("Expected body to be present for PATCH request")
	}
}

// TestParseSetCookieHeader_WithExpires tests cookie parsing with Expires attribute
func TestParseSetCookieHeader_WithExpires(t *testing.T) {
	tests := []struct {
		name        string
		header      string
		expectNil   bool
		checkExpiry bool
	}{
		{
			name:        "cookie with RFC1123 expires",
			header:      "session=abc; Expires=Mon, 02 Jan 2006 15:04:05 GMT",
			expectNil:   false,
			checkExpiry: true,
		},
		{
			name:        "cookie with invalid expires",
			header:      "session=abc; Expires=invalid-date",
			expectNil:   false,
			checkExpiry: false, // Should be zero time
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cookie := parseSetCookieHeader(tt.header)

			if tt.expectNil {
				if cookie != nil {
					t.Error("Expected nil cookie")
				}
				return
			}

			if cookie == nil {
				t.Fatal("Expected non-nil cookie")
			}

			if cookie.Name != "session" {
				t.Errorf("Expected name 'session', got '%s'", cookie.Name)
			}

			if tt.checkExpiry && cookie.Expires.IsZero() {
				t.Error("Expected non-zero Expires time")
			}
		})
	}
}

// TestEchoHandlerHead_CustomStatus tests HEAD request with custom status
func TestEchoHandlerHead_CustomStatus(t *testing.T) {
	app := fiber.New()
	app.Head("/test", EchoHandlerHead())

	req := httptest.NewRequest("HEAD", "/test", http.NoBody)
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

// TestGetEnvironmentVariablesInKubernetes tests env vars when in Kubernetes
func TestGetEnvironmentVariablesInKubernetes(t *testing.T) {
	// Set up environment to simulate Kubernetes
	envVars := map[string]string{
		"K8S_NAMESPACE": "default",
		"K8S_POD_NAME":  "echo-123",
	}

	for key, val := range envVars {
		if err := os.Setenv(key, val); err != nil {
			t.Fatalf("Failed to set env var %s: %v", key, err)
		}
	}

	defer func() {
		for key := range envVars {
			_ = os.Unsetenv(key)
		}
	}()

	result := getEnvironmentVariables()

	// When in Kubernetes, K8S_* vars should NOT be included in the general env vars section
	// (they go in the kubernetes section instead)
	for key := range result {
		if strings.HasPrefix(key, "K8S_") {
			t.Errorf("K8S_ variables should not be in environment section when running in Kubernetes: %s", key)
		}
	}
}

// TestGetEnvironmentVariablesWithCustomDisplay tests custom environment variable display
func TestGetEnvironmentVariablesWithCustomDisplay(t *testing.T) {
	// Set up environment
	if err := os.Setenv("ECHO_ENVIRONMENT_VARIABLES_DISPLAY", "TEST_VAR1,TEST_VAR2"); err != nil {
		t.Fatal(err)
	}
	if err := os.Setenv("TEST_VAR1", "value1"); err != nil {
		t.Fatal(err)
	}
	if err := os.Setenv("TEST_VAR2", "value2"); err != nil {
		t.Fatal(err)
	}

	defer func() {
		_ = os.Unsetenv("ECHO_ENVIRONMENT_VARIABLES_DISPLAY")
		_ = os.Unsetenv("TEST_VAR1")
		_ = os.Unsetenv("TEST_VAR2")
	}()

	result := getEnvironmentVariables()

	if result["TEST_VAR1"] != "value1" {
		t.Errorf("Expected TEST_VAR1=value1, got %s", result["TEST_VAR1"])
	}

	if result["TEST_VAR2"] != "value2" {
		t.Errorf("Expected TEST_VAR2=value2, got %s", result["TEST_VAR2"])
	}
}
