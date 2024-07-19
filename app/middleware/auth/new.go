package auth

import (
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/limanmys/render-engine/internal/liman"
	"github.com/limanmys/render-engine/pkg/helpers"
	"github.com/limanmys/render-engine/pkg/logger"
)

// Creates a new authorization instance
func New() fiber.Handler {
	return authorization
}

type Cookie struct {
	Token string `cookie:"token"`
}

// authorization Middleware auths users before requests
func authorization(c *fiber.Ctx) error {
	if len(c.FormValue("liman-token")) > 0 {
		user, err := liman.AuthWithAccessToken(
			strings.Trim(c.FormValue("liman-token"), ""),
		)

		if err != nil {
			return logger.FiberError(fiber.StatusUnauthorized, err.Error())
		}

		c.Locals("user_id", user)
		return c.Next()
	}

	if len(string(c.Request().Header.Peek("Authorization"))) > 0 {
		code, err := helpers.LaravelAesDecrypt("token", c.FormValue("token"))
		if err != nil {
			return jwtValidation(c, c.FormValue("token"))
		}

		return jwtValidation(c, code)
	}

	cookie := new(Cookie)
	c.CookieParser(cookie)

	if len(cookie.Token) > 0 {
		decoded, err := url.QueryUnescape(cookie.Token)
		if err != nil {
			logger.FiberError(fiber.StatusUnauthorized, "invalid authorization token (cookie), "+err.Error())
		}

		if len(decoded) < 1 {
			return logger.FiberError(fiber.StatusUnauthorized, "authorization token is missing")
		}

		code, err := helpers.LaravelAesDecrypt("token", decoded)
		if err != nil {
			return jwtValidation(c, decoded)
		}

		return jwtValidation(c, code)
	}

	if len(c.FormValue("token")) > 0 {
		code, err := helpers.LaravelAesDecrypt("token", c.FormValue("token"))
		if err != nil {
			return jwtValidation(c, c.FormValue("token"))
		}

		return jwtValidation(c, code)
	}

	return logger.FiberError(fiber.StatusUnauthorized, "authorization token is missing")
}

func jwtValidation(c *fiber.Ctx, code string) error {
	token, err := jwt.Parse(code, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, logger.FiberError(fiber.StatusUnauthorized, "invalid authorization token")
		}
		return []byte(helpers.Env("JWT_SECRET", "")), nil
	})

	if err != nil {
		return logger.FiberError(fiber.StatusUnauthorized, "invalid authorization token")
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		c.Locals("user_id", claims["sub"])
		c.Locals("token", code)
		return c.Next()
	} else {
		return logger.FiberError(fiber.StatusUnauthorized, "invalid authorization token")
	}
}
