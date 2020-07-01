package main

import (
	"log"

	. "github.com/kiranbhalerao123/gotter/app"
	. "github.com/kiranbhalerao123/gotter/config"
	. "github.com/kiranbhalerao123/gotter/router"
)

func main() {
	app := SetupApp()
	SetupDB()
	SetupRouter(app)

	if err := app.Listen(3000); err != nil {
		log.Fatal(err)
	}
}
