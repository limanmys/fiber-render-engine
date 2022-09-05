package helpers

import "github.com/gofiber/fiber/v2"

func GetFormData(c *fiber.Ctx) map[string]string {
	multipart, err := c.MultipartForm()

	if err != nil {
		return make(map[string]string)
	}

	formValues := make(map[string]string)
	for key, value := range multipart.Value {
		if len(value) > 0 {
			formValues[key] = value[0]
		}
	}

	return formValues
}
