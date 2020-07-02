package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	. "github.com/franela/goblin"
	. "github.com/kiranbhalerao123/gotter/app"
	. "github.com/kiranbhalerao123/gotter/config"
	. "github.com/kiranbhalerao123/gotter/handlers/testutils"
	. "github.com/kiranbhalerao123/gotter/router"
	"github.com/stretchr/testify/assert"
)

func TestUserRoutes(t *testing.T) {
	g := Goblin(t)

	app := SetupApp()
	SetupDB()
	SetupRouter(app)

	g.Describe("User Routes Test", func() {
		g.BeforeEach(func() {
			err := Mongo.DB.Drop(context.Background())

			if err != nil {
				panic(err)
			}
		})

		g.Describe("GET User Route Suits", func() {
			g.It("returns 400 on invalid request", func() {
				// http.Request
				req := httptest.NewRequest(
					"GET",
					"/api/v1/user",
					nil,
				)

				resp, _ := app.Test(req, -1)
				g.Assert(400).Equal(resp.StatusCode)
			})

			g.It("returns user details on valid inputs", func() {
				// first step is to signup the user
				resp, inputs, _ := TSignup(app)
				g.Assert(resp.StatusCode).Equal(201)

				// sec step is to login the user
				resp, data := TLogin(app, TLoginInputs{
					Email:    inputs.Email,
					Password: inputs.Password,
				})
				g.Assert(resp.StatusCode).Equal(200)

				req := MakeRequest(Req{
					Method: "GET",
					Target: "/api/v1/user",
					Body:   nil,
					Options: Opt{
						Header: Map{
							"Authorization": "Bearer " + data.Data.Token,
						},
					},
				})

				resp, _ = app.Test(req, -1)

				var output struct {
					ID       string `json:"id"`
					Email    string `json:"email"`
					UserName string `json:"username"`
				}

				err := json.NewDecoder(resp.Body).Decode(&output)
				if err != nil {
					panic(err)
				}

				g.Assert(resp.StatusCode).Equal(200)
				g.Assert(output.Email).Equal(inputs.Email)
				g.Assert(output.UserName).Equal(inputs.UserName)
				assert.NotNil(t, output.ID)
			})

			g.It("updates the user details", func() {
				// first step is to signup the user
				resp, inputs, _ := TSignup(app)
				g.Assert(resp.StatusCode).Equal(201)

				// sec step is to login the user
				resp, data := TLogin(app, TLoginInputs{
					Email:    inputs.Email,
					Password: inputs.Password,
				})
				g.Assert(resp.StatusCode).Equal(200)

				type UpdateInputs struct {
					UserName string `json:"username"`
					Password string `json:"password"`
				}

				update := UpdateInputs{
					UserName: "kiran_up",
					Password: "up_password",
				}

				buf := new(bytes.Buffer)
				err := json.NewEncoder(buf).Encode(&update)

				if err != nil {
					panic(err)
				}

				req := MakeRequest(Req{
					Method: "PUT",
					Target: "/api/v1/user",
					Body:   buf,
					Options: Opt{
						Header: Map{
							"Authorization": "Bearer " + data.Data.Token,
							"Content-Type":  "application/json",
						},
					},
				})

				resp, _ = app.Test(req, -1)

				var output struct {
					ID       string `json:"id"`
					UserName string `json:"username"`
				}

				err = json.NewDecoder(resp.Body).Decode(&output)
				if err != nil {
					panic(err)
				}

				g.Assert(resp.StatusCode).Equal(200)
				g.Assert(output.UserName).Equal(update.UserName)
				assert.NotNil(t, output.ID)

				// try login with old credentials
				resp, data = TLogin(app, TLoginInputs{
					Email:    inputs.Email,
					Password: inputs.Password,
				})
				g.Assert(resp.StatusCode).Equal(401)
			})
		})
	})
}
