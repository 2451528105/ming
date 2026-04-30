package tools

import (
	"crypto/rand"
	"encoding/base64"
)

func GenerateToken() string {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return ""
	}
	token := base64.RawURLEncoding.EncodeToString(tokenBytes)
	return token
}
