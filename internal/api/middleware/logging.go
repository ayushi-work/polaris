package middleware

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
)

func Logger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		elapsed := time.Since(start)
		log.Printf("%s %s %d %s", c.Method(), c.Path(), c.Response().StatusCode(), elapsed)
		return err
	}
}
