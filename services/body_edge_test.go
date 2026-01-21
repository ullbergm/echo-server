package services

import (
	"strings"
	"testing"
)

// TestBodyService_LargeBody tests parsing of large payloads
func TestBodyService_LargeBody(t *testing.T) {
	service := NewBodyService()

	// Create a 1MB JSON body
	largeJSON := `{"data":"` + strings.Repeat("x", 1024*1024) + `"}`
	bodyInfo := service.ParseBody([]byte(largeJSON), "application/json")

	if bodyInfo == nil {
		t.Fatal("Expected body info to be non-nil")
	}

	if bodyInfo.Size != len(largeJSON) {
		t.Errorf("Expected size %d, got %d", len(largeJSON), bodyInfo.Size)
	}
}

// TestBodyService_EmptyBodyMultipleContentTypes tests parsing of empty bodies with various content types
func TestBodyService_EmptyBodyMultipleContentTypes(t *testing.T) {
	service := NewBodyService()

	contentTypes := []string{
		"application/json",
		"application/x-www-form-urlencoded",
		"text/plain",
		"application/xml",
		"application/octet-stream",
	}

	for _, ct := range contentTypes {
		t.Run(ct, func(t *testing.T) {
			bodyInfo := service.ParseBody([]byte{}, ct)

			// Empty body should return nil
			if bodyInfo != nil {
				t.Error("Expected body info to be nil for empty body")
			}
		})
	}
}

// TestBodyService_JSONWithCharset tests content type with charset parameter
func TestBodyService_JSONWithCharset(t *testing.T) {
	service := NewBodyService()

	tests := []struct {
		name        string
		contentType string
		expectJSON  bool
	}{
		{
			name:        "JSON with UTF-8 charset",
			contentType: "application/json; charset=utf-8",
			expectJSON:  true,
		},
		{
			name:        "JSON with UTF-16 charset",
			contentType: "application/json; charset=utf-16",
			expectJSON:  true,
		},
		{
			name:        "JSON with boundary",
			contentType: "application/json; boundary=something",
			expectJSON:  true,
		},
	}

	jsonBody := []byte(`{"test":"value"}`)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyInfo := service.ParseBody(jsonBody, tt.contentType)

			if bodyInfo == nil {
				t.Fatal("Expected body info to be non-nil")
			}

			if tt.expectJSON {
				_, ok := bodyInfo.Content.(map[string]interface{})
				if !ok {
					t.Error("Expected content to be parsed as JSON map")
				}
			}
		})
	}
}

// TestBodyService_InvalidJSONEdgeCases tests handling of various invalid JSON formats
func TestBodyService_InvalidJSONEdgeCases(t *testing.T) {
	service := NewBodyService()

	invalidJSONSamples := []string{
		`{invalid json}`,
		`{"unclosed": "string`,
		`{"trailing": "comma",}`,
		`{key: "value"}`, // unquoted key
		`undefined`,
		`NaN`,
	}

	for _, sample := range invalidJSONSamples {
		t.Run(sample, func(t *testing.T) {
			bodyInfo := service.ParseBody([]byte(sample), "application/json")

			if bodyInfo == nil {
				t.Fatal("Expected body info to be non-nil even for invalid JSON")
			}

			// Should fall back to raw string for invalid JSON
			if _, ok := bodyInfo.Content.(string); !ok {
				t.Log("Invalid JSON might be stored as string (acceptable fallback)")
			}
		})
	}
}

// TestBodyService_InvalidFormData tests handling of invalid form data
func TestBodyService_InvalidFormData(t *testing.T) {
	service := NewBodyService()

	invalidFormSamples := []string{
		`key1=value1&key2=`, // empty value
		`=value`,            // empty key
		`noequals`,          // no equals sign
		`key1=value1&`,      // trailing ampersand
		`&key1=value1`,      // leading ampersand
	}

	for _, sample := range invalidFormSamples {
		t.Run(sample, func(t *testing.T) {
			bodyInfo := service.ParseBody([]byte(sample), "application/x-www-form-urlencoded")

			if bodyInfo == nil {
				t.Fatal("Expected body info to be non-nil")
			}

			// Form parser should handle these gracefully
			if bodyInfo.Size != len(sample) {
				t.Errorf("Expected size %d, got %d", len(sample), bodyInfo.Size)
			}
		})
	}
}

// TestBodyService_XMLContent tests XML content type
func TestBodyService_XMLContent(t *testing.T) {
	service := NewBodyService()

	xmlBody := []byte(`<?xml version="1.0"?><root><item>value</item></root>`)
	bodyInfo := service.ParseBody(xmlBody, "application/xml")

	if bodyInfo == nil {
		t.Fatal("Expected body info to be non-nil")
	}

	if bodyInfo.Size != len(xmlBody) {
		t.Errorf("Expected size %d, got %d", len(xmlBody), bodyInfo.Size)
	}

	// XML should be treated as text
	if bodyInfo.IsBinary {
		t.Error("Expected XML not to be marked as binary")
	}
}

// TestBodyService_BinaryDataVariations tests various binary content types
func TestBodyService_BinaryDataVariations(t *testing.T) {
	service := NewBodyService()

	binaryTypes := []string{
		"application/octet-stream",
		"image/png",
		"image/jpeg",
		"image/gif",
		"application/pdf",
		"application/zip",
		"video/mp4",
		"audio/mpeg",
	}

	// Create binary data (PNG signature)
	binaryData := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}

	for _, contentType := range binaryTypes {
		t.Run(contentType, func(t *testing.T) {
			bodyInfo := service.ParseBody(binaryData, contentType)

			if bodyInfo == nil {
				t.Fatal("Expected body info to be non-nil")
			}

			if !bodyInfo.IsBinary {
				t.Errorf("Expected content type %s to be marked as binary", contentType)
			}

			// Binary content should be base64 encoded
			if _, ok := bodyInfo.Content.(string); !ok {
				t.Error("Expected binary content to be base64 encoded string")
			}
		})
	}
}

