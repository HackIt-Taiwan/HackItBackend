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
	"go.mongodb.org/mongo-driver/mongo"
	"github.com/golang-jwt/jwt/v5"
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

	accessToken, err := utils.GenerateNewAccessToken(user.ID, []string{})
	if err != nil {
		return utils.ResponseMsg(c, 500, "Failed to generate access token", nil)
	}

	jwtResponse := models.VerificationSession{
		UserID:    user.ID,
		Jwt:       accessToken,
	}
	return utils.ResponseMsg(c, 200, "User verified successfully", jwtResponse)
}


func AssignRFIDCard(c *fiber.Ctx) error {
	// Create a new request structure to receive card number and JWT from client
	type RFIDCardRequest struct {
		CardNumber string `json:"card_number"` // Card Number
		JWT        string `json:"JWT"`         // JWT
	}

	// Parse request body
	var request RFIDCardRequest
	if err := c.BodyParser(&request); err != nil {
		return utils.ResponseMsg(c, 400, "Invalid request body", nil)
	}

	// Check if CardNumber or JWT is empty
	if request.CardNumber == "" || request.JWT == "" {
		return utils.ResponseMsg(c, 400, "CardNumber and JWT are required", nil)
	}

	// Verify JWT token
	token, err := jwt.Parse(request.JWT, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET_KEY")), nil
	})
	if err != nil {
		return utils.ResponseMsg(c, 401, "Invalid JWT", nil)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return utils.ResponseMsg(c, 401, "Invalid JWT claims", nil)
	}

	userID, ok := claims["id"].(string)
	if !ok {
		return utils.ResponseMsg(c, 401, "Invalid user ID in JWT", nil)
	}

	// Prepare context
	ctx := context.Background()

	// Check if the user exists in the database
	var existingUser models.Users
	err = database.Db.Collection("users").FindOne(ctx, bson.M{"id": userID}).Decode(&existingUser)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return utils.ResponseMsg(c, 404, "User not found", nil)
		}
		return utils.ResponseMsg(c, 500, "Failed to check if user exists", nil)
	}

	// Check if the card already exists in the database
	var existingCard models.RFIDCard
	err = database.Db.Collection("rfid_cards").FindOne(ctx, bson.M{"card_number": request.CardNumber}).Decode(&existingCard)
	if err == nil {
		return utils.ResponseMsg(c, 400, "Card is already assigned to another user", nil)
	} else if err != mongo.ErrNoDocuments {
		return utils.ResponseMsg(c, 500, "Failed to check if card exists", nil)
	}

	// Create new RFID card record
	rfidCard := models.RFIDCard{
		UserID:     userID,
		CardNumber: request.CardNumber,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Insert new card info into database
	_, err = database.Db.Collection("rfid_cards").InsertOne(ctx, rfidCard)
	if err != nil {
		return utils.ResponseMsg(c, 500, "Failed to assign RFID card to user", nil)
	}

	// Return success message
	return utils.ResponseMsg(c, 200, "RFID card assigned successfully", nil)
}
