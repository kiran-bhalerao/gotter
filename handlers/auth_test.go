package handlers_test

import (
	"context"
	"net/http/httptest"
	"testing"

	. "github.com/franela/goblin"
	. "github.com/kiranbhalerao123/gotter/app"
	. "github.com/kiranbhalerao123/gotter/config"
	. "github.com/kiranbhalerao123/gotter/router"

	. "github.com/kiranbhalerao123/gotter/handlers/testutils"
	"github.com/stretchr/testify/assert"
)

func TestAuthRoutes(t *testing.T) {
	g := Goblin(t)

	app := SetupApp()
	SetupDB()
	SetupRouter(app)

	g.Describe("Auth Routes Tests", func() {
		g.BeforeEach(func() {
			err := Mongo.DB.Drop(context.Background())

			if err != nil {
				panic(err)
			}
		})

		g.Describe("Signup Route Suits", func() {
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
				resp, inputs, data := TSignup(app)

				g.Assert(resp.StatusCode).Equal(201)
				g.Assert(data.Email).Equal(inputs.Email)
				g.Assert(data.UserName).Equal(inputs.UserName)
			})
		})

		g.Describe("Login Route Suits", func() {
			g.It("returns 400 on invalid request", func() {
				// http.Request
				req := httptest.NewRequest(
					"POST",
					"/api/v1/login",
					nil,
				)

				// Perform the request plain with the app.
				// The -1 disables request latency.
				resp, _ := app.Test(req, -1)
				g.Assert(400).Equal(resp.StatusCode)
			})

			g.It("returns 401 on wrong email or password combination", func() {
				email := "kiran@abc.com"
				password := "kiran123"

				resp, _ := TLogin(app, TLoginInputs{
					Email:    email,
					Password: password,
				})

				g.Assert(resp.StatusCode).Equal(401)
			})

			g.It("returns a token on valid user inputs", func() {
				// first step is to signup the user
				resp, inputs, _ := TSignup(app)
				g.Assert(resp.StatusCode).Equal(201)

				resp, data := TLogin(app, TLoginInputs{
					Email:    inputs.Email,
					Password: inputs.Password,
				})

				g.Assert(resp.StatusCode).Equal(200)
				assert.NotNil(t, data.Data.Token)
			})
		})
	})
}
