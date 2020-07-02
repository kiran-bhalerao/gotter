package testutils

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/gofiber/fiber"
)

type TLoginInputs struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type TLoginOutput struct {
	Data struct {
		Token string `json:"token"`
	} `json:"data"`
	Message string `json:"message"`
}

func TLogin(app *fiber.App, inputs TLoginInputs) (*http.Response, TLoginOutput) {
	buf := new(bytes.Buffer)
	err := json.NewEncoder(buf).Encode(&inputs)

	if err != nil {
		panic(err)
	}

	req := httptest.NewRequest("POST", "/api/v1/login", buf)
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req, -1)

	var data TLoginOutput

	if resp.StatusCode == 200 {
		decoder := json.NewDecoder(resp.Body)
		err := decoder.Decode(&data)

		if err != nil {
			panic(err)
		}
	}

	return resp, data
}
