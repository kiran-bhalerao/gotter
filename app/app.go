package app

import (
	"github.com/gofiber/cors"
	"github.com/gofiber/fiber"
)

func SetupApp() *fiber.App {
	app := fiber.New()

	app.Use(cors.New())

	return app
}
