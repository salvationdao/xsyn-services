package crypto

import (
	"encoding/base64"

	"github.com/ninja-software/terror/v2"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword encrypts a plaintext string and returns the hashed version in base64
func HashPassword(password string) string {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(hashed)
}

// ComparePassword will compare the base64 encoded hash (from db) and compare to the supplied password (from login request).
// Will return error on fail, or wrong password. Returns nil for a correct password.
func ComparePassword(hashB64, password string) error {
	storedHash, err := base64.StdEncoding.DecodeString(hashB64)
	if err != nil {
		return terror.Error(err, "")
	}

	err = bcrypt.CompareHashAndPassword(storedHash, []byte(password))
	if err != nil {
		return terror.Error(err, "")
	}
	return nil
}
