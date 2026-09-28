package handlers

import (
	"io"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestVerifyKeepsLegacyResponsesAndOptsIntoSSHErrors(t *testing.T) {
	for _, detailed := range []bool{false, true} {
		app := fiber.New()
		app.Post("/verify", Verify)
		values := url.Values{
			"ip_address": {"192.0.2.1"}, "username": {"test"},
			"password": {"invalid-key-format"}, "port": {"22"}, "key_type": {"ssh_certificate"},
		}
		if detailed {
			values.Set("detailed_errors", "1")
		}
		request := httptest.NewRequest("POST", "/verify", strings.NewReader(values.Encode()))
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		response, err := app.Test(request)
		if err != nil {
			t.Fatal(err)
		}
		body, err := io.ReadAll(response.Body)
		response.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
		if response.StatusCode != 201 {
			t.Fatalf("legacy error status changed: %d", response.StatusCode)
		}
		if detailed && string(body) != `{"code":"SSH_PRIVATE_KEY_INVALID"}` {
			t.Fatalf("expected safe detailed error, got %s", body)
		}
		if !detailed && string(body) != "nok" {
			t.Fatalf("expected legacy nok, got %s", body)
		}
	}
}
