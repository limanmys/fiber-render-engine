package permission

import (
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/limanmys/render-engine/app/models"
)

func TestFileRoutesEnforceTypedPermissions(t *testing.T) {
	const serverID = "11111111-1111-4111-8111-111111111111"
	const extensionID = "22222222-2222-4222-8222-222222222222"
	server := models.Permission{Type: "server", Key: "id", Value: serverID}
	extension := models.Permission{Type: "extension", Key: "id", Value: extensionID}
	for _, route := range []string{"/putFile", "/getFile", "/script"} {
		for _, item := range []struct {
			name        string
			permissions []models.Permission
			status      int
			admin       bool
		}{
			{name: "no server access", permissions: []models.Permission{extension}, status: 403},
			{name: "no extension access", permissions: []models.Permission{server}, status: 403},
			{name: "wrong object type", permissions: []models.Permission{{Type: "variable", Key: "id", Value: serverID}, extension}, status: 403},
			{name: "wrong key", permissions: []models.Permission{{Type: "server", Key: "name", Value: serverID}, extension}, status: 403},
			{name: "authorized", permissions: []models.Permission{server, extension}, status: 204},
			{name: "administrator", status: 204, admin: true},
		} {
			t.Run(route+"/"+item.name, func(t *testing.T) {
				app := fiber.New()
				app.Use(func(c *fiber.Ctx) error { c.Locals("user_id", "caller"); return c.Next() })
				app.Use(permissionHandler(func(user *models.User) (*models.User, error) {
					if user.ID != "caller" {
						t.Fatal("caller identity was changed")
					}
					if item.admin {
						user.Status = 1
					}
					return user, nil
				}, func(*models.User) ([]models.Permission, error) { return item.permissions, nil }, func(code int, message string) *fiber.Error { return fiber.NewError(code, message) }))
				app.Post(route, func(c *fiber.Ctx) error { return c.SendStatus(204) })
				request := httptest.NewRequest("POST", route, strings.NewReader(url.Values{"server_id": {serverID}, "extension_id": {extensionID}}.Encode()))
				request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				response, err := app.Test(request)
				if err != nil {
					t.Fatal(err)
				}
				response.Body.Close()
				if response.StatusCode != item.status {
					t.Fatalf("expected %d, got %d", item.status, response.StatusCode)
				}
			})
		}
	}
}
