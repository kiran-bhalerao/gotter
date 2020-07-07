package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	. "github.com/franela/goblin"
	. "github.com/kiranbhalerao123/gotter/app"
	. "github.com/kiranbhalerao123/gotter/config"
	. "github.com/kiranbhalerao123/gotter/handlers/testutils"
	. "github.com/kiranbhalerao123/gotter/router"
	"github.com/stretchr/testify/assert"
)

func TestPostsRoute(t *testing.T) {
	g := Goblin(t)

	app := SetupApp()
	SetupDB()
	SetupRouter(app)

	g.Describe("Post Routes Test", func() {
		g.BeforeEach(func() {
			err := Mongo.DB.Drop(context.Background())

			if err != nil {
				panic(err)
			}
		})

		g.Describe("Create Post Route Suit", func() {
			g.It("returns 400 on invalid inputs::post", func() {
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
					Method: "POST",
					Target: "/api/v1/post",
					Options: Opt{
						Header: Map{
							"Authorization": "Bearer " + data.Data.Token,
						},
					},
				})

				resp, _ = app.Test(req, -1)

				g.Assert(resp.StatusCode).Equal(400)
			})

			g.It("return 401 on valid inputs but without auth token::post", func() {
				var inputs struct {
					Title       string `json:"title"`
					Description string `json:"description"`
				}

				buf := new(bytes.Buffer)
				err := json.NewEncoder(buf).Encode(&inputs)
				if err != nil {
					panic(err)
				}

				req := MakeRequest(Req{
					Method: "POST",
					Target: "/api/v1/post",
					Body:   buf,
				})

				resp, _ := app.Test(req, -1)
				g.Assert(resp.StatusCode).Equal(401)
			})

			g.It("return 201 on valid inputs and token::post", func() {
				// first step is to signup the user
				resp, user, _ := TSignup(app)
				g.Assert(resp.StatusCode).Equal(201)

				// sec step is to login the user
				resp, data := TLogin(app, TLoginInputs{
					Email:    user.Email,
					Password: user.Password,
				})
				g.Assert(resp.StatusCode).Equal(200)

				type Inputs struct {
					Title       string `json:"title"`
					Description string `json:"description"`
				}

				inputs := Inputs{
					Title:       "test title",
					Description: "test description",
				}

				buf := new(bytes.Buffer)
				err := json.NewEncoder(buf).Encode(&inputs)
				if err != nil {
					panic(err)
				}

				req := MakeRequest(Req{
					Method: "POST",
					Target: "/api/v1/post",
					Body:   buf,
					Options: Opt{
						Header: Map{
							"Authorization": "Bearer " + data.Data.Token,
							"Content-Type":  "application/json",
						},
					},
				})

				resp, _ = app.Test(req, -1)

				// assert for success
				g.Assert(resp.StatusCode).Equal(201)

				// if success then assert remaining
				var out struct {
					ID          string `json:"id"`
					Title       string `json:"title"`
					Description string `json:"description"`
					Author      struct {
						ID       string `json:"id"`
						Username string `json:"username"`
					} `json:"author"`
				}

				err = json.NewDecoder(resp.Body).Decode(&out)
				if err != nil {
					panic(err)
				}

				g.Assert(out.Title).Equal(inputs.Title)
				g.Assert(out.Author.Username).Equal(user.UserName)
				assert.NotNil(t, out.ID)
				assert.NotNil(t, out.Author.ID)
			})
		})
	})
}
