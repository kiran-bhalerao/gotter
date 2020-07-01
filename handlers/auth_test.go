package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/franela/goblin"
	"github.com/gofiber/fiber"
	. "github.com/kiranbhalerao123/gotter/app"
	. "github.com/kiranbhalerao123/gotter/config"
	. "github.com/kiranbhalerao123/gotter/router"
	"github.com/stretchr/testify/assert"
)

type TSignupOutput struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	UserName string `json:"username"`
}

type TSignInputs struct {
	Email    string `json:"email"`
	UserName string `json:"username"`
	Password string `json:"password"`
}

func TSignup(app *fiber.App) (resp *http.Response, inputs TSignInputs, data TSignupOutput) {
	email := "kiran@gmail.com"
	username := "kiran"
	password := "kiran123"

	inputs = TSignInputs{
		Email:    email,
		Password: password,
		UserName: username,
	}

	buf := new(bytes.Buffer)
	err := json.NewEncoder(buf).Encode(&inputs)

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

	resp, _ = app.Test(req, -1)

	var p []byte
	_, err = resp.Body.Read(p)

	if err != nil {
		panic(err)
	}

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&data)

	if err != nil {
		panic(err)
	}

	return resp, inputs, data
}

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
				type LoginInputs struct {
					Email    string `json:"email"`
					Password string `json:"password"`
				}

				email := "kiran@abc.com"
				password := "kiran123"

				body := LoginInputs{
					Email:    email,
					Password: password,
				}

				buf := new(bytes.Buffer)
				err := json.NewEncoder(buf).Encode(&body)

				if err != nil {
					panic(err)
				}

				req := httptest.NewRequest("POST", "/api/v1/login", buf)
				req.Header.Set("Content-Type", "application/json")

				resp, _ := app.Test(req, -1)

				var p []byte
				_, err = resp.Body.Read(p)

				if err != nil {
					panic(err)
				}

				g.Assert(resp.StatusCode).Equal(401)
			})

			g.It("returns a token on valid user inputs", func() {
				// first step is to signup the user
				resp, inputs, _ := TSignup(app)
				g.Assert(resp.StatusCode).Equal(201)

				type LoginInputs struct {
					Email    string `json:"email"`
					Password string `json:"password"`
				}

				body := LoginInputs{
					Email:    inputs.Email,
					Password: inputs.Password,
				}

				buf := new(bytes.Buffer)
				err := json.NewEncoder(buf).Encode(&body)

				if err != nil {
					panic(err)
				}

				req := httptest.NewRequest("POST", "/api/v1/login", buf)
				req.Header.Set("Content-Type", "application/json")

				resp, _ = app.Test(req, -1)

				var p []byte
				_, err = resp.Body.Read(p)

				if err != nil {
					panic(err)
				}

				var data struct {
					Data struct {
						Token string `json:"token"`
					} `json:"data"`
					Message string `json:"message"`
				}

				decoder := json.NewDecoder(resp.Body)
				err = decoder.Decode(&data)

				if err != nil {
					panic(err)
				}

				g.Assert(resp.StatusCode).Equal(200)
				assert.NotNil(t, data.Data.Token)
			})
		})
	})
}
