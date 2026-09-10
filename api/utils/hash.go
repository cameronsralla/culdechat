package utils

import (
	"crypto/sha256"
	"encoding/hex"
)

// HashOpaque hashes a high-entropy secret (invite token, refresh token) for storage/lookup.
func HashOpaque(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}
