package passwords

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword (password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)

	return string(hash), err
}

func CheckPassword (password, hash string) bool {
	err := bcrypt.CompareHashAndPassword(
		[]byte(hash),
		[]byte(password),
	)

	return err==nil
}

// Token hashing using SHA256
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