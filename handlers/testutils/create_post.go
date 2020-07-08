package testutils

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/gofiber/fiber"
)

type TCreatePostInputs struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type TCreatePostResponse struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Author      struct {
		ID       string `json:"id"`
		Username string `json:"username"`
	} `json:"author"`
}

func TCreatePost(app *fiber.App, token string) (*http.Response, TCreatePostInputs, TCreatePostResponse) {

	inputs := TCreatePostInputs{
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
				"Authorization": "Bearer " + token,
				"Content-Type":  "application/json",
			},
		},
	})

	resp, _ := app.Test(req, -1)

	var output TCreatePostResponse

	if resp.StatusCode == 201 {
		err := json.NewDecoder(resp.Body).Decode(&output)
		if err != nil {
			panic(err)
		}
	}

	return resp, inputs, output
}
