package routes

import (
	"JSanches/CMD/services"

	"github.com/gofiber/fiber/v2"
)

func PublicRoutes(app *fiber.App, handler *services.Handler) {
	app.Post("/register", handler.RegisterHandler)
	app.Post("/login", handler.LoginHandler)
}
