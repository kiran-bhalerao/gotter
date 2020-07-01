package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http/httptest"
	"testing"

	. "github.com/franela/goblin"
	. "github.com/kiranbhalerao123/gotter/app"
	. "github.com/kiranbhalerao123/gotter/config"
	. "github.com/kiranbhalerao123/gotter/router"
)

func TestAuthRoutes(t *testing.T) {
	g := Goblin(t)

	app := SetupApp()
	SetupDB()
	SetupRouter(app)

	g.Describe("Signup Route Suits", func() {
		g.BeforeEach(func() {
			err := Mongo.DB.Drop(context.Background())

			if err != nil {
				panic(err)
			}
		})

		g.It("returns 400 on invalid request", func() {
			// http.Request
			req := httptest.NewRequest(
				"POST",
				"/api/v1/signup",
				nil,
			)

			// Perform the request plain with the app.
			// The -1 disables request latency.
			resp, _ := app.Test(req, -1)
			g.Assert(400).Equal(resp.StatusCode)
		})

		g.It("returns 201 on valid email, username and password", func() {
			type SignInputs struct {
				Email    string `json:"email"`
				UserName string `json:"username"`
				Password string `json:"password"`
			}

			email := "kiran@abc.com"
			username := "kiran"
			password := "kiran123"

			body := SignInputs{
				Email:    email,
				Password: password,
				UserName: username,
			}

			buf := new(bytes.Buffer)
			err := json.NewEncoder(buf).Encode(&body)

			if err != nil {
				log.Fatal(err)
			}

			// http.Request
			req := httptest.NewRequest(
				"POST",
				"/api/v1/signup",
				buf,
			)
			req.Header.Set("Content-Type", "application/json")

			resp, _ := app.Test(req, -1)

			// Print the body to the stdout
			// io.Copy(os.Stdout, resp.Body)

			g.Assert(201).Equal(resp.StatusCode)
		})
	})
}
