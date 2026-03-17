package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTmanger struct {
	AccessSecret  string
	RefreshSecret string
	AccessTTL     time.Duration
	RefreshTTL    time.Duration
	MaxSession    time.Duration
}

func GenerateJWT(access, refresh string, accessttl, refreshttl, maxSession time.Duration) *JWTmanger {
	return &JWTmanger{
		AccessSecret:  access,
		RefreshSecret: refresh,
		AccessTTL:     accessttl,
		RefreshTTL:    refreshttl,
		MaxSession:    maxSession,
	}
}

func (j *JWTmanger) GenerateAccessToken(userid, role string) (string, error) {

	claims := jwt.MapClaims{
		"user_id": userid,
		"role":    role,
		"exp":     time.Now().Add(15 * time.Minute).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(j.AccessSecret))
}

func (j *JWTmanger) GenerateRefreshToken(userID, role, sessionID string) (string, error) {

	claims := jwt.MapClaims{
		"user_id":    userID,
		"role":       role,
		"session_id": sessionID,
		"exp":        time.Now().Add(7 * 24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(j.RefreshSecret))
}

func (j *JWTmanger) ValidateAccess(tokenstr string) (map[string]interface{}, error) {
	token, err := jwt.Parse(tokenstr, func(t *jwt.Token) (any, error) {
		return []byte(j.AccessSecret), nil
	})

	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

func (j *JWTmanger) ValidateRefresh(tokenstr string) (map[string]interface{}, error) {
	token, err := jwt.Parse(tokenstr, func(t *jwt.Token) (any, error) {
		return []byte(j.RefreshSecret), nil
	})

	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
