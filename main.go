package main

import (
	"log"

	"github.com/gofiber/cors"
	"github.com/gofiber/fiber"
)

var App *fiber.App

func init() {
	App = fiber.New()
}

func main() {
	App.Use(cors.New())

	if err := App.Listen(3000); err != nil {
		log.Fatal(err)
	}
}
