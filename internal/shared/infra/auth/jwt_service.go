package auth

import (
	"errors"
	"fmt"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

type TokenClaims struct {
	UserID string
	Name   string
	Email  string
	Role   string
}

type JWTService interface {
	ValidateToken(tokenString string) (*TokenClaims, error)
}

type JWTServiceImpl struct {
	secretKey []byte
}

type customClaims struct {
	UserID string `json:"user_id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func NewJWTService() *JWTServiceImpl {
	secretKey := os.Getenv("JWT_SECRET_KEY")
	if secretKey == "" {
		panic("JWT_SECRET_KEY environment variable is not set")
	}

	return &JWTServiceImpl{secretKey: []byte(secretKey)}
}

func (s *JWTServiceImpl) ValidateToken(tokenString string) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &customClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secretKey, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*customClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return &TokenClaims{
		UserID: claims.UserID,
		Name:   claims.Name,
		Email:  claims.Email,
		Role:   claims.Role,
	}, nil
}
