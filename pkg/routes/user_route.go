package routes

import (
	"hackitbackend/app/controllers"

	"github.com/gofiber/fiber/v2"
)

// Related to user routes
func UserRoutes(app *fiber.App) {
	userGroup := app.Group("/users")

	userGroup.Post("/create", controllers.CreateUsers)
}