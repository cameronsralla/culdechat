package utils

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTConfig holds env-driven settings for JWTs.
type JWTConfig struct {
	Secret         string
	Issuer         string
	AccessTokenTTL time.Duration
}

func getJWTConfig() JWTConfig {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = InsecureJWTSecret
	}
	issuer := os.Getenv("JWT_ISSUER")
	if issuer == "" {
		issuer = "culdechat"
	}
	ttlSeconds := 3600 // 1 hour default
	if v := os.Getenv("JWT_ACCESS_TTL_SECONDS"); v != "" {
		if d, err := time.ParseDuration(v + "s"); err == nil {
			ttlSeconds = int(d.Seconds())
		}
	}
	return JWTConfig{
		Secret:         secret,
		Issuer:         issuer,
		AccessTokenTTL: time.Duration(ttlSeconds) * time.Second,
	}
}

// Claims represents our JWT claims.
type Claims struct {
	UserID  string `json:"uid"`
	Unit    string `json:"unit"`
	IsAdmin bool   `json:"adm"`
	jwt.RegisteredClaims
}

// GenerateAccessToken creates a signed JWT for the given user ID, unit number, and admin flag.
func GenerateAccessToken(userID string, unit string, isAdmin bool) (string, error) {
	cfg := getJWTConfig()
	now := time.Now()
	claims := Claims{
		UserID:  userID,
		Unit:    unit,
		IsAdmin: isAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    cfg.Issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(cfg.AccessTokenTTL)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.Secret))
}

// ParseAndValidateToken parses and validates a JWT string.
func ParseAndValidateToken(tokenString string) (*Claims, error) {
	cfg := getJWTConfig()
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(cfg.Secret), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token claims")
}
