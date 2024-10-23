package main

import (
	"os"

	"hackitbackend/pkg/middleware"
	"hackitbackend/pkg/routes"
	"hackitbackend/pkg/utils"
	"hackitbackend/platform/database"

	"github.com/gofiber/fiber/v2"

	_ "github.com/joho/godotenv/autoload" // load .env file automatically
)

func main() {
	// Define Fiber config.
	app := fiber.New()
	database.Connect()

	// Middlewares.
	middleware.FiberMiddleware(app)

	routes.RFIDRoutes(app)
	routes.UserRoutes(app)
	routes.NotFoundRoute(app)

	// Start server (with or without graceful shutdown).
	if os.Getenv("STAGE_STATUS") == "dev" {
		utils.StartServer(app)
	} else {
		utils.StartServerWithGracefulShutdown(app)
	}
}
