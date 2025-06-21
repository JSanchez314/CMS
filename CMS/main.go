package main

import (
	"JSanches/CMD/database"
	"JSanches/CMD/routes"
	"JSanches/CMD/services"
	"JSanches/CMD/utils"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/joho/godotenv"
)

func main() {

	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	database.PostgresInit()
	database.MongoInit()
	database.RedisInit()

	port := os.Getenv("PORT")
	if port == "" {
		port = "4000"
	} else {
		port = ":" + port
	}

	userRepo := services.NewUserRepository(database.PostgresGetDB())
	tokenService := services.NewTokenService(os.Getenv("SECRET_KEY"))
	handler := &services.Handler{
		UserRepo:     userRepo,
		TokenService: tokenService,
	}

	contentHandler := services.NewContentHandler(
		&services.PGClient{},
		&services.MongoClient{},
		&services.RedisClient{},
	)

	app := fiber.New(fiber.Config{
		IdleTimeout: 10 * time.Second,
	})

	routes.PublicRoutes(app, handler)

	protected := app.Use(utils.AuthMiddleware())

	routes.RegisterContentRoutes(protected.Group("/collection"), contentHandler)

	go func() {
		if err := app.Listen(port); err != nil {
			log.Fatal("Error starting server: ", err)
		}
	}()

	app.Use(compress.New())
}
