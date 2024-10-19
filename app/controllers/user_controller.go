package controllers

import (
	"context"
	"time"

	"hackitbackend/app/models"
	"hackitbackend/pkg/utils"
	"hackitbackend/platform/database"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var ctx = context.Background()

func CreateUsers(c *fiber.Ctx) error {
	users := new(models.Users)

	if err := c.BodyParser(users); err != nil {
		return utils.ResponseMsg(c, 400, err.Error(), nil)
	}

	// Generate a new UUID and check if it already exists
	for {
		users.ID = uuid.New().String()
		
		// Check if the UUID already exists
		existingUser := new(models.Users)
		err := database.Db.Collection("users").FindOne(ctx, bson.M{"_id": users.ID}).Decode(existingUser)
		
		if err == mongo.ErrNoDocuments {
			// UUID doesn't exist, we can use it
			break
		} else if err != nil {
			// An error occurred during the database query
			return utils.ResponseMsg(c, 500, "Error checking user ID", err.Error())
		}
		// If we reach here, the UUID already exists, so we'll generate a new one in the next iteration
	}

	users.CreatedAt = time.Now()
	users.UpdatedAt = time.Now()
	var err error
	users.VerificationCode, err = utils.GenerateVerificationCode()
	if err != nil {
		return utils.ResponseMsg(c, 500, "Failed to generate verification code", err.Error())
	}

	_, err = database.Db.Collection("users").InsertOne(ctx, users)
	if err != nil {
		return utils.ResponseMsg(c, 500, "Failed to insert data", err.Error())
	}

	// TODO: Send verification email with the code

	return utils.ResponseMsg(c, 200, "User created successfully. Check email for verification code.", users)
}
