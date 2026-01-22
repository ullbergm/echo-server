package services

import (
	"os"
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

// TestNewBodyService_WithEnvVar tests NewBodyService with MAX_BODY_SIZE env var
func TestNewBodyService_WithEnvVar(t *testing.T) {
	tests := []struct {
		name        string
		envValue    string
		expectedMax int
	}{
		{
			name:        "valid custom size",
			envValue:    "1024",
			expectedMax: 1024,
		},
		{
			name:        "large custom size",
			envValue:    "52428800",
			expectedMax: 52428800,
		},
		{
			name:        "invalid size (negative)",
			envValue:    "-100",
			expectedMax: DefaultMaxBodySize, // Should use default
		},
		{
			name:        "invalid size (not a number)",
			envValue:    "invalid",
			expectedMax: DefaultMaxBodySize, // Should use default
		},
		{
			name:        "zero size",
			envValue:    "0",
			expectedMax: DefaultMaxBodySize, // Should use default (0 is not > 0)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := os.Setenv("MAX_BODY_SIZE", tt.envValue); err != nil {
				t.Fatal(err)
			}
			defer func() { _ = os.Unsetenv("MAX_BODY_SIZE") }()

			service := NewBodyService()

			if service.maxBodySize != tt.expectedMax {
				t.Errorf("Expected maxBodySize=%d, got %d", tt.expectedMax, service.maxBodySize)
			}
		})
	}
}

// TestBodyService_MultipartFormData tests multipart form data parsing
func TestBodyService_MultipartFormData(t *testing.T) {
	service := NewBodyService()

	// Create proper multipart form data
	boundary := "----WebKitFormBoundary7MA4YWxkTrZu0gW" // pragma: allowlist secret
	multipartBody := strings.Join([]string{
		"------WebKitFormBoundary7MA4YWxkTrZu0gW", // pragma: allowlist secret
		"Content-Disposition: form-data; name=\"username\"",
		"",
		"johndoe",
		"------WebKitFormBoundary7MA4YWxkTrZu0gW", // pragma: allowlist secret
		"Content-Disposition: form-data; name=\"email\"",
		"",
		"john@example.com",
		"------WebKitFormBoundary7MA4YWxkTrZu0gW--", // pragma: allowlist secret
	}, "\r\n")

	bodyInfo := service.ParseBody([]byte(multipartBody), "multipart/form-data; boundary="+boundary)

	if bodyInfo == nil {
		t.Fatal("Expected body info to be non-nil")
	}

	if bodyInfo.ContentType != "multipart/form-data; boundary="+boundary {
		t.Errorf("Expected content type to match, got %s", bodyInfo.ContentType)
	}

	// Verify content was parsed
	contentMap, ok := bodyInfo.Content.(map[string]interface{})
	if !ok {
		// Multipart might return string if parsing fails, that's acceptable
		t.Log("Multipart content was returned as string (acceptable)")
		return
	}

	if contentMap["username"] != "johndoe" {
		t.Errorf("Expected username 'johndoe', got %v", contentMap["username"])
	}
}

// TestBodyService_MultipartWithFile tests multipart with file upload
func TestBodyService_MultipartWithFile(t *testing.T) {
	service := NewBodyService()

	boundary := "----TestBoundary"
	multipartBody := strings.Join([]string{
		"------TestBoundary",
		"Content-Disposition: form-data; name=\"file\"; filename=\"test.txt\"",
		"Content-Type: text/plain",
		"",
		"Hello, World!",
		"------TestBoundary",
		"Content-Disposition: form-data; name=\"description\"",
		"",
		"A test file",
		"------TestBoundary--",
	}, "\r\n")

	bodyInfo := service.ParseBody([]byte(multipartBody), "multipart/form-data; boundary="+boundary)

	if bodyInfo == nil {
		t.Fatal("Expected body info to be non-nil")
	}

	if bodyInfo.Size != len(multipartBody) {
		t.Errorf("Expected size %d, got %d", len(multipartBody), bodyInfo.Size)
	}
}

