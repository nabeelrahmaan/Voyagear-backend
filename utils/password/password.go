package passwords

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"

	"golang.org/x/crypto/bcrypt"
)

// Hashing string using bcrypt algorithm (mainly for passwords) - max length 72 bit
func HashPassword (password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)

	return string(hash), err
}

// Comparing stored hash and password from request using bcrypt algorithm
func CheckPassword (password, hash string) bool {
	err := bcrypt.CompareHashAndPassword(
		[]byte(hash),
		[]byte(password),
	)

	return err==nil
}

// Token hashing using SHA256 algorithm
func HashToken (token string) string {
	hashToken := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hashToken[:])
}

func CompareTokens (token string, storedHash string) bool {
	hash := HashToken(token)

	// It compares all byte. prevent timing attacks
	return subtle.ConstantTimeCompare(
		[]byte(hash),
		[]byte(storedHash),
	) == 1
}