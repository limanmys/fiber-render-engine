package helpers

import (
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func GetFormData(c *fiber.Ctx) map[string]string {
	multipart, err := c.MultipartForm()

	if err != nil {
		body := BodyParser(c.Body())

		return body
	}

	formValues := make(map[string]string)
	for key, value := range multipart.Value {
		if len(value) > 0 {
			formValues[key] = value[0]
		}
	}

	return formValues
}

func BodyParser(body []byte) map[string]string {
	values, err := url.ParseQuery(string(body))
	if err != nil {
		return make(map[string]string)
	}

	obj := map[string]string{}
	for k, v := range values {
		if strings.Contains(strings.ToLower(k), "password") || strings.Contains(strings.ToLower(k), "token") {
			continue
		}

		if len(v) > 0 {
			obj[k] = v[0]
		}
	}

	return obj
}