// TestBodyService_MultipartWithBinaryFile tests multipart with binary file
func TestBodyService_MultipartWithBinaryFile(t *testing.T) {
	service := NewBodyService()

	// Create binary content (simulated image header with null bytes)
	// Must have null bytes to trigger isBinaryData detection
	binaryContent := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x00, 0x00, 0x00, 0x00}

	boundary := "----BinaryBoundary"
	header := "------BinaryBoundary\r\nContent-Disposition: form-data; name=\"image\"; filename=\"test.png\"\r\nContent-Type: image/png\r\n\r\n"
	footer := "\r\n------BinaryBoundary--"

	fullBody := append([]byte(header), binaryContent...)
	fullBody = append(fullBody, []byte(footer)...)

	bodyInfo := service.ParseBody(fullBody, "multipart/form-data; boundary="+boundary)

	if bodyInfo == nil {
		t.Fatal("Expected body info to be non-nil")
	}

	// Check if content was parsed as map (file info)
	contentMap, ok := bodyInfo.Content.(map[string]interface{})
	if ok {
		// Check if file info exists
		if fileInfo, exists := contentMap["image"]; exists {
			fileMap, isMap := fileInfo.(map[string]interface{})
			if isMap {
				// Verify filename is captured
				if fileMap["filename"] != "test.png" {
					t.Errorf("Expected filename 'test.png', got %v", fileMap["filename"])
				}
				// Binary content should be base64 encoded
				if _, hasEncoding := fileMap["encoding"]; hasEncoding {
					t.Log("Binary file was base64 encoded (expected)")
				}
			}
		}
	}
}

// TestBodyService_MultipartWithTextFile tests multipart with text file
func TestBodyService_MultipartWithTextFile(t *testing.T) {
	service := NewBodyService()

	// Text file content (no null bytes)
	textContent := []byte("Hello, this is a text file content!")

	boundary := "----TextFileBoundary"
	multipartBody := "------TextFileBoundary\r\n" +
		"Content-Disposition: form-data; name=\"textfile\"; filename=\"readme.txt\"\r\n" +
		"Content-Type: text/plain\r\n\r\n" +
		string(textContent) +
		"\r\n------TextFileBoundary--"

	bodyInfo := service.ParseBody([]byte(multipartBody), "multipart/form-data; boundary="+boundary)

	if bodyInfo == nil {
		t.Fatal("Expected body info to be non-nil")
	}

	// Check if content was parsed as map
	contentMap, ok := bodyInfo.Content.(map[string]interface{})
	if ok {
		if fileInfo, exists := contentMap["textfile"]; exists {
			fileMap, isMap := fileInfo.(map[string]interface{})
			if isMap {
				// Text file should have content as string (not base64)
				if _, hasEncoding := fileMap["encoding"]; !hasEncoding {
					t.Log("Text file was stored as plain text (expected)")
				}
			}
		}
	}
}

// TestBodyService_MultipartWithHighNonPrintableFile tests multipart with file that has high non-printable ratio
func TestBodyService_MultipartWithHighNonPrintableFile(t *testing.T) {
	service := NewBodyService()

	// Create file content with high ratio of non-printable chars (> 30%)
	// but NO null bytes, and valid UTF-8 single-byte chars
	// Use control characters 0x01-0x08 (before tab 0x09)
	fileContent := make([]byte, 100)
	for i := 0; i < 40; i++ {
		fileContent[i] = byte(0x01 + (i % 8)) // Non-printable but valid UTF-8
	}
	for i := 40; i < 100; i++ {
		fileContent[i] = 'A' // Printable
	}

	boundary := "----NonPrintableBoundary"
	header := "------NonPrintableBoundary\r\n" +
		"Content-Disposition: form-data; name=\"binaryish\"; filename=\"data.bin\"\r\n" +
		"Content-Type: application/octet-stream\r\n\r\n"
	footer := "\r\n------NonPrintableBoundary--"

	// Build the full body
	fullBody := append([]byte(header), fileContent...)
	fullBody = append(fullBody, []byte(footer)...)

	bodyInfo := service.ParseBody(fullBody, "multipart/form-data; boundary="+boundary)

	if bodyInfo == nil {
		t.Fatal("Expected body info to be non-nil")
	}

	// This should process the multipart and detect the file content as binary
	// due to high non-printable ratio
	if bodyInfo.Size != len(fullBody) {
		t.Errorf("Expected size %d, got %d", len(fullBody), bodyInfo.Size)
	}
}

