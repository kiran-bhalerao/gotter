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
)

func TestCommentsRoute(t *testing.T) {
	g := Goblin(t)

	app := SetupApp()
	SetupDB()
	SetupRouter(app)

	g.Describe("Comment Routes Test", func() {
		g.BeforeEach(func() {
			err := Mongo.DB.Drop(context.Background())

			if err != nil {
				panic(err)
			}
		})

		g.Describe("Create Comment Route Suits", func() {
			g.It("returns 400 on invalid inputs @CREATE_COMMENT", func() {
				// first step is to signup the user
				resp, userInputs, _ := TSignup(app)
				g.Assert(resp.StatusCode).Equal(201)

				// sec step is to login the user
				resp, userLogin := TLogin(app, TLoginInputs{
					Email:    userInputs.Email,
					Password: userInputs.Password,
				})
				g.Assert(resp.StatusCode).Equal(200)

				req := MakeRequest(Req{
					Method: "POST",
					Target: "/api/v1/comment",
					Options: Opt{
						Header: Map{
							"Authorization": "Bearer " + userLogin.Data.Token,
						},
					},
				})

				resp, _ = app.Test(req, -1)

				g.Assert(resp.StatusCode).Equal(400)
			})

			g.It("returns 401 on valid inputs but without auth token @CREATE_COMMENT", func() {
				var inputs struct {
					PostId  string `json:"postId"`
					Message string `json:"message"`
				}

				buf := new(bytes.Buffer)
				err := json.NewEncoder(buf).Encode(&inputs)
				if err != nil {
					panic(err)
				}

				req := MakeRequest(Req{
					Method: "POST",
					Target: "/api/v1/comment",
					Body:   buf,
				})

				resp, _ := app.Test(req, -1)

				g.Assert(resp.StatusCode).Equal(401)
			})

			g.It("create a comment on valid inputs and valid auth token @CREATE_COMMENT", func() {
				// first step is to signup the user
				resp, userInputs, _ := TSignup(app)
				g.Assert(resp.StatusCode).Equal(201)

				// sec step is to login the user
				resp, userLogin := TLogin(app, TLoginInputs{
					Email:    userInputs.Email,
					Password: userInputs.Password,
				})
				g.Assert(resp.StatusCode).Equal(200)

				// create a post
				resp, _, post := TCreatePost(app, userLogin.Data.Token)

				// assert for success
				g.Assert(resp.StatusCode).Equal(201)

				type CreateCommentInputs struct {
					PostId  string `json:"postId"`
					Message string `json:"message"`
				}

				input := CreateCommentInputs{
					PostId:  post.ID,
					Message: "Test Comment",
				}

				buf := new(bytes.Buffer)
				err := json.NewEncoder(buf).Encode(&input)
				if err != nil {
					panic(err)
				}

				req := MakeRequest(Req{
					Method: "POST",
					Target: "/api/v1/comment",
					Body:   buf,
					Options: Opt{
						Header: Map{
							"Authorization": "Bearer " + userLogin.Data.Token,
							"Content-Type":  "application/json",
						},
					},
				})

				resp, _ = app.Test(req, -1)

				g.Assert(resp.StatusCode).Equal(201)
			})
		})

		g.Describe("Delete Comment Route Suits", func() {
			g.It("returns 400 on invalid inputs @DELETE_COMMENT", func() {
				// first step is to signup the user
				resp, userInputs, _ := TSignup(app)
				g.Assert(resp.StatusCode).Equal(201)

				// sec step is to login the user
				resp, userLogin := TLogin(app, TLoginInputs{
					Email:    userInputs.Email,
					Password: userInputs.Password,
				})
				g.Assert(resp.StatusCode).Equal(200)

				req := MakeRequest(Req{
					Method: "POST",
					Target: "/api/v1/comment/123",
					Options: Opt{
						Header: Map{
							"Authorization": "Bearer " + userLogin.Data.Token,
						},
					},
				})

				resp, _ = app.Test(req, -1)

				g.Assert(resp.StatusCode).Equal(400)
			})
		})
	})
}
