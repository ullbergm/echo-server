package services

import (
	"encoding/base64"
	"testing"
)

func TestBodyService_ParseJSON(t *testing.T) {
	service := NewBodyService()

	jsonBody := []byte(`{"name":"John","age":30,"active":true}`)
	bodyInfo := service.ParseBody(jsonBody, "application/json")

	if bodyInfo == nil {
		t.Fatal("Expected body info to be non-nil")
	}

	if bodyInfo.Size != len(jsonBody) {
		t.Errorf("Expected size %d, got %d", len(jsonBody), bodyInfo.Size)
	}

	if bodyInfo.IsBinary {
		t.Error("Expected IsBinary to be false for JSON")
	}

	contentMap, ok := bodyInfo.Content.(map[string]interface{})
	if !ok {
		t.Fatal("Expected content to be a map")
	}

	if contentMap["name"] != "John" {
		t.Errorf("Expected name to be 'John', got %v", contentMap["name"])
	}

	if contentMap["age"] != float64(30) {
		t.Errorf("Expected age to be 30, got %v", contentMap["age"])
	}
}

func TestBodyService_ParseFormURLEncoded(t *testing.T) {
	service := NewBodyService()

	formBody := []byte("username=johndoe&email=john%40example.com&role=admin")
	bodyInfo := service.ParseBody(formBody, "application/x-www-form-urlencoded")

	if bodyInfo == nil {
		t.Fatal("Expected body info to be non-nil")
	}

	contentMap, ok := bodyInfo.Content.(map[string]interface{})
	if !ok {
		t.Fatal("Expected content to be a map")
	}

	if contentMap["username"] != "johndoe" {
		t.Errorf("Expected username to be 'johndoe', got %v", contentMap["username"])
	}

	if contentMap["email"] != "john@example.com" {
		t.Errorf("Expected email to be 'john@example.com', got %v", contentMap["email"])
	}
}

func TestBodyService_ParsePlainText(t *testing.T) {
	service := NewBodyService()

	textBody := []byte("This is plain text content")
	bodyInfo := service.ParseBody(textBody, "text/plain")

	if bodyInfo == nil {
		t.Fatal("Expected body info to be non-nil")
	}

	if bodyInfo.IsBinary {
		t.Error("Expected IsBinary to be false for text")
	}

	content, ok := bodyInfo.Content.(string)
	if !ok {
		t.Fatal("Expected content to be a string")
	}

	if content != string(textBody) {
		t.Errorf("Expected content '%s', got '%s'", string(textBody), content)
	}
}

func TestBodyService_ParseBinaryData(t *testing.T) {
	service := NewBodyService()

	// Create binary data with null bytes
	binaryData := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00}
	bodyInfo := service.ParseBody(binaryData, "application/octet-stream")

	if bodyInfo == nil {
		t.Fatal("Expected body info to be non-nil")
	}

	if !bodyInfo.IsBinary {
		t.Error("Expected IsBinary to be true for binary data")
	}

	// Verify content is base64 encoded
	content, ok := bodyInfo.Content.(string)
	if !ok {
		t.Fatal("Expected content to be a string (base64)")
	}

	// Decode and verify
	decoded, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		t.Fatalf("Failed to decode base64: %v", err)
	}

	if len(decoded) != len(binaryData) {
		t.Errorf("Expected decoded length %d, got %d", len(binaryData), len(decoded))
	}
}

func TestBodyService_MaxBodySize(t *testing.T) {
	service := NewBodyService()

	// Create body larger than max size
	largeBody := make([]byte, service.maxBodySize+1000)
	for i := range largeBody {
		largeBody[i] = 'A'
	}

	bodyInfo := service.ParseBody(largeBody, "text/plain")

	if bodyInfo == nil {
		t.Fatal("Expected body info to be non-nil")
	}

	if !bodyInfo.Truncated {
		t.Error("Expected Truncated to be true for oversized body")
	}

	if bodyInfo.Size != len(largeBody) {
		t.Errorf("Expected size to be original size %d, got %d", len(largeBody), bodyInfo.Size)
	}
}

func TestBodyService_EmptyBody(t *testing.T) {
	service := NewBodyService()

	bodyInfo := service.ParseBody([]byte{}, "application/json")

	if bodyInfo != nil {
		t.Error("Expected body info to be nil for empty body")
	}
}

func TestBodyService_InvalidJSON(t *testing.T) {
	service := NewBodyService()

	invalidJSON := []byte(`{"name": invalid}`)
	bodyInfo := service.ParseBody(invalidJSON, "application/json")

	if bodyInfo == nil {
		t.Fatal("Expected body info to be non-nil")
	}

	// Should fallback to string representation
	content, ok := bodyInfo.Content.(string)
	if !ok {
		t.Fatal("Expected content to be a string for invalid JSON")
	}

	if content != string(invalidJSON) {
		t.Errorf("Expected content to be original string, got %s", content)
	}
}

func TestBodyService_IsBinaryData(t *testing.T) {
	service := NewBodyService()

	tests := []struct {
		name     string
		data     []byte
		expected bool
	}{
		{
			name:     "plain text",
			data:     []byte("Hello, World!"),
			expected: false,
		},
		{
			name:     "text with newlines",
			data:     []byte("Line 1\nLine 2\nLine 3"),
			expected: false,
		},
		{
			name:     "binary with null bytes",
			data:     []byte{0x00, 0x01, 0x02, 0x03},
			expected: true,
		},
		{
			name:     "invalid UTF-8",
			data:     []byte{0xFF, 0xFE, 0xFD},
			expected: true,
		},
		{
			name:     "JSON",
			data:     []byte(`{"key":"value"}`),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.isBinaryData(tt.data)
			if result != tt.expected {
				t.Errorf("Expected isBinaryData to be %v, got %v", tt.expected, result)
			}
		})
	}
}
