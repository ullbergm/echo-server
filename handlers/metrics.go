package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

// MetricsHandler handles Prometheus metrics requests
func MetricsHandler(c *fiber.Ctx) error {
	handler := fasthttpadaptor.NewFastHTTPHandler(promhttp.Handler())
	handler(c.Context())
	return nil
}
