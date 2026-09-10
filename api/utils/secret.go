package utils

import (
	"errors"
	"os"
	"strings"
)

const InsecureJWTSecret = "dev-insecure-secret-change-me"

// RequireJWTSecret refuses to run with a missing or well-known JWT secret unless
// CULDECHAT_ALLOW_INSECURE_JWT=true (local docker / tests).
func RequireJWTSecret() error {
	secret := strings.TrimSpace(os.Getenv("JWT_SECRET"))
	allow := strings.EqualFold(os.Getenv("CULDECHAT_ALLOW_INSECURE_JWT"), "true")
	if secret == "" || secret == InsecureJWTSecret {
		if !allow {
			return errors.New("JWT_SECRET must be set to a unique non-default value; set CULDECHAT_ALLOW_INSECURE_JWT=true only for local development")
		}
		if secret == "" {
			_ = os.Setenv("JWT_SECRET", InsecureJWTSecret)
			Warnf("JWT_SECRET missing; using insecure default because CULDECHAT_ALLOW_INSECURE_JWT=true")
		} else {
			Warnf("JWT_SECRET is the insecure default; allowed only because CULDECHAT_ALLOW_INSECURE_JWT=true")
		}
	}
	return nil
}
