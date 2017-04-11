package gitlab

import (
	"net/http"
	"net/http/httptest"
	"golang.org/x/oauth2"
)

var (
	mux *http.ServeMux
	server *httptest.Server
)

func setup() (*Client, func(), error)  {
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)

	teardown := func() {
		server.Close()
	}
	NewEngine(&Config{
		BasePath: server.URL,
	})

	c := NewClient(&oauth2.Token{}, "qwerty")

	return c, teardown, nil
}
