package handlers

import (
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/ullbergm/echo-server/models"
	"github.com/ullbergm/echo-server/services"
)

// Version is injected from main package
var Version string

// TemplateData wraps the response with additional template data
type TemplateData struct {
	models.EchoResponse
	PageTitle string
	Version   string
}

// EchoHandler handles echo requests for all paths and methods
func EchoHandler(jwtService *services.JWTService, bodyService *services.BodyService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Build echo response
		response := buildEchoResponse(c, jwtService, bodyService)

		// Handle response cookies via x-set-cookie header
		setResponseCookies(c)

		// Get custom status code if provided
		statusCode := getCustomStatusCode(c)

		// Content negotiation
		acceptHeader := utils.UnsafeString(c.Request().Header.Peek("Accept"))
		if strings.Contains(acceptHeader, "text/html") {
			// Use Fiber template engine for HTML
			pageTitle := os.Getenv("ECHO_PAGE_TITLE")
			if pageTitle == "" {
				pageTitle = "Echo Server - Request Information"
			}

			templateData := TemplateData{
				EchoResponse: response,
				PageTitle:    pageTitle,
				Version:      Version,
			}
			c.Status(statusCode)
			return c.Render("echo", templateData)
		}

		// Default to JSON
		return c.Status(statusCode).JSON(response)
	}
}

// EchoHandlerHead handles HEAD requests (no body)
func EchoHandlerHead() fiber.Handler {
	return func(c *fiber.Ctx) error {
		statusCode := getCustomStatusCode(c)

		// Set appropriate content type based on Accept header
		acceptHeader := utils.UnsafeString(c.Request().Header.Peek("Accept"))
		if strings.Contains(acceptHeader, "text/html") {
			c.Set("Content-Type", "text/html; charset=utf-8")
		} else {
			c.Set("Content-Type", "application/json")
		}

		return c.SendStatus(statusCode)
	}
}

func buildEchoResponse(c *fiber.Ctx, jwtService *services.JWTService, bodyService *services.BodyService) models.EchoResponse {
	response := models.EchoResponse{
		Request:    buildRequestInfo(c, bodyService),
		Server:     buildServerInfo(),
		Kubernetes: getKubernetesInfo(),
	}

	// Decode JWT tokens if present
	jwtTokens := jwtService.ExtractAndDecodeJWTs(buildHeadersMap(c))
	if len(jwtTokens) > 0 {
		response.JwtTokens = jwtTokens
	}

	return response
}

func buildRequestInfo(c *fiber.Ctx, bodyService *services.BodyService) models.RequestInfo {
	requestInfo := models.RequestInfo{
		Method:        c.Method(),
		Path:          c.Path(),
		Query:         utils.UnsafeString(c.Request().URI().QueryString()),
		Headers:       buildHeadersMap(c),
		RemoteAddress: getRemoteAddress(c),
		Compression:   getCompressionInfo(c),
		Cookies:       parseCookies(c),
	}

	// Parse body for POST, PUT, PATCH, DELETE methods
	method := c.Method()
	if method == "POST" || method == "PUT" || method == "PATCH" || method == "DELETE" {
		bodyBytes := c.Body()
		if len(bodyBytes) > 0 {
			contentType := string(c.Request().Header.ContentType())
			requestInfo.Body = bodyService.ParseBody(bodyBytes, contentType)
		}
	}

	// Add compression info
	requestInfo.Compression = getCompressionInfo(c)

	return requestInfo
}

func buildHeadersMap(c *fiber.Ctx) map[string]string {
	headers := make(map[string]string)
	for key, value := range c.Request().Header.All() {
		headers[string(key)] = string(value)
	}
	return headers
}

func buildServerInfo() models.ServerInfo {
	hostname, _ := os.Hostname()
	hostAddress := getHostAddress()

	info := models.ServerInfo{
		Hostname:    hostname,
		HostAddress: hostAddress,
		Environment: getEnvironmentVariables(),
	}

	return info
}

func getHostAddress() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

