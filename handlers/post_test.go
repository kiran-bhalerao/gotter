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
			g.It("returns 400 on invalid inputs @CREATE_POST", func() {
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
					Target: "/api/v1/post",
					Options: Opt{
						Header: Map{
							"Authorization": "Bearer " + userLogin.Data.Token,
						},
					},
				})

				resp, _ = app.Test(req, -1)

				g.Assert(resp.StatusCode).Equal(400)
			})

			g.It("return 401 on valid inputs but without auth token @CREATE_POST", func() {
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

			g.It("return 201 on valid inputs and token @CREATE_POST", func() {
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
				resp, inputs, out := TCreatePost(app, userLogin.Data.Token)

				// assert for success
				g.Assert(resp.StatusCode).Equal(201)

				g.Assert(out.Title).Equal(inputs.Title)
				g.Assert(out.Author.Username).Equal(userInputs.UserName)
				assert.NotNil(t, out.ID)
				assert.NotNil(t, out.Author.ID)
			})
		})

		g.Describe("Update Post Route Suit", func() {
			g.It("returns 400 on invalid inputs @UPDATE_POST", func() {
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
					Method: "PUT",
					Target: "/api/v1/post/123",
					Options: Opt{
						Header: Map{
							"Authorization": "Bearer " + userLogin.Data.Token,
						},
					},
				})

				resp, _ = app.Test(req, -1)

				g.Assert(resp.StatusCode).Equal(400)
			})

			g.It("return 401 on valid inputs but without auth token @UPDATE_POST", func() {
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
					Method: "PUT",
					Target: "/api/v1/post/123",
					Body:   buf,
				})

				resp, _ := app.Test(req, -1)
				g.Assert(resp.StatusCode).Equal(401)
			})

			g.It("returns 200 after updating the post successfully @UPDATE_POST", func() {
				// first step is to signup the user
				resp, userInputs, _ := TSignup(app)
				g.Assert(resp.StatusCode).Equal(201)

				// sec step is to login the user
				resp, userLogin := TLogin(app, TLoginInputs{
					Email:    userInputs.Email,
					Password: userInputs.Password,
				})
				g.Assert(resp.StatusCode).Equal(200)

				// third step is create post
				resp, inputs, post := TCreatePost(app, userLogin.Data.Token)

				// assert for success
				g.Assert(resp.StatusCode).Equal(201)

				//try to update the post
				type UpdatePostInput struct {
					Title       string `json:"title"`
					Description string `json:"description"`
				}

				updates := UpdatePostInput{
					Title:       "updated title",
					Description: "updated description",
				}

				buf := new(bytes.Buffer)
				err := json.NewEncoder(buf).Encode(&updates)

				if err != nil {
					panic(err)
				}

				req := MakeRequest(Req{
					Method: "PUT",
					Body:   buf,
					Target: "/api/v1/post/" + post.ID,
					Options: Opt{
						Header: Map{
							"Authorization": "Bearer " + userLogin.Data.Token,
							"Content-Type":  "application/json",
						},
					},
				})

				resp, _ = app.Test(req, -1)
				g.Assert(resp.StatusCode).Equal(200)

				var updatePostResponse struct {
					ID          string `json:"id"`
					Title       string `json:"title"`
					Description string `json:"description"`
					Author      struct {
						ID       string `json:"id"`
						Username string `json:"username"`
					} `json:"author"`
				}

				err = json.NewDecoder(resp.Body).Decode(&updatePostResponse)
				if err != nil {
					panic(err)
				}

				g.Assert(updatePostResponse.Title).Equal(updates.Title)
				g.Assert(updatePostResponse.Description).Equal(updates.Description)
				g.Assert(updatePostResponse.Author.Username).Equal(userInputs.UserName)

				assert.NotNil(t, updatePostResponse.ID)
				assert.NotEqual(t, updates.Title, inputs.Title)
				assert.NotEqual(t, updates.Description, inputs.Description)
			})
		})

		g.Describe("Delete Post Route Suit", func() {
			g.It("returns 400 on invalid inputs @DELETE_POST", func() {
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
					Method: "DELETE",
					Target: "/api/v1/post/123",
					Options: Opt{
						Header: Map{
							"Authorization": "Bearer " + userLogin.Data.Token,
						},
					},
				})

				resp, _ = app.Test(req, -1)

				g.Assert(resp.StatusCode).Equal(400)
			})

			g.It("returns 401 on valid inputs but without auth token @DELETE_POST", func() {
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

				req := MakeRequest(Req{
					Method: "DELETE",
					Target: "/api/v1/post/" + post.ID,
				})

				resp, _ = app.Test(req, -1)
				g.Assert(resp.StatusCode).Equal(401)
			})

			g.It("returns 200 on successful deletion of a post @DELETE_POST", func() {
				// first step is to signup the user
				resp, user, _ := TSignup(app)
				g.Assert(resp.StatusCode).Equal(201)

				// sec step is to login the user
				resp, data := TLogin(app, TLoginInputs{
					Email:    user.Email,
					Password: user.Password,
				})
				g.Assert(resp.StatusCode).Equal(200)

				// create a post
				resp, _, output := TCreatePost(app, data.Data.Token)

				// assert for success
				g.Assert(resp.StatusCode).Equal(201)

				req := MakeRequest(Req{
					Method: "DELETE",
					Target: "/api/v1/post/" + output.ID,
					Options: Opt{
						Header: Map{
							"Authorization": "Bearer " + data.Data.Token,
						},
					},
				})

				resp, _ = app.Test(req, -1)
				g.Assert(resp.StatusCode).Equal(200)
			})

			g.It("does not allow another user to delete the post @DELETE_POST", func() {
				// first step is to signup the user
				resp, userOneInputs, _ := TSignup(app)
				g.Assert(resp.StatusCode).Equal(201)

				// sec step is to login the user
				resp, userOneLogin := TLogin(app, TLoginInputs{
					Email:    userOneInputs.Email,
					Password: userOneInputs.Password,
				})
				g.Assert(resp.StatusCode).Equal(200)

				// create a post
				resp, _, post := TCreatePost(app, userOneLogin.Data.Token)

				// assert for success
				g.Assert(resp.StatusCode).Equal(201)

				// create another user
				resp, userTwoInputs, _ := TSignup(app, TSignInputs{
					Email:    "sec@user.com",
					UserName: "sec_user",
					Password: "password",
				})
				g.Assert(resp.StatusCode).Equal(201)

				// login second user and get the token of it
				resp, userTwoLogin := TLogin(app, TLoginInputs{
					Email:    userTwoInputs.Email,
					Password: userTwoInputs.Password,
				})
				g.Assert(resp.StatusCode).Equal(200)

				req := MakeRequest(Req{
					Method: "DELETE",
					Target: "/api/v1/post/" + post.ID,
					Options: Opt{
						Header: Map{
							"Authorization": "Bearer " + userTwoLogin.Data.Token,
						},
					},
				})

				resp, _ = app.Test(req, -1)
				g.Assert(resp.StatusCode).Equal(500)
			})
		})
	})
}