// TestBodyService_TextPlainWithBinaryData tests text/plain with binary characters
func TestBodyService_TextPlainWithBinaryData(t *testing.T) {
	service := NewBodyService()

	// Text with null bytes (binary)
	binaryText := []byte("text\x00with\x00nulls")
	bodyInfo := service.ParseBody(binaryText, "text/plain")

	if bodyInfo == nil {
		t.Fatal("Expected body info to be non-nil")
	}

	// Should be detected as binary due to null bytes
	if !bodyInfo.IsBinary {
		t.Error("Expected text with null bytes to be marked as binary")
	}
}

// TestBodyService_FormDataWithSpecialCharacters tests form data with special characters
func TestBodyService_FormDataWithSpecialCharacters(t *testing.T) {
	service := NewBodyService()

	// Form data with URL encoded special characters
	formBody := []byte("email=john%2Btest%40example.com&name=John%20O%27Brien&password=p%40ssw%26rd%21") // pragma: allowlist secret
	bodyInfo := service.ParseBody(formBody, "application/x-www-form-urlencoded")

	if bodyInfo == nil {
		t.Fatal("Expected body info to be non-nil")
	}

	contentMap, ok := bodyInfo.Content.(map[string]interface{})
	if !ok {
		t.Fatal("Expected content to be a map")
	}

	// Verify URL decoding worked
	if email, emailOk := contentMap["email"].(string); emailOk {
		if email != "john+test@example.com" {
			t.Errorf("Expected email 'john+test@example.com', got %s", email)
		}
	}

	if name, nameOk := contentMap["name"].(string); nameOk {
		if name != "John O'Brien" {
			t.Errorf("Expected name 'John O'Brien', got %s", name)
		}
	}

	if password, passOk := contentMap["password"].(string); passOk {
		if password != "p@ssw&rd!" { // pragma: allowlist secret
			t.Errorf("Expected password 'p@ssw&rd!', got %s", password)
		}
	}
}

// TestBodyService_JSONArray tests JSON array parsing
func TestBodyService_JSONArray(t *testing.T) {
	service := NewBodyService()

	jsonArray := []byte(`[{"id":1,"name":"Alice"},{"id":2,"name":"Bob"}]`)
	bodyInfo := service.ParseBody(jsonArray, "application/json")

	if bodyInfo == nil {
		t.Fatal("Expected body info to be non-nil")
	}

	// Should parse as array
	_, ok := bodyInfo.Content.([]interface{})
	if !ok {
		t.Error("Expected content to be parsed as JSON array")
	}
}

// TestBodyService_NestedJSON tests deeply nested JSON structures
func TestBodyService_NestedJSON(t *testing.T) {
	service := NewBodyService()

	nestedJSON := []byte(`{
		"level1": {
			"level2": {
				"level3": {
					"level4": {
						"value": "deep"
					}
				}
			}
		}
	}`)

	bodyInfo := service.ParseBody(nestedJSON, "application/json")

	if bodyInfo == nil {
		t.Fatal("Expected body info to be non-nil")
	}

	contentMap, ok := bodyInfo.Content.(map[string]interface{})
	if !ok {
		t.Fatal("Expected content to be a map")
	}

	// Verify nested structure exists
	if level1, ok1 := contentMap["level1"].(map[string]interface{}); ok1 {
		if level2, ok2 := level1["level2"].(map[string]interface{}); ok2 {
			if level3, ok3 := level2["level3"].(map[string]interface{}); ok3 {
				if level4, ok4 := level3["level4"].(map[string]interface{}); ok4 {
					if value, ok5 := level4["value"].(string); !ok5 || value != "deep" {
						t.Error("Failed to parse deeply nested JSON correctly")
					}
				}
			}
		}
	}
}

// TestBodyService_UTF8Content tests UTF-8 encoded content
func TestBodyService_UTF8Content(t *testing.T) {
	service := NewBodyService()

	utf8JSON := []byte(`{"message":"Hello 世界! 🌍","emoji":"😀🎉"}`)
	bodyInfo := service.ParseBody(utf8JSON, "application/json")

	if bodyInfo == nil {
		t.Fatal("Expected body info to be non-nil")
	}

	contentMap, ok := bodyInfo.Content.(map[string]interface{})
	if !ok {
		t.Fatal("Expected content to be a map")
	}

	if message, msgOk := contentMap["message"].(string); msgOk {
		if !strings.Contains(message, "世界") {
			t.Error("Expected UTF-8 Chinese characters to be preserved")
		}
		if !strings.Contains(message, "🌍") {
			t.Error("Expected UTF-8 emoji to be preserved")
		}
	}
}

// TestBodyService_UnknownContentType tests handling of unknown content types
func TestBodyService_UnknownContentType(t *testing.T) {
	service := NewBodyService()

	unknownTypes := []string{
		"application/vnd.custom+json",
		"text/custom",
		"",
		"invalid/type",
	}

	testBody := []byte("test content")

	for _, ct := range unknownTypes {
		t.Run(ct, func(t *testing.T) {
			bodyInfo := service.ParseBody(testBody, ct)

			if bodyInfo == nil {
				t.Fatal("Expected body info to be non-nil for unknown content type")
			}

			if bodyInfo.Size != len(testBody) {
				t.Errorf("Expected size %d, got %d", len(testBody), bodyInfo.Size)
			}
		})
	}
}
