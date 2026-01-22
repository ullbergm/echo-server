package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/template/html/v3"
	"github.com/ullbergm/echo-server/handlers"
	"github.com/ullbergm/echo-server/services"
)

// Version is set via ldflags at build time
// Build with: go build -ldflags "-X main.Version=$(git describe --tags --abbrev=0)"
var Version = "dev"

var startTime = time.Now()

func main() {
	// Initialize template engine with custom functions
	engine := html.New("./templates", ".html")

	// Add custom template functions
	engine.AddFunc("FormatJWTValue", func(key string, value interface{}) string {
		displayValue := formatValue(value)

		// Format timestamps (exp, iat, nbf)
		if key == "exp" || key == "iat" || key == "nbf" {
			if timestamp, ok := value.(float64); ok {
				t := time.Unix(int64(timestamp), 0).UTC()
				displayValue = strconv.FormatInt(int64(timestamp), 10) + " (" + t.Format("Mon Jan 02 15:04:05 UTC 2006") + ")"
			}
		}

		return displayValue
	})

	engine.AddFunc("FormatBodyContent", func(content interface{}) string {
		switch v := content.(type) {
		case string:
			return v
		case map[string]interface{}, []interface{}:
			// Pretty print JSON
			jsonBytes, err := json.MarshalIndent(v, "", "  ")
			if err != nil {
				return fmt.Sprintf("%v", v)
			}
			return string(jsonBytes)
		default:
			return fmt.Sprintf("%v", v)
		}
	})

	// Create Fiber app with template engine
	// Get prefork setting from environment (default: false)
	prefork := false
	if preforkEnv := os.Getenv("FIBER_PREFORK"); preforkEnv != "" {
		if parsed, err := strconv.ParseBool(preforkEnv); err == nil {
			prefork = parsed
		}
	}

	// Create Fiber app with template engine and performance optimizations
	app := fiber.New(fiber.Config{
		AppName:               "Echo Server v" + Version,
		DisableStartupMessage: false,
		Views:                 engine,
		// Use goccy/go-json for faster JSON encoding/decoding
		JSONEncoder: json.Marshal,
		JSONDecoder: json.Unmarshal,
		// Enable prefork for multi-core scalability (optional, configurable via FIBER_PREFORK env var)
		Prefork: prefork,
	})

	// Middleware
	app.Use(recover.New())

	// Compression middleware
	app.Use(compress.New(compress.Config{
		Level: compress.LevelDefault, // Default compression level
		Next: func(c *fiber.Ctx) bool {
			// Skip compression for healthcheck endpoints
			path := c.Path()
			return path == "/healthz/live" || path == "/healthz/ready"
		},
	}))

	// Favicon middleware
	app.Use(favicon.New(favicon.Config{
		File: "./public/favicon.ico",
		URL:  "/favicon.ico",
	}))

	// Get log healthchecks setting from environment (default: false)
	logHealthchecks := false
	if logHealthchecksEnv := os.Getenv("LOG_HEALTHCHECKS"); logHealthchecksEnv != "" {
		if parsed, err := strconv.ParseBool(logHealthchecksEnv); err == nil {
			logHealthchecks = parsed
		}
	}

	// Comprehensive logging middleware with optional healthcheck filtering
	app.Use(logger.New(logger.Config{
		Format:     "[${time}] ${status} - ${latency} ${method} ${path} - IP: ${ip} - UA: ${ua}\n",
		TimeFormat: "2006-01-02 15:04:05",
		TimeZone:   "UTC",
		Next: func(c *fiber.Ctx) bool {
			// Skip logging healthcheck and monitor requests unless explicitly enabled
			if !logHealthchecks {
				path := c.Path()
				return path == "/healthz/live" || path == "/healthz/ready" || path == "/monitor"
			}
			return false
		},
	}))

	// Get readiness delay from environment
	delaySeconds := 0
	if delayEnv := os.Getenv("HEALTH_READINESS_DELAY_SECONDS"); delayEnv != "" {
		if parsed, err := strconv.Atoi(delayEnv); err == nil {
			delaySeconds = parsed
		}
	}

	// Health check middleware using Fiber's built-in middleware
	app.Use(healthcheck.New(healthcheck.Config{
		LivenessProbe: func(c *fiber.Ctx) bool {
			// Always healthy if the server is running
			return true
		},
		LivenessEndpoint: "/healthz/live",
		ReadinessProbe: func(c *fiber.Ctx) bool {
			// Check if startup delay has passed
			if delaySeconds > 0 {
				elapsed := time.Since(startTime)
				required := time.Duration(delaySeconds) * time.Second
				return elapsed >= required
			}
			return true
		},
		ReadinessEndpoint: "/healthz/ready",
	}))

	// Initialize services
	jwtService := services.NewJWTService()
	bodyService := services.NewBodyService()
	metricsService := services.NewMetricsService()

	// Set version in handlers package
	handlers.Version = Version

	// Metrics middleware
	app.Use(func(c *fiber.Ctx) error {
		return metricsService.MetricsMiddleware(c)
	})

	// Monitor dashboard endpoint
	app.Get("/monitor", monitor.New(monitor.Config{
		Title:   "Echo Server Monitor",
		Refresh: 3 * time.Second,
	}))

	// Prometheus metrics endpoint
	app.Get("/metrics", handlers.MetricsHandler)

	// Request builder UI endpoint
	app.Get("/builder", handlers.BuilderHandler())

	// Echo handlers for all HTTP methods (wildcard path)
	app.Get("/*", handlers.EchoHandler(jwtService, bodyService))
	app.Post("/*", handlers.EchoHandler(jwtService, bodyService))
	app.Put("/*", handlers.EchoHandler(jwtService, bodyService))
	app.Patch("/*", handlers.EchoHandler(jwtService, bodyService))
	app.Delete("/*", handlers.EchoHandler(jwtService, bodyService))
	app.Options("/*", handlers.EchoHandler(jwtService, bodyService))
	app.Head("/*", handlers.EchoHandlerHead())

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Check if TLS is enabled
	tlsEnabled := false
	if tlsEnv := os.Getenv("TLS_ENABLED"); tlsEnv != "" {
		if parsed, err := strconv.ParseBool(tlsEnv); err == nil {
			tlsEnabled = parsed
		}
	}

	if tlsEnabled {
		// TLS is enabled, start both HTTP and HTTPS servers
		startDualStackServers(app, port)
	} else {
		// TLS is disabled, start only HTTP server
		log.Printf("Echo Server starting on port %s (HTTP only)", port)
		if err := app.Listen(":" + port); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}
}