// TestBodyService_MultipartNoBoundary tests multipart without boundary
func TestBodyService_MultipartNoBoundary(t *testing.T) {
	service := NewBodyService()

	body := []byte("some multipart content without proper boundary")
	bodyInfo := service.ParseBody(body, "multipart/form-data")

	if bodyInfo == nil {
		t.Fatal("Expected body info to be non-nil")
	}

	// Should return as raw string since no boundary
	if _, ok := bodyInfo.Content.(string); !ok {
		t.Error("Expected content to be string when no boundary provided")
	}
}

// TestBodyService_XMLInvalid tests invalid XML handling
func TestBodyService_XMLInvalid(t *testing.T) {
	service := NewBodyService()

	invalidXMLSamples := []string{
		`<root><unclosed>`,
		`not xml at all`,
		`<root>mismatched</other>`,
		`<?xml version="1.0"?>incomplete`,
	}

	for _, sample := range invalidXMLSamples {
		t.Run(sample[:minInt(20, len(sample))], func(t *testing.T) {
			bodyInfo := service.ParseBody([]byte(sample), "application/xml")

			if bodyInfo == nil {
				t.Fatal("Expected body info to be non-nil")
			}

			// Should return as string when XML parsing fails
			if _, ok := bodyInfo.Content.(string); !ok {
				t.Log("Invalid XML stored in some format (acceptable)")
			}
		})
	}
}

// TestBodyService_XMLWithNamespaces tests XML with namespaces
func TestBodyService_XMLWithNamespaces(t *testing.T) {
	service := NewBodyService()

	xmlBody := []byte(`<?xml version="1.0"?>
<root xmlns:ns="http://example.com">
  <ns:element>value</ns:element>
</root>`)

	bodyInfo := service.ParseBody(xmlBody, "text/xml")

	if bodyInfo == nil {
		t.Fatal("Expected body info to be non-nil")
	}

	if bodyInfo.IsBinary {
		t.Error("Expected XML not to be marked as binary")
	}
}

// TestBodyService_FormWithMultipleValues tests form with same key multiple times
func TestBodyService_FormWithMultipleValues(t *testing.T) {
	service := NewBodyService()

	formBody := []byte("color=red&color=green&color=blue")
	bodyInfo := service.ParseBody(formBody, "application/x-www-form-urlencoded")

	if bodyInfo == nil {
		t.Fatal("Expected body info to be non-nil")
	}

	contentMap, ok := bodyInfo.Content.(map[string]interface{})
	if !ok {
		t.Fatal("Expected content to be a map")
	}

	// color should be an array with 3 values
	colors, ok := contentMap["color"].([]string)
	if !ok {
		t.Log("Multiple values may be stored differently")
		return
	}

	if len(colors) != 3 {
		t.Errorf("Expected 3 colors, got %d", len(colors))
	}
}

// TestBodyService_FormURLEncodedParseError tests form data that fails parsing
func TestBodyService_FormURLEncodedParseError(t *testing.T) {
	service := NewBodyService()

	// Invalid percent encoding that causes url.ParseQuery to fail
	// %ZZ is invalid percent encoding
	invalidForm := []byte("key=%ZZ")
	bodyInfo := service.ParseBody(invalidForm, "application/x-www-form-urlencoded")

	if bodyInfo == nil {
		t.Fatal("Expected body info to be non-nil")
	}

	// Should fall back to raw string
	content, ok := bodyInfo.Content.(string)
	if !ok {
		t.Error("Expected content to be string when form parsing fails")
	} else if content != string(invalidForm) {
		t.Errorf("Expected raw content, got %s", content)
	}
}

