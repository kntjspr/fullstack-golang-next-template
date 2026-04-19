package auth

import (
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword converts plaintext password into bcrypt hash.
func HashPassword(password string) (string, error) {
	if password == "" {
		return "", errors.New("password cannot be empty")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

// ComparePassword verifies plaintext password against stored bcrypt hash.
func ComparePassword(passwordHash, password string) error {
	if strings.TrimSpace(passwordHash) == "" || password == "" {
		return errors.New("invalid password input")
	}

	return bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
}