func getEnvironmentVariables() map[string]string {
	envVars := make(map[string]string)

	// Check if we're running in Kubernetes
	isKubernetes := os.Getenv("K8S_NAMESPACE") != "" && os.Getenv("K8S_POD_NAME") != ""

	// Determine which environment variables to display
	varsToDisplay := os.Getenv("ECHO_ENVIRONMENT_VARIABLES_DISPLAY")
	if varsToDisplay != "" {
		vars := strings.Split(varsToDisplay, ",")
		for _, v := range vars {
			v = strings.TrimSpace(v)
			if value := os.Getenv(v); value != "" {
				envVars[v] = value
			}
		}
	} else {
		// Default to HOSTNAME
		if hostname := os.Getenv("HOSTNAME"); hostname != "" {
			envVars["HOSTNAME"] = hostname
		}
	}

	// If not in Kubernetes, also include K8S_* variables
	if !isKubernetes {
		for _, env := range os.Environ() {
			pair := strings.SplitN(env, "=", 2)
			if len(pair) == 2 && strings.HasPrefix(pair[0], "K8S_") {
				envVars[pair[0]] = pair[1]
			}
		}
	}

	return envVars
}

func getKubernetesInfo() *models.KubernetesInfo {
	namespace := os.Getenv("K8S_NAMESPACE")
	podName := os.Getenv("K8S_POD_NAME")

	if namespace == "" || podName == "" {
		return nil
	}

	info := &models.KubernetesInfo{
		Namespace:   namespace,
		PodName:     podName,
		PodIP:       os.Getenv("K8S_POD_IP"),
		NodeName:    os.Getenv("K8S_NODE_NAME"),
		ServiceHost: os.Getenv("KUBERNETES_SERVICE_HOST"),
		ServicePort: os.Getenv("KUBERNETES_SERVICE_PORT"),
	}

	// Collect labels
	labels := make(map[string]string)
	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		if len(pair) == 2 && strings.HasPrefix(pair[0], "K8S_LABEL_") {
			labelName := strings.TrimPrefix(pair[0], "K8S_LABEL_")
			labels[labelName] = pair[1]
		}
	}
	if len(labels) > 0 {
		info.Labels = labels
	}

	// Collect annotations
	annotations := make(map[string]string)
	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		if len(pair) == 2 && strings.HasPrefix(pair[0], "K8S_ANNOTATION_") {
			annotationName := strings.TrimPrefix(pair[0], "K8S_ANNOTATION_")
			annotations[annotationName] = pair[1]
		}
	}
	if len(annotations) > 0 {
		info.Annotations = annotations
	}

	return info
}

func getRemoteAddress(c *fiber.Ctx) string {
	// Check X-Forwarded-For header
	if xff := c.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}

	// Check X-Real-IP header
	if xri := c.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Use remote address from context
	return c.IP()
}

func getCustomStatusCode(c *fiber.Ctx) int {
	statusHeader := c.Get("x-set-response-status-code")
	if statusHeader == "" {
		return fiber.StatusOK
	}

	statusCode, err := strconv.Atoi(statusHeader)
	if err != nil || statusCode < 200 || statusCode > 599 {
		return fiber.StatusOK
	}

	return statusCode
}

// parseCookies extracts all cookies from the request
func parseCookies(c *fiber.Ctx) []models.CookieInfo {
	cookies := []models.CookieInfo{}

	// Get all cookies from the request
	for key, value := range c.Request().Header.Cookies() {
		cookie := models.CookieInfo{
			Name:  string(key),
			Value: string(value),
		}

		cookies = append(cookies, cookie)
	}

	return cookies
}

// setResponseCookies sets cookies in the response based on x-set-cookie header
func setResponseCookies(c *fiber.Ctx) {
	// Check for x-set-cookie header
	setCookieHeader := c.Get("x-set-cookie")
	if setCookieHeader == "" {
		return
	}

	// Parse the x-set-cookie header value
	// Format: name=value; Domain=example.com; Path=/; Expires=...; HttpOnly; Secure; SameSite=Strict
	cookie := parseSetCookieHeader(setCookieHeader)
	if cookie != nil {
		c.Cookie(cookie)
	}
}

