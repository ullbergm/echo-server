package services

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"strings"

	"github.com/ullbergm/echo-server/models"
)

// JWTService handles JWT token decoding
type JWTService struct {
	headerNames []string
}

// NewJWTService creates a new JWT service
func NewJWTService() *JWTService {
	headerNames := os.Getenv("JWT_HEADER_NAMES")
	if headerNames == "" {
		headerNames = "Authorization,X-JWT-Token,X-Auth-Token,JWT-Token"
	}

	names := strings.Split(headerNames, ",")
	for i, name := range names {
		names[i] = strings.TrimSpace(name)
	}

	return &JWTService{
		headerNames: names,
	}
}

// ExtractAndDecodeJWTs extracts and decodes JWT tokens from request headers
func (s *JWTService) ExtractAndDecodeJWTs(headers map[string]string) map[string]models.JwtInfo {
	jwtInfoMap := make(map[string]models.JwtInfo)

	if len(headers) == 0 {
		return jwtInfoMap
	}

	// Check all configured JWT header names
	for _, headerName := range s.headerNames {
		token := headers[headerName]
		if token == "" {
			// Try lowercase
			token = headers[strings.ToLower(headerName)]
		}

		if token != "" {
			// Strip "Bearer " prefix if present
			if strings.HasPrefix(strings.ToLower(token), "bearer ") {
				token = strings.TrimSpace(token[7:])
			}

			if jwtInfo := s.decodeJWT(token); jwtInfo != nil {
				jwtInfoMap[headerName] = *jwtInfo
			}
		}
	}

	return jwtInfoMap
}

// decodeJWT decodes a JWT token and extracts header and payload
func (s *JWTService) decodeJWT(token string) *models.JwtInfo {
	if token == "" {
		return nil
	}

	// JWT format: header.payload.signature
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil
	}

	jwtInfo := &models.JwtInfo{
		RawToken: s.truncateToken(token),
	}

	// Decode header (first part)
	if header, err := s.decodeBase64URL(parts[0]); err == nil {
		jwtInfo.Header = header
	} else {
		return nil
	}

	// Decode payload (second part)
	if payload, err := s.decodeBase64URL(parts[1]); err == nil {
		jwtInfo.Payload = payload
	} else {
		return nil
	}

	return jwtInfo
}

// decodeBase64URL decodes a Base64URL encoded string
func (s *JWTService) decodeBase64URL(input string) (map[string]interface{}, error) {
	// Base64URL uses - and _ instead of + and /
	input = strings.ReplaceAll(input, "-", "+")
	input = strings.ReplaceAll(input, "_", "/")

	// Add padding if needed
	padding := (4 - len(input)%4) % 4
	input += strings.Repeat("=", padding)

	// Decode base64
	decoded, err := base64.StdEncoding.DecodeString(input)
	if err != nil {
		return nil, err
	}

	// Parse JSON
	var result map[string]interface{}
	if unmarshalErr := json.Unmarshal(decoded, &result); unmarshalErr != nil {
		return nil, unmarshalErr
	}

	return result, nil
}

// truncateToken truncates token for display purposes
func (s *JWTService) truncateToken(token string) string {
	if len(token) <= 30 {
		return token
	}
	return token[:10] + "..." + token[len(token)-10:]
}
