package main

import (
	"github.com/gofiber/fiber"
)

var App *fiber.App

func init() {
	App = fiber.New()
}