func formatValue(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	case int:
		return strconv.Itoa(val)
	case int64:
		return strconv.FormatInt(val, 10)
	case bool:
		return strconv.FormatBool(val)
	default:
		return fmt.Sprintf("%v", val)
	}
}

// startDualStackServers starts both HTTP and HTTPS servers
func startDualStackServers(app *fiber.App, httpPort string) {
	// Get TLS configuration
	tlsPort := os.Getenv("TLS_PORT")
	if tlsPort == "" {
		tlsPort = "8443"
	}

	certFile := os.Getenv("TLS_CERT_FILE")
	if certFile == "" {
		certFile = "/certs/tls.crt"
	}

	keyFile := os.Getenv("TLS_KEY_FILE")
	if keyFile == "" {
		keyFile = "/certs/tls.key"
	}

	// Initialize TLS service
	tlsService := services.NewTLSService()

	// Load or generate certificate
	cert, err := tlsService.GetOrGenerateCertificate(certFile, keyFile)
	if err != nil {
		log.Fatalf("Failed to get TLS certificate: %v", err)
	}

	// Store certificate information in environment for handlers to access
	storeCertificateInfo(&cert)

	// Create TLS config
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}

	// Use WaitGroup to manage both servers
	var wg sync.WaitGroup
	wg.Add(2)

	// Start HTTP server in goroutine
	go func() {
		defer wg.Done()
		log.Printf("Echo Server starting HTTP server on port %s", httpPort)
		if listenErr := app.Listen(":" + httpPort); listenErr != nil {
			log.Printf("HTTP server error: %v", listenErr)
		}
	}()

	// Start HTTPS server in goroutine
	go func() {
		defer wg.Done()
		log.Printf("Echo Server starting HTTPS server on port %s", tlsPort)

		// Use ListenTLSWithCertificate for custom TLS config
		ln, listenErr := tls.Listen("tcp", ":"+tlsPort, tlsConfig)
		if listenErr != nil {
			log.Printf("Failed to create TLS listener: %v", listenErr)
			return
		}

		if listenerErr := app.Listener(ln); listenerErr != nil {
			log.Printf("HTTPS server error: %v", listenerErr)
		}
	}()

	log.Printf("Dual-stack servers running - HTTP:%s HTTPS:%s", httpPort, tlsPort)
	wg.Wait()
}

// storeCertificateInfo stores certificate information in environment variables
// so handlers can access it. This follows the existing pattern in the codebase
// of using environment variables for configuration and metadata (see K8S_* vars).
// Environment variables are used here for simplicity and consistency with the
// existing architecture, where handlers access server metadata via os.Getenv().
func storeCertificateInfo(cert *tls.Certificate) {
	x509Cert, err := services.ParseCertificate(cert)
	if err != nil {
		log.Printf("Warning: Failed to parse certificate: %v", err)
		return
	}

	// Store certificate info in environment with _ prefix to indicate internal use
	if setErr := os.Setenv("_TLS_CERT_SUBJECT", x509Cert.Subject.String()); setErr != nil {
		log.Printf("Warning: Failed to set _TLS_CERT_SUBJECT: %v", setErr)
	}
	if setErr := os.Setenv("_TLS_CERT_ISSUER", x509Cert.Issuer.String()); setErr != nil {
		log.Printf("Warning: Failed to set _TLS_CERT_ISSUER: %v", setErr)
	}
	if setErr := os.Setenv("_TLS_CERT_NOT_BEFORE", x509Cert.NotBefore.Format(time.RFC3339)); setErr != nil {
		log.Printf("Warning: Failed to set _TLS_CERT_NOT_BEFORE: %v", setErr)
	}
	if setErr := os.Setenv("_TLS_CERT_NOT_AFTER", x509Cert.NotAfter.Format(time.RFC3339)); setErr != nil {
		log.Printf("Warning: Failed to set _TLS_CERT_NOT_AFTER: %v", setErr)
	}
	if setErr := os.Setenv("_TLS_CERT_SERIAL", x509Cert.SerialNumber.String()); setErr != nil {
		log.Printf("Warning: Failed to set _TLS_CERT_SERIAL: %v", setErr)
	}
	if len(x509Cert.DNSNames) > 0 {
		if setErr := os.Setenv("_TLS_CERT_DNS_NAMES", strings.Join(x509Cert.DNSNames, ",")); setErr != nil {
			log.Printf("Warning: Failed to set _TLS_CERT_DNS_NAMES: %v", setErr)
		}
	}
}
