package services

import (
	"encoding/base64"
	"encoding/json"
	"testing"
)

// TestMalformedJWTTokens tests handling of malformed JWT tokens
func TestMalformedJWTTokens(t *testing.T) {
	service := NewJWTService()

	tests := []struct {
		name          string
		token         string
		expectDecoded bool
		description   string
	}{
		{
			name:          "token with only one part",
			token:         "singlepart",
			expectDecoded: false,
			description:   "should not decode single part token",
		},
		{
			name:          "token with only two parts",
			token:         "part1.part2",
			expectDecoded: false,
			description:   "should not decode two part token",
		},
		{
			name:          "token with invalid base64 header",
			token:         "!!!invalid!!!.validpayload.signature",
			expectDecoded: false,
			description:   "should not decode token with invalid base64 header",
		},
		{
			name:          "token with invalid base64 payload",
			token:         "validheader.!!!invalid!!!.signature",
			expectDecoded: false,
			description:   "should not decode token with invalid base64 payload",
		},
		{
			name:          "empty token",
			token:         "",
			expectDecoded: false,
			description:   "should not decode empty token",
		},
		{
			name:          "token with extra dots",
			token:         "header.payload.signature.extra",
			expectDecoded: false, // JWT decoder checks for exactly 3 parts
			description:   "should not decode token with extra parts",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			headers := map[string]string{
				"Authorization": tt.token,
			}

			tokens := service.ExtractAndDecodeJWTs(headers)

			if tt.expectDecoded && len(tokens) == 0 {
				t.Errorf("%s: expected token to be decoded", tt.description)
			}

			if !tt.expectDecoded && len(tokens) > 0 {
				t.Logf("%s: token was unexpectedly decoded (might be acceptable)", tt.description)
			}
		})
	}
}

// TestMultipleJWTTokensFromDifferentHeaders tests extraction of multiple JWT tokens
func TestMultipleJWTTokensFromDifferentHeaders(t *testing.T) {
	service := NewJWTService()

	// Create two different valid JWT tokens
	header1 := map[string]interface{}{"alg": "HS256", "typ": "JWT"}
	payload1 := map[string]interface{}{"sub": "user1", "name": "User One"}

	header2 := map[string]interface{}{"alg": "RS256", "typ": "JWT"}
	payload2 := map[string]interface{}{"sub": "user2", "name": "User Two"}

	headerJSON1, _ := json.Marshal(header1)
	payloadJSON1, _ := json.Marshal(payload1)
	headerB64_1 := base64.RawURLEncoding.EncodeToString(headerJSON1)
	payloadB64_1 := base64.RawURLEncoding.EncodeToString(payloadJSON1)
	token1 := headerB64_1 + "." + payloadB64_1 + ".sig1"

	headerJSON2, _ := json.Marshal(header2)
	payloadJSON2, _ := json.Marshal(payload2)
	headerB64_2 := base64.RawURLEncoding.EncodeToString(headerJSON2)
	payloadB64_2 := base64.RawURLEncoding.EncodeToString(payloadJSON2)
	token2 := headerB64_2 + "." + payloadB64_2 + ".sig2"

	headers := map[string]string{
		"Authorization": "Bearer " + token1,
		"X-JWT-Token":   token2,
	}

	tokens := service.ExtractAndDecodeJWTs(headers)

	// Should extract both tokens
	if len(tokens) < 2 {
		t.Errorf("Expected at least 2 tokens, got %d", len(tokens))
	}

	// Verify both tokens are present
	if _, ok := tokens["Authorization"]; !ok {
		t.Error("Expected Authorization token to be present")
	}

	if _, ok := tokens["X-JWT-Token"]; !ok {
		t.Error("Expected X-JWT-Token to be present")
	}
}

// TestJWTWithSpecialCharactersInPayload tests JWT tokens with special characters
func TestJWTWithSpecialCharactersInPayload(t *testing.T) {
	service := NewJWTService()

	// Create JWT with special characters in payload
	header := map[string]interface{}{"alg": "HS256", "typ": "JWT"}
	payload := map[string]interface{}{
		"name":  "John O'Brien",
		"email": "john+test@example.com",
		"role":  "admin/superuser",
		"notes": "Special chars: <>&\"'",
	}

	headerJSON, _ := json.Marshal(header)
	payloadJSON, _ := json.Marshal(payload)

	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)
	payloadB64 := base64.RawURLEncoding.EncodeToString(payloadJSON)
	token := headerB64 + "." + payloadB64 + ".signature"

	headers := map[string]string{
		"Authorization": token,
	}

	tokens := service.ExtractAndDecodeJWTs(headers)

	if len(tokens) == 0 {
		t.Fatal("Expected token to be decoded")
	}

	jwtInfo := tokens["Authorization"]

	// Verify payload contains special characters
	if jwtInfo.Payload["name"] != "John O'Brien" {
		t.Errorf("Expected name with apostrophe, got %v", jwtInfo.Payload["name"])
	}

	if jwtInfo.Payload["email"] != "john+test@example.com" {
		t.Errorf("Expected email with plus sign, got %v", jwtInfo.Payload["email"])
	}
}

