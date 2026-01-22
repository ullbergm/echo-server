package services

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"mime"
	"mime/multipart"
	"net/url"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/ullbergm/echo-server/models"
)

const (
	// DefaultMaxBodySize is the default maximum body size (10MB)
	DefaultMaxBodySize = 10 * 1024 * 1024

	// Binary data detection constants
	binaryThreshold    = 0.3  // 30% non-printable chars threshold
	minPrintableChar   = 32   // ASCII space character
	charNewline        = '\n' // Newline character
	charCarriageReturn = '\r' // Carriage return character
	charTab            = '\t' // Tab character
)

// BodyService handles request body parsing
type BodyService struct {
	maxBodySize int
}

// NewBodyService creates a new body service
func NewBodyService() *BodyService {
	maxSize := DefaultMaxBodySize

	// Check for environment variable override
	if maxSizeEnv := os.Getenv("MAX_BODY_SIZE"); maxSizeEnv != "" {
		if parsed, err := strconv.Atoi(maxSizeEnv); err == nil && parsed > 0 {
			maxSize = parsed
		}
	}

	return &BodyService{
		maxBodySize: maxSize,
	}
}

// ParseBody parses the request body based on Content-Type
func (s *BodyService) ParseBody(bodyBytes []byte, contentType string) *models.BodyInfo {
	if len(bodyBytes) == 0 {
		return nil
	}

	bodyInfo := &models.BodyInfo{
		ContentType: contentType,
		Size:        len(bodyBytes),
	}

	// Check if body size exceeds limit
	if len(bodyBytes) > s.maxBodySize {
		bodyInfo.Truncated = true
		bodyBytes = bodyBytes[:s.maxBodySize]
	}

	// Parse media type
	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		mediaType = contentType
	}

	// Check if binary data
	if s.isBinaryData(bodyBytes) {
		bodyInfo.IsBinary = true
		bodyInfo.Content = base64.StdEncoding.EncodeToString(bodyBytes)
		return bodyInfo
	}

	// Parse based on content type
	switch {
	case strings.HasPrefix(mediaType, "application/json"):
		bodyInfo.Content = s.parseJSON(bodyBytes)
	case strings.HasPrefix(mediaType, "application/xml") || strings.HasPrefix(mediaType, "text/xml"):
		bodyInfo.Content = s.parseXML(bodyBytes)
	case strings.HasPrefix(mediaType, "application/x-www-form-urlencoded"):
		bodyInfo.Content = s.parseFormURLEncoded(bodyBytes)
	case strings.HasPrefix(mediaType, "multipart/form-data"):
		bodyInfo.Content = s.parseMultipartForm(bodyBytes, params)
	case strings.HasPrefix(mediaType, "text/"):
		bodyInfo.Content = string(bodyBytes)
	default:
		// Try to parse as text if it's valid UTF-8
		if utf8.Valid(bodyBytes) {
			bodyInfo.Content = string(bodyBytes)
		} else {
			bodyInfo.IsBinary = true
			bodyInfo.Content = base64.StdEncoding.EncodeToString(bodyBytes)
		}
	}

	return bodyInfo
}

// isBinaryData checks if the data is binary
func (s *BodyService) isBinaryData(data []byte) bool {
	// Check for null bytes (common in binary data)
	if bytes.IndexByte(data, 0) != -1 {
		return true
	}

	// Check if data is valid UTF-8
	if !utf8.Valid(data) {
		return true
	}

	// Check for high ratio of non-printable characters
	nonPrintable := 0
	for _, b := range data {
		if b < minPrintableChar && b != charNewline && b != charCarriageReturn && b != charTab {
			nonPrintable++
		}
	}

	// If more than threshold non-printable, consider it binary
	return float64(nonPrintable)/float64(len(data)) > binaryThreshold
}

// parseJSON parses JSON body
func (s *BodyService) parseJSON(data []byte) interface{} {
	var result interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		// If JSON parsing fails, return as string
		return string(data)
	}
	return result
}

// parseXML parses XML body
// Note: Go's encoding/xml doesn't support unmarshaling to map[string]interface{}
// so we return the XML as a string for display purposes
func (s *BodyService) parseXML(data []byte) interface{} {
	// Return XML as string - Go's xml.Unmarshal requires struct definitions
	// for proper parsing, which we don't have for arbitrary XML
	return string(data)
}

// parseFormURLEncoded parses URL-encoded form data
func (s *BodyService) parseFormURLEncoded(data []byte) interface{} {
	values, err := url.ParseQuery(string(data))
	if err != nil {
		return string(data)
	}

	// Convert to map[string]interface{} for consistent JSON output
	result := make(map[string]interface{})
	for key, vals := range values {
		if len(vals) == 1 {
			result[key] = vals[0]
		} else {
			result[key] = vals
		}
	}
	return result
}

// parseMultipartForm parses multipart form data
func (s *BodyService) parseMultipartForm(data []byte, params map[string]string) interface{} {
	boundary, ok := params["boundary"]
	if !ok {
		return string(data)
	}

	reader := multipart.NewReader(bytes.NewReader(data), boundary)
	result := make(map[string]interface{})

	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return string(data)
		}

		partData, err := io.ReadAll(part)
		if err != nil {
			continue
		}

		fieldName := part.FormName()
		if fieldName == "" {
			continue
		}

		// Check if it's a file upload
		if part.FileName() != "" {
			fileInfo := map[string]interface{}{
				"filename": part.FileName(),
				"size":     len(partData),
			}

			// For binary files, use base64
			if s.isBinaryData(partData) {
				fileInfo["content"] = base64.StdEncoding.EncodeToString(partData)
				fileInfo["encoding"] = "base64"
			} else {
				fileInfo["content"] = string(partData)
			}

			result[fieldName] = fileInfo
		} else {
			// Regular form field
			result[fieldName] = string(partData)
		}
	}

	return result
}
