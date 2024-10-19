package utils

import (
    "crypto/rand"
    "fmt"
    "math/big"
)

func GenerateVerificationCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(899999))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", 100000+n.Int64()), nil
}
