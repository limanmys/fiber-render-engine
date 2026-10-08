package handlers

import (
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestFileEndpointsRequireExtensionContextBeforeAccessingConnections(t *testing.T) {
	for _, item := range []struct {
		route   string
		handler fiber.Handler
	}{
		{"/putFile", PutFile}, {"/getFile", GetFile}, {"/script", ScriptRunner},
	} {
		t.Run(item.route, func(t *testing.T) {
			app := fiber.New()
			app.Post(item.route, item.handler)
			form := url.Values{"server_id": {"server"}, "local_path": {"/etc/fixture"}, "remote_path": {"fixture"}, "root": {"no"}, "parameters": {" "}}
			request := httptest.NewRequest("POST", item.route, strings.NewReader(form.Encode()))
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			response, err := app.Test(request)
			if err != nil {
				t.Fatal(err)
			}
			response.Body.Close()
			if response.StatusCode != 400 {
				t.Fatalf("expected missing-context 400, got %d", response.StatusCode)
			}
		})
	}
}
