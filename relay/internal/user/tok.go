package user

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type TokenService struct {
	secret []byte
}

func NewTokenService(secret string) *TokenService {
	return &TokenService{
		secret: []byte(secret),
	}
}

func (t *TokenService) Generate(userID uuid.UUID) (string, error) {
	now := time.Now()

	claims := jwt.MapClaims{
		"sub": userID.String(),
		"iat": now.Unix(),
		"exp": now.Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString(t.secret)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}

	return signedToken, nil
}
