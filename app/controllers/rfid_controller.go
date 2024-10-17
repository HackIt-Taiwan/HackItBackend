package controllers

import (
	"context"
	"os"
	"time"

	"hackitbackend/app/models"
	"hackitbackend/pkg/utils"
	"hackitbackend/platform/database"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
)

func VerifyUser(c *fiber.Ctx) error {
	verificationRequest := new(models.VerificationRequest)

	// Parse the request body
	if err := c.BodyParser(verificationRequest); err != nil {
		return utils.ResponseMsg(c, 400, err.Error(), nil)
	}

	// Ensure the verification code and secret are not empty
	if verificationRequest.ValidCode == "" || verificationRequest.Secret == "" {
		return utils.ResponseMsg(c, 400, "Valid code and secret are required", nil)
	}

	// Check if the secret key is correct
	if verificationRequest.Secret != os.Getenv("OFFICIAL_SECRET_KEY") {
		return utils.ResponseMsg(c, 401, "Invalid secret key", nil)
	}

	// Prepare the query
	ctx := context.Background()
	var user models.Users

	var err error
	if verificationRequest.UserID != "" {
		// Check if user exists by UserID
		err = database.Db.Collection("users").FindOne(ctx, bson.M{"id": verificationRequest.UserID}).Decode(&user)
	} else if verificationRequest.UserNumber != "" {
		// Check if user exists by UserNumber (assuming UserNumber is stored as a unique identifier)
		err = database.Db.Collection("users").FindOne(ctx, bson.M{"user_number": verificationRequest.UserNumber}).Decode(&user)
	} else {
		return utils.ResponseMsg(c, 400, "Either user_id or user_number must be provided", nil)
	}

	if err != nil {
		return utils.ResponseMsg(c, 404, "User not found", nil)
	}

	// Check if the user's verification code matches
	if user.VerificationCode != verificationRequest.ValidCode {
		return utils.ResponseMsg(c, 401, "Invalid verification code", nil)
	}

	// Verification success
	user.IsVerified = true
	user.UpdatedAt = time.Now()

	// Update user status
	_, err = database.Db.Collection("users").UpdateOne(ctx, bson.M{"id": user.ID}, bson.M{"$set": user})
	if err != nil {
		return utils.ResponseMsg(c, 500, "Failed to update user status", err.Error())
	}

	return utils.ResponseMsg(c, 200, "User verified successfully", user)
}
