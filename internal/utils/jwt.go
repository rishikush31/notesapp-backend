package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
)

// ---------------- ACCESS TOKEN ----------------

func GenerateAccessToken(
	userID string,
	secret string,
	ttl time.Duration,
) (string, error) {

	claims := jwt.RegisteredClaims{
		Subject:   userID,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ---------------- REFRESH TOKEN ----------------

type RefreshTokenClaims struct {
	UserID     string `json:"user_id"`
	TokenID    string `json:"token_id"`
	DeviceInfo string `json:"device_info"`
	jwt.RegisteredClaims
}

func VerifyRefreshToken(tokenStr, secret string) (*RefreshTokenClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenStr,
		&RefreshTokenClaims{},
		func(t *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		},
	)

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*RefreshTokenClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
