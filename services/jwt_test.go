package services

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"testing"
)

func TestNewJWTService(t *testing.T) {
	tests := []struct {
		name        string
		envValue    string
		expectedLen int
	}{
		{
			name:        "default headers",
			envValue:    "",
			expectedLen: 4,
		},
		{
			name:        "custom single header",
			envValue:    "Custom-Header",
			expectedLen: 1,
		},
		{
			name:        "custom multiple headers",
			envValue:    "Header1,Header2,Header3",
			expectedLen: 3,
		},
		{
			name:        "headers with spaces",
			envValue:    " Header1 , Header2 , Header3 ",
			expectedLen: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				if err := os.Setenv("JWT_HEADER_NAMES", tt.envValue); err != nil {
					t.Fatalf("Failed to set JWT_HEADER_NAMES: %v", err)
				}
				defer func() { _ = os.Unsetenv("JWT_HEADER_NAMES") }()
			}

			service := NewJWTService()
			if len(service.headerNames) != tt.expectedLen {
				t.Errorf("Expected %d headers, got %d", tt.expectedLen, len(service.headerNames))
			}
		})
	}
}

func TestExtractAndDecodeJWTs(t *testing.T) {
	// Create a valid JWT token for testing
	header := map[string]interface{}{"alg": "HS256", "typ": "JWT"}
	payload := map[string]interface{}{"sub": "1234567890", "name": "John Doe", "iat": 1516239022.0}

	headerJSON, _ := json.Marshal(header)
	payloadJSON, _ := json.Marshal(payload)

	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)
	payloadB64 := base64.RawURLEncoding.EncodeToString(payloadJSON)
	signature := "SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"

	validToken := headerB64 + "." + payloadB64 + "." + signature

	tests := []struct {
		name          string
		headers       map[string]string
		expectedCount int
		expectedName  string
	}{
		{
			name:          "empty headers",
			headers:       map[string]string{},
			expectedCount: 0,
		},
		{
			name: "valid Authorization header",
			headers: map[string]string{
				"Authorization": validToken,
			},
			expectedCount: 1,
			expectedName:  "Authorization",
		},
		{
			name: "valid Authorization header with Bearer prefix",
			headers: map[string]string{
				"Authorization": "Bearer " + validToken,
			},
			expectedCount: 1,
			expectedName:  "Authorization",
		},
		{
			name: "valid X-JWT-Token header",
			headers: map[string]string{
				"X-JWT-Token": validToken,
			},
			expectedCount: 1,
			expectedName:  "X-JWT-Token",
		},
		{
			name: "multiple valid headers",
			headers: map[string]string{
				"Authorization": validToken,
				"X-JWT-Token":   validToken,
			},
			expectedCount: 2,
		},
		{
			name: "invalid token",
			headers: map[string]string{
				"Authorization": "invalid.token",
			},
			expectedCount: 0,
		},
		{
			name: "lowercase header name",
			headers: map[string]string{
				"authorization": validToken,
			},
			expectedCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewJWTService()
			result := service.ExtractAndDecodeJWTs(tt.headers)

			if len(result) != tt.expectedCount {
				t.Errorf("Expected %d decoded tokens, got %d", tt.expectedCount, len(result))
			}

			if tt.expectedName != "" {
				if _, ok := result[tt.expectedName]; !ok {
					t.Errorf("Expected token under key %s, but not found", tt.expectedName)
				}
			}

			// Verify decoded content if tokens were found
			if tt.expectedCount > 0 {
				for _, jwtInfo := range result {
					if jwtInfo.Header == nil {
						t.Error("Expected header to be decoded")
					}
					if jwtInfo.Payload == nil {
						t.Error("Expected payload to be decoded")
					}
					if jwtInfo.RawToken == "" {
						t.Error("Expected raw token to be set")
					}
				}
			}
		})
	}
}

func TestDecodeJWT(t *testing.T) {
	service := NewJWTService()

	tests := []struct {
		name      string
		token     string
		expectNil bool
	}{
		{
			name:      "empty token",
			token:     "",
			expectNil: true,
		},
		{
			name:      "invalid format - no dots",
			token:     "invalidtoken",
			expectNil: true,
		},
		{
			name:      "invalid format - only two parts",
			token:     "header.payload",
			expectNil: true,
		},
		{
			name:      "invalid base64 in header",
			token:     "!!!invalid!!!.eyJzdWIiOiIxMjM0NTY3ODkwIn0.signature",
			expectNil: true,
		},
		{
			name:      "invalid base64 in payload",
			token:     "eyJhbGciOiJIUzI1NiJ9.!!!invalid!!!.signature",
			expectNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.decodeJWT(tt.token)
			if tt.expectNil && result != nil {
				t.Error("Expected nil result for invalid token")
			}
		})
	}
}

func TestDecodeBase64URL(t *testing.T) {
	service := NewJWTService()

	tests := []struct {
		name      string
		input     string
		expectErr bool
	}{
		{
			name:      "valid base64url",
			input:     base64.RawURLEncoding.EncodeToString([]byte(`{"test":"value"}`)),
			expectErr: false,
		},
		{
			name:      "invalid base64",
			input:     "!!!invalid!!!",
			expectErr: true,
		},
		{
			name:      "invalid json",
			input:     base64.RawURLEncoding.EncodeToString([]byte(`not json`)),
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.decodeBase64URL(tt.input)
			if tt.expectErr && err == nil {
				t.Error("Expected error but got nil")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestTruncateToken(t *testing.T) {
	service := NewJWTService()

	tests := []struct {
		name     string
		token    string
		expected string
	}{
		{
			name:     "short token",
			token:    "short",
			expected: "short",
		},
		{
			name:     "token exactly 30 chars",
			token:    "123456789012345678901234567890",
			expected: "123456789012345678901234567890",
		},
		{
			name:     "long token",
			token:    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ",
			expected: "eyJhbGciOi...MjM5MDIyfQ", // Last 10 chars of the actual token
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.truncateToken(tt.token)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}
