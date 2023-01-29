package helpers

import (
	"net/url"

	"github.com/gofiber/fiber/v2"
)

// GetFormData extracts form data or body data and returns it
func GetFormData(c *fiber.Ctx) map[string]string {
	multipart, err := c.MultipartForm()

	if err != nil {
		body := BodyParser(c.Body())

		return body
	}

	formValues := make(map[string]string)
	for key, value := range multipart.Value {
		if key == "" {
			continue
		}

		if len(value) > 0 {
			formValues[key] = value[0]
		}
	}

	return formValues
}

// BodyParser parses query and returns it
func BodyParser(body []byte) map[string]string {
	values, err := url.ParseQuery(string(body))
	if err != nil {
		return make(map[string]string)
	}

	obj := map[string]string{}
	for k, v := range values {
		if len(v) > 0 {
			obj[k] = v[0]
		}
	}

	return obj
}