// parseSetCookieHeader parses the x-set-cookie header value and returns a Fiber cookie
func parseSetCookieHeader(headerValue string) *fiber.Cookie {
	parts := strings.Split(headerValue, ";")

	// Parse name=value
	nameValue := strings.TrimSpace(parts[0])
	if nameValue == "" {
		return nil
	}
	nvParts := strings.SplitN(nameValue, "=", 2)
	if len(nvParts) != 2 {
		return nil
	}

	cookie := &fiber.Cookie{
		Name:  strings.TrimSpace(nvParts[0]),
		Value: strings.TrimSpace(nvParts[1]),
	}

	// Parse attributes
	for i := 1; i < len(parts); i++ {
		attr := strings.TrimSpace(parts[i])
		attrParts := strings.SplitN(attr, "=", 2)
		attrName := strings.ToLower(strings.TrimSpace(attrParts[0]))

		switch attrName {
		case "domain":
			if len(attrParts) == 2 {
				cookie.Domain = strings.TrimSpace(attrParts[1])
			}
		case "path":
			if len(attrParts) == 2 {
				cookie.Path = strings.TrimSpace(attrParts[1])
			}
		case "expires":
			if len(attrParts) == 2 {
				// Parse the expires date
				expiresStr := strings.TrimSpace(attrParts[1])
				// Try parsing common date formats
				if t, err := parseExpires(expiresStr); err == nil {
					cookie.Expires = t
				}
			}
		case "max-age":
			if len(attrParts) == 2 {
				if maxAge, err := strconv.Atoi(strings.TrimSpace(attrParts[1])); err == nil {
					cookie.MaxAge = maxAge
				}
			}
		case "httponly":
			cookie.HTTPOnly = true
		case "secure":
			cookie.Secure = true
		case "samesite":
			if len(attrParts) == 2 {
				sameSite := strings.ToLower(strings.TrimSpace(attrParts[1]))
				switch sameSite {
				case "strict":
					cookie.SameSite = "Strict"
				case "lax":
					cookie.SameSite = "Lax"
				case "none":
					cookie.SameSite = "None"
				}
			}
		}
	}

	return cookie
}

// parseExpires tries to parse various date formats for cookie expiry
func parseExpires(dateStr string) (time.Time, error) {
	// RFC 1123 format (standard for HTTP dates)
	if t, err := time.Parse(time.RFC1123, dateStr); err == nil {
		return t, nil
	}

	// RFC 850 format
	if t, err := time.Parse(time.RFC850, dateStr); err == nil {
		return t, nil
	}

	// ANSI C asctime() format
	if t, err := time.Parse(time.ANSIC, dateStr); err == nil {
		return t, nil
	}

	// ISO 8601 format
	if t, err := time.Parse(time.RFC3339, dateStr); err == nil {
		return t, nil
	}

	return time.Time{}, strconv.ErrSyntax
}

func getCompressionInfo(c *fiber.Ctx) *models.CompressionInfo {
	acceptEncoding := c.Get("Accept-Encoding")

	if acceptEncoding == "" {
		return &models.CompressionInfo{
			Supported: false,
		}
	}

	// Parse accepted encodings
	encodings := []string{}
	parts := strings.Split(acceptEncoding, ",")
	for _, part := range parts {
		encoding := strings.TrimSpace(strings.Split(part, ";")[0])
		if encoding != "" {
			encodings = append(encodings, encoding)
		}
	}

	// Check what encoding will be used in the response
	responseEncoding := ""
	contentEncoding := string(c.Response().Header.Peek("Content-Encoding"))
	if contentEncoding != "" {
		responseEncoding = contentEncoding
	}

	return &models.CompressionInfo{
		AcceptedEncodings: encodings,
		ResponseEncoding:  responseEncoding,
		Supported:         len(encodings) > 0,
	}
}
