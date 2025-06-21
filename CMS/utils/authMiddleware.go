package utils

import (
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		cookie := c.Get("auth_token")
		if cookie == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		secretKey, exist := os.LookupEnv("SECRET_KEY")
		if !exist {
			panic("SECRET_KEY not found")
		}

		token, err := jwt.ParseWithClaims(cookie, &AuthClaims{}, func(t *jwt.Token) (interface{}, error) {
			return []byte(secretKey), nil
		})

		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		claims, ok := token.Claims.(*AuthClaims)
		if !ok || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		c.Locals("user_id", claims.Id)
		c.Locals("username", claims.Username)

		return c.Next()
	}

}
