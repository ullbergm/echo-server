package services

import (
	"testing"
)

// FuzzBodyParseJSON tests JSON body parsing with random inputs
func FuzzBodyParseJSON(f *testing.F) {
	seeds := [][]byte{
		// Valid JSON
		[]byte(`{"name":"John","age":30}`),
		[]byte(`[1,2,3,4,5]`),
		[]byte(`"string"`),
		[]byte(`123`),
		[]byte(`true`),
		[]byte(`null`),
		// Nested structures
		[]byte(`{"a":{"b":{"c":{"d":"deep"}}}}`),
		[]byte(`[[[[[1]]]]]`),
		// Large numbers
		[]byte(`{"num":9007199254740993}`),
		[]byte(`{"float":1.7976931348623157e+308}`),
		// Unicode
		[]byte(`{"unicode":"日本語テスト"}`),
		[]byte(`{"emoji":"🔥🚀💻"}`),
		// Edge cases
		[]byte(`{}`),
		[]byte(`[]`),
		[]byte(``),
		// Invalid JSON
		[]byte(`{invalid}`),
		[]byte(`{"unclosed":`),
		[]byte(`{key: "no quotes"}`),
		[]byte(`{'single': 'quotes'}`),
		// Potentially dangerous
		[]byte(`{"__proto__":"polluted"}`),
		[]byte(`{"constructor":"test"}`),
		// Whitespace
		[]byte(`   {"spaced": "json"}   `),
		// Null bytes
		[]byte("{\"\x00null\": \"bytes\"}"),
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	service := NewBodyService()

	f.Fuzz(func(t *testing.T, body []byte) {
		// Should not panic regardless of input
		result := service.ParseBody(body, "application/json")

		// Empty body should return nil
		if len(body) == 0 {
			if result != nil {
				t.Error("Expected nil for empty body")
			}
			return
		}

		// Non-empty body should return non-nil result
		if result == nil {
			t.Error("Expected non-nil result for non-empty body")
			return
		}

		// Size should match
		if result.Size != len(body) && !result.Truncated {
			t.Errorf("Size mismatch: expected %d, got %d", len(body), result.Size)
		}
	})
}

// FuzzBodyParseFormURLEncoded tests URL-encoded form parsing
func FuzzBodyParseFormURLEncoded(f *testing.F) {
	seeds := [][]byte{
		[]byte("key=value"),
		[]byte("key1=value1&key2=value2"),
		[]byte("encoded=%20%21%40%23"),
		[]byte("unicode=%E6%97%A5%E6%9C%AC%E8%AA%9E"),
		[]byte("array=1&array=2&array=3"),
		[]byte("empty="),
		[]byte("=nokey"),
		[]byte(""),
		[]byte("&&&&"),
		[]byte("key=value=with=equals"),
		[]byte("special=<script>alert('xss')</script>"),
		[]byte("null=%00"),
		[]byte("percent=%"),
		[]byte("invalid=%ZZ"),
		// Very long values
		[]byte("long=" + string(make([]byte, 1000))),
		// SQL injection patterns
		[]byte("id=1' OR '1'='1"),
		// Path traversal
		[]byte("file=../../../etc/passwd"),
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	service := NewBodyService()

	f.Fuzz(func(t *testing.T, body []byte) {
		// Should not panic
		result := service.ParseBody(body, "application/x-www-form-urlencoded")

		if len(body) == 0 {
			if result != nil {
				t.Error("Expected nil for empty body")
			}
			return
		}

		if result == nil {
			t.Error("Expected non-nil result for non-empty body")
		}
	})
}

// FuzzBodyParseMultipart tests multipart form data parsing
func FuzzBodyParseMultipart(f *testing.F) {
	// Valid multipart body
	validMultipart := []byte("--boundary\r\n" +
		"Content-Disposition: form-data; name=\"field1\"\r\n\r\n" +
		"value1\r\n" +
		"--boundary--\r\n")

	seeds := [][]byte{
		validMultipart,
		[]byte("--boundary\r\n--boundary--"),
		[]byte(""),
		[]byte("no boundary markers"),
		[]byte("--\r\n--"),
		// Malformed headers
		[]byte("--boundary\r\nmalformed header\r\n\r\ndata\r\n--boundary--"),
		// Missing final boundary
		[]byte("--boundary\r\nContent-Disposition: form-data; name=\"test\"\r\n\r\nvalue"),
		// File upload simulation
		[]byte("--boundary\r\n" +
			"Content-Disposition: form-data; name=\"file\"; filename=\"test.txt\"\r\n" +
			"Content-Type: text/plain\r\n\r\n" +
			"file contents\r\n" +
			"--boundary--"),
		// Binary-like content
		[]byte("--boundary\r\n" +
			"Content-Disposition: form-data; name=\"binary\"; filename=\"test.bin\"\r\n\r\n" +
			"\x00\x01\x02\x03\r\n" +
			"--boundary--"),
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	service := NewBodyService()

	f.Fuzz(func(t *testing.T, body []byte) {
		contentType := "multipart/form-data; boundary=boundary"
		// Should not panic
		result := service.ParseBody(body, contentType)

		if len(body) == 0 && result != nil {
			t.Error("Expected nil for empty body")
		}
	})
}

// FuzzBodyParseBinaryDetection tests binary data detection
func FuzzBodyParseBinaryDetection(f *testing.F) {
	seeds := [][]byte{
		// Pure text
		[]byte("Hello, World!"),
		// Text with newlines
		[]byte("Line1\nLine2\r\nLine3"),
		// Binary data
		{0x00, 0x01, 0x02, 0x03, 0xFF},
		// Mixed content
		append([]byte("text"), 0x00, 0x01),
		// High ASCII
		{0x80, 0x81, 0x82, 0xFF},
		// Control characters
		{0x01, 0x02, 0x03, 0x04, 0x05},
		// Valid UTF-8 multibyte
		[]byte("日本語"),
		[]byte("🔥🚀💻"),
		// Invalid UTF-8
		{0xFF, 0xFE, 0x00, 0x01},
		// PDF header
		[]byte("%PDF-1.4"),
		// PNG header
		{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A},
		// JPEG header
		{0xFF, 0xD8, 0xFF, 0xE0},
		// ZIP header
		{0x50, 0x4B, 0x03, 0x04},
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	service := NewBodyService()

	f.Fuzz(func(t *testing.T, body []byte) {
		// Should not panic
		result := service.ParseBody(body, "application/octet-stream")

		if len(body) == 0 {
			if result != nil {
				t.Error("Expected nil for empty body")
			}
			return
		}

		if result == nil {
			t.Error("Expected non-nil result for non-empty body")
		}
	})
}

// FuzzBodyParseXML tests XML body parsing
func FuzzBodyParseXML(f *testing.F) {
	seeds := [][]byte{
		// Valid XML
		[]byte(`<?xml version="1.0"?><root><item>value</item></root>`),
		[]byte(`<simple>text</simple>`),
		// Self-closing tags
		[]byte(`<empty/>`),
		// Attributes
		[]byte(`<tag attr="value">content</tag>`),
		// CDATA
		[]byte(`<data><![CDATA[<not parsed>]]></data>`),
		// Comments
		[]byte(`<root><!-- comment --><item/></root>`),
		// Invalid XML
		[]byte(`<unclosed>`),
		[]byte(`not xml at all`),
		[]byte(`<tag>mismatched</other>`),
		// XXE attempt (should be handled safely)
		[]byte(`<?xml version="1.0"?><!DOCTYPE foo [<!ENTITY xxe SYSTEM "file:///etc/passwd">]><root>&xxe;</root>`),
		// Empty
		[]byte(``),
		// Unicode in XML
		[]byte(`<root>日本語テスト</root>`),
		// Very deeply nested
		[]byte(`<a><b><c><d><e><f><g><h>deep</h></g></f></e></d></c></b></a>`),
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	service := NewBodyService()

	f.Fuzz(func(t *testing.T, body []byte) {
		// Should not panic
		result := service.ParseBody(body, "application/xml")

		if len(body) == 0 && result != nil {
			t.Error("Expected nil for empty body")
		}
	})
}

// FuzzBodyParseContentType tests various content type handling
func FuzzBodyParseContentType(f *testing.F) {
	contentTypes := []string{
		"application/json",
		"application/xml",
		"text/xml",
		"text/plain",
		"text/html",
		"application/x-www-form-urlencoded",
		"multipart/form-data; boundary=----WebKitFormBoundary",
		"application/octet-stream",
		"image/png",
		"audio/mpeg",
		"",
		"invalid",
		"application/json; charset=utf-8",
		"text/plain; charset=iso-8859-1",
		// Malformed
		"application/json;",
		"application/json; ",
		"application/json; charset=",
		// Unknown types
		"application/x-custom-type",
		"x-custom/type",
	}

	for _, ct := range contentTypes {
		f.Add(ct)
	}

	service := NewBodyService()
	testBody := []byte(`{"test": "data"}`)

	f.Fuzz(func(t *testing.T, contentType string) {
		// Should not panic regardless of content type
		result := service.ParseBody(testBody, contentType)

		if result == nil {
			t.Error("Expected non-nil result for non-empty body")
		}
	})
}

// FuzzBodyLargeInput tests handling of large inputs
func FuzzBodyLargeInput(f *testing.F) {
	// Start with various sizes
	sizes := []int{0, 1, 100, 1000, 10000}
	for _, size := range sizes {
		data := make([]byte, size)
		for i := range data {
			data[i] = byte('a' + (i % 26))
		}
		f.Add(data)
	}

	service := NewBodyService()

	f.Fuzz(func(t *testing.T, body []byte) {
		// Should not panic even with large inputs
		result := service.ParseBody(body, "text/plain")

		if len(body) == 0 {
			if result != nil {
				t.Error("Expected nil for empty body")
			}
			return
		}

		if result == nil {
			t.Error("Expected non-nil result for non-empty body")
			return
		}

		// Verify truncation works correctly
		if result.Size > service.maxBodySize && !result.Truncated {
			t.Error("Large body should be marked as truncated")
		}
	})
}
