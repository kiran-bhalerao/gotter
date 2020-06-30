package utils

import (
	"os"
)

func GoDotEnvVariable(key string) string {
	return os.Getenv(key)
}