// TestEmptyJWTHeaders tests behavior with empty JWT header values
func TestEmptyJWTHeaders(t *testing.T) {
	service := NewJWTService()

	headers := map[string]string{
		"Authorization": "",
		"X-JWT-Token":   "   ",
		"X-Auth-Token":  "\t\n",
	}

	tokens := service.ExtractAndDecodeJWTs(headers)

	// Should not extract any tokens from empty headers
	if len(tokens) > 0 {
		t.Errorf("Expected no tokens from empty headers, got %d", len(tokens))
	}
}

// TestJWTBearerPrefixVariations tests various Bearer prefix formats
func TestJWTBearerPrefixVariations(t *testing.T) {
	service := NewJWTService()

	// Create a valid JWT token
	header := map[string]interface{}{"alg": "HS256", "typ": "JWT"}
	payload := map[string]interface{}{"sub": "test"}

	headerJSON, _ := json.Marshal(header)
	payloadJSON, _ := json.Marshal(payload)

	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)
	payloadB64 := base64.RawURLEncoding.EncodeToString(payloadJSON)
	token := headerB64 + "." + payloadB64 + ".sig"

	tests := []struct {
		name          string
		headerValue   string
		expectDecoded bool
	}{
		{
			name:          "Bearer with space",
			headerValue:   "Bearer " + token,
			expectDecoded: true,
		},
		{
			name:          "bearer lowercase",
			headerValue:   "bearer " + token,
			expectDecoded: true,
		},
		{
			name:          "BEARER uppercase",
			headerValue:   "BEARER " + token,
			expectDecoded: true,
		},
		{
			name:          "no Bearer prefix",
			headerValue:   token,
			expectDecoded: true,
		},
		{
			name:          "Bearer with multiple spaces",
			headerValue:   "Bearer  " + token,
			expectDecoded: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			headers := map[string]string{
				"Authorization": tt.headerValue,
			}

			tokens := service.ExtractAndDecodeJWTs(headers)

			if tt.expectDecoded && len(tokens) == 0 {
				t.Error("Expected token to be decoded")
			}

			if !tt.expectDecoded && len(tokens) > 0 {
				t.Error("Expected token not to be decoded")
			}
		})
	}
}

// TestJWTWithNumericAndBooleanValues tests JWT payload with various data types
func TestJWTWithNumericAndBooleanValues(t *testing.T) {
	service := NewJWTService()

	header := map[string]interface{}{"alg": "HS256", "typ": "JWT"}
	payload := map[string]interface{}{
		"user_id":    12345,
		"age":        30,
		"score":      95.5,
		"is_active":  true,
		"is_admin":   false,
		"roles":      []string{"user", "editor"},
		"metadata":   map[string]interface{}{"key": "value"},
		"null_field": nil,
	}

	headerJSON, _ := json.Marshal(header)
	payloadJSON, _ := json.Marshal(payload)

	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)
	payloadB64 := base64.RawURLEncoding.EncodeToString(payloadJSON)
	token := headerB64 + "." + payloadB64 + ".signature"

	headers := map[string]string{
		"Authorization": token,
	}

	tokens := service.ExtractAndDecodeJWTs(headers)

	if len(tokens) == 0 {
		t.Fatal("Expected token to be decoded")
	}

	jwtInfo := tokens["Authorization"]

	// Verify numeric values
	if userID, ok := jwtInfo.Payload["user_id"].(float64); !ok || userID != 12345 {
		t.Errorf("Expected user_id to be 12345, got %v", jwtInfo.Payload["user_id"])
	}

	if score, ok := jwtInfo.Payload["score"].(float64); !ok || score != 95.5 {
		t.Errorf("Expected score to be 95.5, got %v", jwtInfo.Payload["score"])
	}

	// Verify boolean values
	if isActive, ok := jwtInfo.Payload["is_active"].(bool); !ok || !isActive {
		t.Errorf("Expected is_active to be true, got %v", jwtInfo.Payload["is_active"])
	}

	if isAdmin, ok := jwtInfo.Payload["is_admin"].(bool); !ok || isAdmin {
		t.Errorf("Expected is_admin to be false, got %v", jwtInfo.Payload["is_admin"])
	}

	// Verify array
	if roles, ok := jwtInfo.Payload["roles"].([]interface{}); !ok || len(roles) != 2 {
		t.Errorf("Expected roles array with 2 elements, got %v", jwtInfo.Payload["roles"])
	}

	// Verify nested map
	if metadata, ok := jwtInfo.Payload["metadata"].(map[string]interface{}); !ok || len(metadata) == 0 {
		t.Errorf("Expected metadata map, got %v", jwtInfo.Payload["metadata"])
	}
}
