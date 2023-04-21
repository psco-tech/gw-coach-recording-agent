package configserver

import (
	"log"

	"github.com/gofiber/fiber/v2"
)

func Start() {
	app := fiber.New(fiber.Config{})

	// Serve static files like CSS or images
	app.Static("/static", "./static")

	app.Use("/", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.Next()
	})

	app.Get("/", func(c *fiber.Ctx) error {
		return render(c.Response().BodyWriter(), "connect.html", "/", nil)
	})

	app.Get("/devices", func(c *fiber.Ctx) error {
		return render(c.Response().BodyWriter(), "devices.html", "/devices", nil)
	})

	go func() {
		log.Fatal(app.Listen(":3002"))
	}()
}
