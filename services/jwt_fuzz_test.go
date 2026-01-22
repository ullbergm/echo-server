package services

import (
	"encoding/base64"
	"encoding/json"
	"testing"
)

// FuzzJWTDecode tests the JWT decoding with random inputs
func FuzzJWTDecode(f *testing.F) {
	// Add seed corpus with various JWT-like strings
	seedCorpus := []string{
		// Valid JWT structure (test fixture)
		"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c", // pragma: allowlist secret
		// Empty parts
		"..",
		"...",
		// Single part (test fixture)
		"eyJhbGciOiJIUzI1NiJ9", // pragma: allowlist secret
		// Two parts (test fixture)
		"eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0", // pragma: allowlist secret
		// Four parts
		"a.b.c.d",
		// Invalid base64
		"!!invalid!!.!!base64!!.signature",
		// Valid base64 but invalid JSON
		"dGVzdA.dGVzdA.signature",
		// Empty string
		"",
		// Whitespace
		"   ",
		// Bearer prefix (test fixture)
		"Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.signature", // pragma: allowlist secret
		// Unicode
		"eyJ0ZXN0Ijoiδοκιμή\"}",
		// Very long token (test fixture)
		"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9." + base64.RawURLEncoding.EncodeToString([]byte(`{"data":"`+string(make([]byte, 1000))+`"}`)) + ".signature", // pragma: allowlist secret
		// Null bytes (test fixture)
		"eyJhbGciOiJIUzI1NiJ9.eyJkYXRhIjoiXHUwMDAwIn0.signature", // pragma: allowlist secret
		// Special characters
		"<script>.alert.('xss')",
		// Base64URL specific characters (test fixture)
		"eyJ-_bGciOiJIUzI1NiJ9.eyJz-_WIiOiIxMjM0In0.sig", // pragma: allowlist secret
	}

	for _, seed := range seedCorpus {
		f.Add(seed)
	}

	service := NewJWTService()

	f.Fuzz(func(t *testing.T, token string) {
		// Create headers map with the fuzzed token
		headers := map[string]string{
			"Authorization": token,
		}

		// This should not panic regardless of input
		result := service.ExtractAndDecodeJWTs(headers)

		// Result should always be a valid map (possibly empty)
		if result == nil {
			t.Error("Expected non-nil map result")
		}
	})
}

// FuzzJWTDecodeWithBearer tests JWT decoding with Bearer prefix variations
func FuzzJWTDecodeWithBearer(f *testing.F) {
	seeds := []string{
		"Bearer token",
		"bearer token",
		"BEARER token",
		"Bearer  token",
		"Bearer",
		"Bearer ",
		"Bearertoken",
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	service := NewJWTService()

	f.Fuzz(func(t *testing.T, token string) {
		headers := map[string]string{
			"Authorization": token,
		}

		// Should not panic
		result := service.ExtractAndDecodeJWTs(headers)
		if result == nil {
			t.Error("Expected non-nil map result")
		}
	})
}

// FuzzJWTBase64Decode tests base64URL decoding robustness
func FuzzJWTBase64Decode(f *testing.F) {
	seeds := []string{
		// Valid base64URL
		"eyJhbGciOiJIUzI1NiJ9",
		// Standard base64 (with + and /)
		"eyJhbGciOiJIUzI1NiJ9+/==",
		// Padding variations
		"dGVzdA",
		"dGVzdA=",
		"dGVzdA==",
		// Invalid characters
		"invalid!@#$%",
		// Empty
		"",
		// Whitespace
		"   ",
		// Very long
		base64.RawURLEncoding.EncodeToString(make([]byte, 10000)),
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	service := NewJWTService()

	f.Fuzz(func(t *testing.T, part string) {
		// Construct a token with the fuzzed part as header
		token := part + ".eyJzdWIiOiJ0ZXN0In0.signature"

		headers := map[string]string{
			"Authorization": token,
		}

		// Should not panic
		result := service.ExtractAndDecodeJWTs(headers)
		if result == nil {
			t.Error("Expected non-nil map result")
		}
	})
}

// FuzzJWTMultipleHeaders tests with multiple header configurations
func FuzzJWTMultipleHeaders(f *testing.F) {
	f.Add("token1", "token2", "token3")
	f.Add("", "", "")
	f.Add("Bearer valid", "invalid", "")
	f.Add("a.b.c", "x.y.z", "1.2.3")

	service := NewJWTService()

	f.Fuzz(func(t *testing.T, auth, xjwt, xauth string) {
		headers := map[string]string{
			"Authorization": auth,
			"X-JWT-Token":   xjwt,
			"X-Auth-Token":  xauth,
		}

		// Should not panic
		result := service.ExtractAndDecodeJWTs(headers)
		if result == nil {
			t.Error("Expected non-nil map result")
		}
	})
}

// FuzzJWTPayloadContent tests with various JSON payload contents
func FuzzJWTPayloadContent(f *testing.F) {
	payloads := []string{
		`{"sub":"user"}`,
		`{"array":[1,2,3]}`,
		`{"nested":{"key":"value"}}`,
		`{"unicode":"日本語"}`,
		`{"emoji":"🔐"}`,
		`{"number":9007199254740993}`,
		`{"float":1.7976931348623157e+308}`,
		`{"null":null}`,
		`{"bool":true}`,
		`{}`,
	}

	for _, payload := range payloads {
		f.Add(payload)
	}

	service := NewJWTService()

	f.Fuzz(func(t *testing.T, payloadJSON string) {
		// Create a valid-looking JWT with the fuzzed payload
		header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256"}`))
		payload := base64.RawURLEncoding.EncodeToString([]byte(payloadJSON))
		token := header + "." + payload + ".signature"

		headers := map[string]string{
			"Authorization": token,
		}

		// Should not panic
		result := service.ExtractAndDecodeJWTs(headers)
		if result == nil {
			t.Error("Expected non-nil map result")
		}

		// If we got a result, verify it's valid
		if jwtInfo, ok := result["Authorization"]; ok {
			// Header and Payload should be maps or nil
			if jwtInfo.Header != nil {
				if _, err := json.Marshal(jwtInfo.Header); err != nil {
					t.Errorf("Header should be marshalable: %v", err)
				}
			}
			if jwtInfo.Payload != nil {
				if _, err := json.Marshal(jwtInfo.Payload); err != nil {
					t.Errorf("Payload should be marshalable: %v", err)
				}
			}
		}
	})
}
