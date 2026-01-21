package handlers

import (
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

// TestAllHTTPMethods tests all supported HTTP methods
func TestAllHTTPMethods(t *testing.T) {
	methods := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			app := fiber.New()
			jwtService := services.NewJWTService()
			bodyService := services.NewBodyService()

			// Register handler for all methods
			app.Add(method, "/test", EchoHandler(jwtService, bodyService))

			var body io.Reader
			if method == "POST" || method == "PUT" || method == "PATCH" {
				body = strings.NewReader(`{"test":true}`)
			}

			req := httptest.NewRequest(method, "/test", body)
			req.Header.Set("Accept", "application/json")
			if body != nil {
				req.Header.Set("Content-Type", "application/json")
			}

			resp, err := app.Test(req, -1)
			if err != nil {
				t.Fatalf("Failed to send %s request: %v", method, err)
			}

			if resp.StatusCode != fiber.StatusOK {
				t.Errorf("Expected status 200 for %s, got %d", method, resp.StatusCode)
			}

			respBody, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("Failed to read response body: %v", err)
			}

			var echoResponse models.EchoResponse
			err = json.Unmarshal(respBody, &echoResponse)
			if err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			if echoResponse.Request.Method != method {
				t.Errorf("Expected method %s, got %s", method, echoResponse.Request.Method)
			}
		})
	}
}

// TestKubernetesInfoWithAnnotations tests K8s info with labels and annotations
func TestKubernetesInfoWithAnnotations(t *testing.T) {
	// Set up complete Kubernetes environment
	envVars := map[string]string{
		"K8S_NAMESPACE":                       "production",
		"K8S_POD_NAME":                        "echo-server-deployment-abc123",
		"K8S_POD_IP":                          "10.244.0.5",
		"K8S_NODE_NAME":                       "node-1",
		"KUBERNETES_SERVICE_HOST":             "10.96.0.1",
		"KUBERNETES_SERVICE_PORT":             "443",
		"K8S_LABEL_app":                       "echo-server",
		"K8S_LABEL_version":                   "v1.0.0",
		"K8S_LABEL_environment":               "production",
		"K8S_ANNOTATION_description":          "Echo server for testing",
		"K8S_ANNOTATION_prometheus_io_scrape": "true",
		"K8S_ANNOTATION_prometheus_io_port":   "8080",
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	var echoResponse models.EchoResponse
	err = json.Unmarshal(body, &echoResponse)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Verify Kubernetes info is present
	if echoResponse.Kubernetes == nil {
		t.Fatal("Expected Kubernetes info to be present")
	}

	k8s := echoResponse.Kubernetes

	// Verify basic fields
	if k8s.Namespace != "production" {
		t.Errorf("Expected namespace 'production', got %s", k8s.Namespace)
	}

	if k8s.PodName != "echo-server-deployment-abc123" {
		t.Errorf("Expected pod name 'echo-server-deployment-abc123', got %s", k8s.PodName)
	}

	if k8s.PodIP != "10.244.0.5" {
		t.Errorf("Expected pod IP '10.244.0.5', got %s", k8s.PodIP)
	}

	if k8s.NodeName != "node-1" {
		t.Errorf("Expected node name 'node-1', got %s", k8s.NodeName)
	}

	// Verify labels
	if len(k8s.Labels) == 0 {
		t.Fatal("Expected labels to be present")
	}

	if k8s.Labels["app"] != "echo-server" {
		t.Errorf("Expected label app='echo-server', got %s", k8s.Labels["app"])
	}

	if k8s.Labels["version"] != "v1.0.0" {
		t.Errorf("Expected label version='v1.0.0', got %s", k8s.Labels["version"])
	}

	if k8s.Labels["environment"] != "production" {
		t.Errorf("Expected label environment='production', got %s", k8s.Labels["environment"])
	}

	// Verify annotations
	if len(k8s.Annotations) == 0 {
		t.Fatal("Expected annotations to be present")
	}

	if k8s.Annotations["description"] != "Echo server for testing" {
		t.Errorf("Expected annotation description='Echo server for testing', got %s", k8s.Annotations["description"])
	}

	if k8s.Annotations["prometheus_io_scrape"] != "true" {
		t.Errorf("Expected annotation prometheus.io/scrape='true', got %s", k8s.Annotations["prometheus_io_scrape"])
	}
}

// TestEnvironmentVariablesDisplay tests custom environment variable display
func TestEnvironmentVariablesDisplay(t *testing.T) {
	tests := []struct {
		name        string
		displayVars string
		setVars     map[string]string
		expectVars  []string
	}{
		{
			name:        "single custom variable",
			displayVars: "MY_VAR",
			setVars:     map[string]string{"MY_VAR": "value1"},
			expectVars:  []string{"MY_VAR"},
		},
		{
			name:        "multiple custom variables",
			displayVars: "VAR1, VAR2, VAR3",
			setVars: map[string]string{
				"VAR1": "value1",
				"VAR2": "value2",
				"VAR3": "value3",
			},
			expectVars: []string{"VAR1", "VAR2", "VAR3"},
		},
		{
			name:        "variables with spaces in list",
			displayVars: " VAR_A , VAR_B , VAR_C ",
			setVars: map[string]string{
				"VAR_A": "a",
				"VAR_B": "b",
				"VAR_C": "c",
			},
			expectVars: []string{"VAR_A", "VAR_B", "VAR_C"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set display environment variable
			if err := os.Setenv("ECHO_ENVIRONMENT_VARIABLES_DISPLAY", tt.displayVars); err != nil {
				t.Fatalf("Failed to set ECHO_ENVIRONMENT_VARIABLES_DISPLAY: %v", err)
			}
			defer func() { _ = os.Unsetenv("ECHO_ENVIRONMENT_VARIABLES_DISPLAY") }()

			// Set test variables
			for key, val := range tt.setVars {
				if err := os.Setenv(key, val); err != nil {
					t.Fatalf("Failed to set env var %s: %v", key, err)
				}
			}
			defer func() {
				for key := range tt.setVars {
					_ = os.Unsetenv(key)
				}
			}()

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

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("Failed to read response body: %v", err)
			}

			var echoResponse models.EchoResponse
			err = json.Unmarshal(body, &echoResponse)
			if err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			// Verify expected variables are present
			for _, varName := range tt.expectVars {
				if _, ok := echoResponse.Server.Environment[varName]; !ok {
					t.Errorf("Expected environment variable %s to be present", varName)
				} else if echoResponse.Server.Environment[varName] != tt.setVars[varName] {
					t.Errorf("Expected %s=%s, got %s", varName, tt.setVars[varName], echoResponse.Server.Environment[varName])
				}
			}
		})
	}
}

