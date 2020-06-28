package main

import (
	"log"

	"github.com/gofiber/cors"
)

func main() {
	// midlewares
	App.Use(cors.New())

	if err := App.Listen(3000); err != nil {
		log.Fatal(err)
	}
}
