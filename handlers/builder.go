package handlers

import (
	"github.com/gofiber/fiber/v2"
)

// BuilderHandler serves the interactive request builder UI
func BuilderHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.Render("builder", nil)
	}
}