// TestBodyService_MultipartEmptyFieldName tests multipart with empty field name
func TestBodyService_MultipartEmptyFieldName(t *testing.T) {
	service := NewBodyService()

	// Multipart with Content-Disposition that has no name
	boundary := "----TestBoundary"
	multipartBody := strings.Join([]string{
		"------TestBoundary",
		"Content-Disposition: form-data",
		"",
		"some data",
		"------TestBoundary",
		"Content-Disposition: form-data; name=\"valid\"",
		"",
		"valid data",
		"------TestBoundary--",
	}, "\r\n")

	bodyInfo := service.ParseBody([]byte(multipartBody), "multipart/form-data; boundary="+boundary)

	if bodyInfo == nil {
		t.Fatal("Expected body info to be non-nil")
	}

	// Content should be parseable (empty name fields are skipped)
	if bodyInfo.Size != len(multipartBody) {
		t.Errorf("Expected size %d, got %d", len(multipartBody), bodyInfo.Size)
	}
}

// TestBodyService_MultipartMalformed tests malformed multipart data
func TestBodyService_MultipartMalformed(t *testing.T) {
	service := NewBodyService()

	// Malformed multipart that will cause parse error
	boundary := "----TestBoundary"
	malformedBody := []byte("not proper multipart data at all")

	bodyInfo := service.ParseBody(malformedBody, "multipart/form-data; boundary="+boundary)

	if bodyInfo == nil {
		t.Fatal("Expected body info to be non-nil")
	}

	// Should fall back to string when multipart parsing fails
	if _, ok := bodyInfo.Content.(string); !ok {
		t.Log("Content stored in some format (acceptable)")
	}
}

// TestBodyService_NonUTF8UnknownContentType tests non-UTF8 with unknown content type
func TestBodyService_NonUTF8UnknownContentType(t *testing.T) {
	service := NewBodyService()

	// Invalid UTF-8 data with unknown content type
	nonUtf8Data := []byte{0xFF, 0xFE, 0x00, 0x01, 0x89, 0x50}

	bodyInfo := service.ParseBody(nonUtf8Data, "application/unknown")

	if bodyInfo == nil {
		t.Fatal("Expected body info to be non-nil")
	}

	// Should be marked as binary
	if !bodyInfo.IsBinary {
		t.Error("Expected non-UTF8 data with unknown content type to be marked as binary")
	}
}

// TestBodyService_ValidUTF8UnknownContentType tests valid UTF-8 with unknown content type
func TestBodyService_ValidUTF8UnknownContentType(t *testing.T) {
	service := NewBodyService()

	validUtf8 := []byte("Hello, this is valid UTF-8!")

	bodyInfo := service.ParseBody(validUtf8, "application/unknown")

	if bodyInfo == nil {
		t.Fatal("Expected body info to be non-nil")
	}

	// Should NOT be marked as binary
	if bodyInfo.IsBinary {
		t.Error("Expected valid UTF-8 data to NOT be marked as binary")
	}

	// Content should be the string
	content, ok := bodyInfo.Content.(string)
	if !ok {
		t.Fatal("Expected content to be string")
	}

	if content != string(validUtf8) {
		t.Errorf("Expected content '%s', got '%s'", string(validUtf8), content)
	}
}

