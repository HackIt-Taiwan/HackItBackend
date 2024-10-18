package routes

import (
	"hackitbackend/app/controllers"

	"github.com/gofiber/fiber/v2"
)

// Rfid routes
func RFIDRoutes(app *fiber.App) {
	userGroup := app.Group("/rfid")

	userGroup.Get("/userVerification", controllers.VerifyUser)
	userGroup.Post("/assignRFIDCard", controllers.AssignRFIDCard)
}
