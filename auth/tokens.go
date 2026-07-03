package auth

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	accessTokenTTL  = 15 * time.Minute
	refreshTokenTTL = 7 * 24 * time.Hour
)

type TokenClaims struct {
	UserID int    `json:"userId"`
	Email  string `json:"email"`
	Type   string `json:"type"`
	JTI    string `json:"jti,omitempty"`
	jwt.RegisteredClaims
}

func jwtSecret() ([]byte, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return nil, fmt.Errorf("JWT_SECRET is not set")
	}
	return []byte(secret), nil
}

func GenerateTokenPair(userID int, email string) (accessToken string, refreshToken string, refreshJTI string, err error) {
	secret, err := jwtSecret()
	if err != nil {
		return "", "", "", err
	}

	now := time.Now()

	accessToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, TokenClaims{
		UserID: userID,
		Email:  email,
		Type:   "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(accessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}).SignedString(secret)
	if err != nil {
		return "", "", "", err
	}

	refreshJTI = uuid.NewString()
	refreshToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, TokenClaims{
		UserID: userID,
		Email:  email,
		Type:   "refresh",
		JTI:    refreshJTI,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        refreshJTI,
			ExpiresAt: jwt.NewNumericDate(now.Add(refreshTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}).SignedString(secret)
	if err != nil {
		return "", "", "", err
	}

	return accessToken, refreshToken, refreshJTI, nil
}

func parseToken(tokenString string) (*TokenClaims, error) {
	secret, err := jwtSecret()
	if err != nil {
		return nil, err
	}

	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return secret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*TokenClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

func ValidateAccessToken(tokenString string) (*TokenClaims, error) {
	claims, err := parseToken(tokenString)
	if err != nil {
		return nil, err
	}
	if claims.Type != "access" {
		return nil, fmt.Errorf("invalid token type")
	}
	return claims, nil
}

func ValidateRefreshToken(tokenString string) (*TokenClaims, error) {
	claims, err := parseToken(tokenString)
	if err != nil {
		return nil, err
	}
	if claims.Type != "refresh" {
		return nil, fmt.Errorf("invalid token type")
	}
	if claims.JTI == "" {
		return nil, fmt.Errorf("refresh token missing jti")
	}
	return claims, nil
}

func RefreshTokenExpiry() time.Time {
	return time.Now().UTC().Add(refreshTokenTTL)
}
