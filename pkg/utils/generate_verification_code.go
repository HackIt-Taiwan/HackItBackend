package utils

import (
	"strconv"

	"golang.org/x/exp/rand"
)

func GenerateVerificationCode() string {
	return strconv.Itoa(100000 + rand.Intn(899999))
}