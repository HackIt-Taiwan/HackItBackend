package utils

import (
	"fmt"
	"os"
)

func ConnectionURLBuilder(n string) (string, error) {
	var url string

	switch n {
	case "fiber":
		url = fmt.Sprintf(
			"%s:%s",
			os.Getenv("SERVER_HOST"),
			os.Getenv("SERVER_PORT"),
		)
	default:
		return "", fmt.Errorf("connection name '%v' is not supported", n)
	}

	return url, nil
}
