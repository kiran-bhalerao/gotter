package utils

import (
	"log"

	"golang.org/x/crypto/bcrypt"
)

type PasswordInterface interface {
	Hash() string
	Compare(hashedPassword string) bool
}

type Password struct {
	Password string
}

func (p Password) Hash() string {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(p.Password), bcrypt.DefaultCost)

	if err != nil {
		log.Fatal(err)
	}
	return string(hashedPassword)
}

func (p Password) Compare(hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(p.Password))

	return err == nil
}
