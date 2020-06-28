package utils

import (
	"github.com/dgrijalva/jwt-go"
)

func CreateJWTToken(data map[string]interface{}) (string, error) {
	mapClaims := make(jwt.MapClaims, len(data))

	for key, val := range data {
		mapClaims[key] = val
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, mapClaims)
	// Sign and get the complete encoded token as a string using the secret
	tokenSecret := GoDotEnvVariable("JWT_SECRET")

	return token.SignedString([]byte(tokenSecret))
}
