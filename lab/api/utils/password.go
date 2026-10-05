package utils

import (
	"errors"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"
)

const MinPasswordLength = 8

// ValidatePassword enforces a minimum length. Complexity rules can be tightened later.
func ValidatePassword(plain string) error {
	if utf8.RuneCountInString(plain) < MinPasswordLength {
		return errors.New("password must be at least 8 characters")
	}
	return nil
}

// HashPassword hashes a plaintext password using bcrypt.
func HashPassword(plain string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

// CheckPassword compares a bcrypt-hashed password with its possible plaintext equivalent.
func CheckPassword(hashed string, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain)) == nil
}
