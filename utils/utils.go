package utils

import (
	"os"
)

func GetenvWithDefault(name, defaultValue string) string {
	v := os.Getenv(name)
	if v != "" {
		return v
	}
	return defaultValue
}
