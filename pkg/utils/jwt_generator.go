package utils

import (
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateNewAccessToken(id string, credentials []string) (string, error) {
	// Set secret key from .env file.
	secret := os.Getenv("JWT_SECRET_KEY")

	// Set expires minutes count for secret key from .env file.
	minutesCount, _ := strconv.Atoi(os.Getenv("JWT_SECRET_KEY_EXPIRE_MINUTES_COUNT"))

	// Create a new claims.
	claims := jwt.MapClaims{}

	// Set public claims:
	claims["id"] = id
	claims["exp"] = time.Now().Add(time.Minute * time.Duration(minutesCount)).Unix()

	// Set private token credentials:
	for _, credential := range credentials {
		claims[credential] = true
	}

	// Create a new JWT access token with claims.
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)

	// Generate token.
	t, err := token.SignedString([]byte(secret))
	if err != nil {
		// Return error, it JWT token generation failed.
		return "", err
	}

	return t, nil
}