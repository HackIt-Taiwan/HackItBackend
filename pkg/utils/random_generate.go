package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
)

func GenerateVerificationCode() (string) {
	n, _ := rand.Int(rand.Reader, big.NewInt(899999))
	
	return fmt.Sprintf("%06d", 100000+n.Int64())
}

func GenerateRandomString(length int) (string, error) {
    bytes := make([]byte, length)
    _, err := rand.Read(bytes)
    if err != nil {
        return "", err
    }
	return hex.EncodeToString(bytes), nil
}
