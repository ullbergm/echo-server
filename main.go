package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/template/html/v2"
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

	// Start server
	log.Printf("Echo Server starting on port %s", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
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