// TestProxyHeaders tests X-Forwarded-For and X-Real-IP handling
func TestProxyHeaders(t *testing.T) {
	tests := []struct {
		name            string
		xForwardedFor   string
		xRealIP         string
		expectedAddress string
	}{
		{
			name:            "single IP in X-Forwarded-For",
			xForwardedFor:   "203.0.113.1",
			expectedAddress: "203.0.113.1",
		},
		{
			name:            "multiple IPs in X-Forwarded-For",
			xForwardedFor:   "203.0.113.1, 198.51.100.1, 192.0.2.1",
			expectedAddress: "203.0.113.1",
		},
		{
			name:            "X-Forwarded-For with spaces",
			xForwardedFor:   " 203.0.113.5 , 198.51.100.5 ",
			expectedAddress: "203.0.113.5",
		},
		{
			name:            "X-Real-IP only",
			xRealIP:         "198.51.100.10",
			expectedAddress: "198.51.100.10",
		},
		{
			name:            "both headers (X-Forwarded-For takes precedence)",
			xForwardedFor:   "203.0.113.20",
			xRealIP:         "198.51.100.20",
			expectedAddress: "203.0.113.20",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			jwtService := services.NewJWTService()
			bodyService := services.NewBodyService()

			app.Get("/test", EchoHandler(jwtService, bodyService))

			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Accept", "application/json")
			if tt.xForwardedFor != "" {
				req.Header.Set("X-Forwarded-For", tt.xForwardedFor)
			}
			if tt.xRealIP != "" {
				req.Header.Set("X-Real-IP", tt.xRealIP)
			}

			resp, err := app.Test(req, -1)
			if err != nil {
				t.Fatalf("Failed to send request: %v", err)
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

			if echoResponse.Request.RemoteAddress != tt.expectedAddress {
				t.Errorf("Expected remote address %s, got %s", tt.expectedAddress, echoResponse.Request.RemoteAddress)
			}
		})
	}
}