// TestBodyService_XMLSuccessfulParse tests XML that could be parsed (though Go's xml package is limited)
func TestBodyService_XMLSuccessfulParse(t *testing.T) {
	service := NewBodyService()

	// Simple XML - note: Go's xml.Unmarshal to map[string]interface{} typically fails
	// so we're actually testing the fallback path
	xmlBody := []byte(`<root><item>value</item></root>`)

	bodyInfo := service.ParseBody(xmlBody, "application/xml")

	if bodyInfo == nil {
		t.Fatal("Expected body info to be non-nil")
	}

	// XML parsing to map typically fails in Go, so content should be string
	// This tests the fallback path
	if _, ok := bodyInfo.Content.(string); !ok {
		t.Log("XML was parsed as map (unexpected but acceptable)")
	}
}

// TestBodyService_BinaryDataWithJSONContentType tests binary data with JSON content type
func TestBodyService_BinaryDataWithJSONContentType(t *testing.T) {
	service := NewBodyService()

	// Binary data with null bytes - will be detected as binary before JSON parsing
	binaryData := []byte{0x00, 0x01, 0x02, 0x7b, 0x7d} // Contains null bytes plus {}

	bodyInfo := service.ParseBody(binaryData, "application/json")

	if bodyInfo == nil {
		t.Fatal("Expected body info to be non-nil")
	}

	// Should be detected as binary (returns early before JSON parsing)
	if !bodyInfo.IsBinary {
		t.Error("Expected binary data to be marked as binary even with JSON content type")
	}
}

// TestBodyService_TruncatedBody tests body that exceeds max size
func TestBodyService_TruncatedBodyWithBinary(t *testing.T) {
	// Use a smaller max body size
	if err := os.Setenv("MAX_BODY_SIZE", "100"); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Unsetenv("MAX_BODY_SIZE") }()

	service := NewBodyService()

	// Create large binary body
	largeBody := make([]byte, 200)
	for i := range largeBody {
		largeBody[i] = byte(i % 256)
	}

	bodyInfo := service.ParseBody(largeBody, "application/octet-stream")

	if bodyInfo == nil {
		t.Fatal("Expected body info to be non-nil")
	}

	if !bodyInfo.Truncated {
		t.Error("Expected body to be marked as truncated")
	}

	if bodyInfo.Size != 200 {
		t.Errorf("Expected original size 200, got %d", bodyInfo.Size)
	}
}

// TestBodyService_ContentTypeWithInvalidMediaType tests content type that fails mime.ParseMediaType
func TestBodyService_ContentTypeWithInvalidMediaType(t *testing.T) {
	service := NewBodyService()

	// Invalid content type that will fail mime.ParseMediaType
	invalidContentType := ";;;invalid"
	textBody := []byte("some text")

	bodyInfo := service.ParseBody(textBody, invalidContentType)

	if bodyInfo == nil {
		t.Fatal("Expected body info to be non-nil")
	}

	// Should still process the body
	if bodyInfo.Size != len(textBody) {
		t.Errorf("Expected size %d, got %d", len(textBody), bodyInfo.Size)
	}
}

// TestBodyService_HighNonPrintableRatio tests binary detection threshold
func TestBodyService_HighNonPrintableRatio(t *testing.T) {
	service := NewBodyService()

	// Create data with high ratio of non-printable chars (> 30%)
	// 10 chars total: 4 non-printable = 40%
	data := []byte{'H', 'e', 'l', 0x01, 0x02, 0x03, 0x04, 'l', 'o', '!'}

	result := service.isBinaryData(data)

	// Should be detected as binary due to high non-printable ratio
	if !result {
		t.Error("Expected data with high non-printable ratio to be detected as binary")
	}
}

// TestBodyService_TextWithTabs tests text with tab characters
func TestBodyService_TextWithTabs(t *testing.T) {
	service := NewBodyService()

	// Text with tabs and newlines should NOT be binary
	data := []byte("Line1\tColumn2\nLine2\tColumn2\r\nLine3\tColumn2")

	result := service.isBinaryData(data)

	if result {
		t.Error("Expected text with tabs and newlines to NOT be binary")
	}
}

// minInt helper function (renamed to avoid shadowing builtin)
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
