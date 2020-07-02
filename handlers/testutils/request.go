package testutils

import (
	"io"
	"net/http"
	"net/http/httptest"
)

type Map map[string]string

type Opt struct {
	Header Map
}

type Req struct {
	Method  string    `json:"mothod"`
	Target  string    `json:"target"`
	Body    io.Reader `json:"body"`
	Options Opt       `json:"options"`
}

func MakeRequest(inputs Req) *http.Request {
	// http.Request
	req := httptest.NewRequest(
		inputs.Method,
		inputs.Target,
		inputs.Body,
	)

	for key, val := range inputs.Options.Header {
		req.Header.Set(key, val)
	}

	return req
}
