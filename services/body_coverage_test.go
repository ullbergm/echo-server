package services

import (
	"os"
	"strings"
	"testing"
)

// TestParseBody_TruncatedBody tests body truncation when exceeding max size
func TestParseBody_TruncatedBody(t *testing.T) {
	service := NewBodyService()

	// Create a body larger than the default max size
	largeBody := []byte(strings.Repeat("x", service.maxBodySize+1000))

	bodyInfo := service.ParseBody(largeBody, "text/plain")

	if bodyInfo == nil {
		t.Fatal("Expected body info to be non-nil")
	}

	if !bodyInfo.Truncated {
		t.Error("Expected body to be marked as truncated")
	}

	if bodyInfo.Size != len(largeBody) {
		t.Errorf("Expected size to be %d, got %d", len(largeBody), bodyInfo.Size)
	}
}

// TestParseBody_InvalidContentType tests parsing with invalid content type
func TestParseBody_InvalidContentType(t *testing.T) {
	service := NewBodyService()

	body := []byte("test content")
	bodyInfo := service.ParseBody(body, "invalid;;;content-type")

	if bodyInfo == nil {
		t.Fatal("Expected body info to be non-nil")
	}

	// Should still parse the body
	if bodyInfo.Content == nil {
		t.Error("Expected content to be set")
	}
}

// TestParseBody_UnknownContentType tests parsing with unknown content type
func TestParseBody_UnknownContentType(t *testing.T) {
	service := NewBodyService()

	// Valid UTF-8 text with unknown content type
	body := []byte("some text content")
	bodyInfo := service.ParseBody(body, "application/x-custom-type")

	if bodyInfo == nil {
		t.Fatal("Expected body info to be non-nil")
	}

	// Should parse as text since it's valid UTF-8
	if bodyInfo.IsBinary {
		t.Error("Expected non-binary content for valid UTF-8")
	}
}

// TestParseBody_InvalidUTF8WithUnknownType tests parsing invalid UTF-8 with unknown content type
func TestParseBody_InvalidUTF8WithUnknownType(t *testing.T) {
	service := NewBodyService()

	// Invalid UTF-8 with unknown content type
	body := []byte{0xFF, 0xFE, 0xFD}
	bodyInfo := service.ParseBody(body, "application/unknown")

	if bodyInfo == nil {
		t.Fatal("Expected body info to be non-nil")
	}

	// Should be marked as binary
	if !bodyInfo.IsBinary {
		t.Error("Expected binary content for invalid UTF-8")
	}
}

// TestNewBodyService_CustomMaxSize tests creating service with custom max body size
func TestNewBodyService_CustomMaxSize(t *testing.T) {
	// Set custom max size via environment variable
	customSize := "5000"
	if err := os.Setenv("MAX_BODY_SIZE", customSize); err != nil {
		t.Fatalf("Failed to set MAX_BODY_SIZE: %v", err)
	}
	defer func() { _ = os.Unsetenv("MAX_BODY_SIZE") }()

	service := NewBodyService()

	if service.maxBodySize != 5000 {
		t.Errorf("Expected max body size 5000, got %d", service.maxBodySize)
	}
}

// TestNewBodyService_InvalidMaxSize tests creating service with invalid max body size
func TestNewBodyService_InvalidMaxSize(t *testing.T) {
	// Set invalid max size
	if err := os.Setenv("MAX_BODY_SIZE", "invalid"); err != nil {
		t.Fatalf("Failed to set MAX_BODY_SIZE: %v", err)
	}
	defer func() { _ = os.Unsetenv("MAX_BODY_SIZE") }()

	service := NewBodyService()

	// Should use default size
	if service.maxBodySize != DefaultMaxBodySize {
		t.Errorf("Expected default max body size %d, got %d", DefaultMaxBodySize, service.maxBodySize)
	}
}

// TestNewBodyService_ZeroMaxSize tests creating service with zero max body size
func TestNewBodyService_ZeroMaxSize(t *testing.T) {
	// Set zero max size
	if err := os.Setenv("MAX_BODY_SIZE", "0"); err != nil {
		t.Fatalf("Failed to set MAX_BODY_SIZE: %v", err)
	}
	defer func() { _ = os.Unsetenv("MAX_BODY_SIZE") }()

	service := NewBodyService()

	// Should use default size (zero is invalid)
	if service.maxBodySize != DefaultMaxBodySize {
		t.Errorf("Expected default max body size %d, got %d", DefaultMaxBodySize, service.maxBodySize)
	}
}

// TestNewBodyService_NegativeMaxSize tests creating service with negative max body size
func TestNewBodyService_NegativeMaxSize(t *testing.T) {
	// Set negative max size
	if err := os.Setenv("MAX_BODY_SIZE", "-1000"); err != nil {
		t.Fatalf("Failed to set MAX_BODY_SIZE: %v", err)
	}
	defer func() { _ = os.Unsetenv("MAX_BODY_SIZE") }()

	service := NewBodyService()

	// Should use default size (negative is invalid)
	if service.maxBodySize != DefaultMaxBodySize {
		t.Errorf("Expected default max body size %d, got %d", DefaultMaxBodySize, service.maxBodySize)
	}
}

// TestParseBody_TextContentType tests parsing various text content types
func TestParseBody_TextContentType(t *testing.T) {
	service := NewBodyService()

	testCases := []struct {
		name        string
		contentType string
		body        []byte
	}{
		{"text/plain", "text/plain", []byte("plain text")},
		{"text/html", "text/html", []byte("<html><body>test</body></html>")},
		{"text/css", "text/css", []byte("body { color: red; }")},
		{"text/javascript", "text/javascript", []byte("console.log('test');")},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			bodyInfo := service.ParseBody(tc.body, tc.contentType)

			if bodyInfo == nil {
				t.Fatal("Expected body info to be non-nil")
			}

			if bodyInfo.IsBinary {
				t.Error("Expected non-binary content for text type")
			}

			content, ok := bodyInfo.Content.(string)
			if !ok {
				t.Error("Expected content to be string")
			}

			if content != string(tc.body) {
				t.Errorf("Expected content %s, got %s", string(tc.body), content)
			}
		})
	}
}

// TestParseMultipartForm_MissingBoundary tests multipart parsing without boundary
func TestParseMultipartForm_MissingBoundary(t *testing.T) {
	service := NewBodyService()

	body := []byte("multipart data without proper boundary")
	bodyInfo := service.ParseBody(body, "multipart/form-data")

	if bodyInfo == nil {
		t.Fatal("Expected body info to be non-nil")
	}

	// Should return the body as-is since no boundary parameter
	content, ok := bodyInfo.Content.(string)
	if !ok {
		t.Error("Expected content to be string when boundary is missing")
	}

	if content != string(body) {
		t.Error("Expected content to match original body")
	}
}

// TestParseBody_EmptyContentType tests parsing with empty content type
func TestParseBody_EmptyContentType(t *testing.T) {
	service := NewBodyService()

	body := []byte("test content")
	bodyInfo := service.ParseBody(body, "")

	if bodyInfo == nil {
		t.Fatal("Expected body info to be non-nil")
	}

	// Should parse as text since it's valid UTF-8
	if bodyInfo.IsBinary {
		t.Error("Expected non-binary content for valid UTF-8 with empty content type")
	}
}
