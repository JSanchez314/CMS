package routes

import (
	"JSanches/CMD/services"

	"github.com/gofiber/fiber/v2"
)

func RegisterContentRoutes(router fiber.Router, ch *services.ContentHandler) {
	router.Post("/", ch.CreateContent)
	router.Get("/", ch.ListContents)
	router.Get("/:id", ch.GetContent)
	router.Put("/:id", ch.UpdateContent)
	router.Delete("/:id", ch.DeleteContent)
}