// TestCustomStatusCodesEdgeCases tests edge cases for custom status codes
func TestCustomStatusCodesEdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		statusHeader   string
		expectedStatus int
	}{
		{
			name:           "minimum valid status (200)",
			statusHeader:   "200",
			expectedStatus: 200,
		},
		{
			name:           "maximum valid status (599)",
			statusHeader:   "599",
			expectedStatus: 599,
		},
		{
			name:           "just below minimum (199)",
			statusHeader:   "199",
			expectedStatus: 200, // defaults to OK
		},
		{
			name:           "just above maximum (600)",
			statusHeader:   "600",
			expectedStatus: 200, // defaults to OK
		},
		{
			name:           "negative number",
			statusHeader:   "-1",
			expectedStatus: 200, // defaults to OK
		},
		{
			name:           "empty string",
			statusHeader:   "",
			expectedStatus: 200, // defaults to OK
		},
		{
			name:           "non-numeric string",
			statusHeader:   "invalid",
			expectedStatus: 200, // defaults to OK
		},
		{
			name:           "float number",
			statusHeader:   "201.5",
			expectedStatus: 200, // defaults to OK
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			jwtService := services.NewJWTService()
			bodyService := services.NewBodyService()

			app.Get("/test", EchoHandler(jwtService, bodyService))

			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Accept", "application/json")
			if tt.statusHeader != "" {
				req.Header.Set("x-set-response-status-code", tt.statusHeader)
			}

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

// TestHeadRequestCustomStatus tests HEAD requests with custom status codes
func TestHeadRequestCustomStatus(t *testing.T) {
	tests := []struct {
		name           string
		statusHeader   string
		expectedStatus int
	}{
		{
			name:           "custom 404 status",
			statusHeader:   "404",
			expectedStatus: 404,
		},
		{
			name:           "custom 503 status",
			statusHeader:   "503",
			expectedStatus: 503,
		},
		{
			name:           "default status",
			statusHeader:   "",
			expectedStatus: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()

			app.Head("/test", EchoHandlerHead())

			req := httptest.NewRequest("HEAD", "/test", nil)
			if tt.statusHeader != "" {
				req.Header.Set("x-set-response-status-code", tt.statusHeader)
			}

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

// TestMultipleCookies tests parsing of multiple cookies
func TestMultipleCookies(t *testing.T) {
	app := fiber.New()
	jwtService := services.NewJWTService()
	bodyService := services.NewBodyService()

	app.Get("/test", EchoHandler(jwtService, bodyService))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Cookie", "session_id=abc123; user_token=xyz789; preferences=theme:dark; lang=en")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
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

	// Verify multiple cookies were parsed
	if len(echoResponse.Request.Cookies) < 2 {
		t.Errorf("Expected at least 2 cookies, got %d", len(echoResponse.Request.Cookies))
	}

	// Create map for easy lookup
	cookieMap := make(map[string]string)
	for _, cookie := range echoResponse.Request.Cookies {
		cookieMap[cookie.Name] = cookie.Value
	}

	// Verify specific cookies
	expectedCookies := map[string]string{
		"session_id": "abc123",
		"user_token": "xyz789",
	}

	for name, expectedValue := range expectedCookies {
		if value, ok := cookieMap[name]; !ok {
			t.Errorf("Expected cookie %s to be present", name)
		} else if value != expectedValue {
			t.Errorf("Expected cookie %s=%s, got %s", name, expectedValue, value)
		}
	}
}
