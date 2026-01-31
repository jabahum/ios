package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
)

func RandomString(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)[:n]
}

func Sha256Bytes(s string) []byte {
	h := sha256.Sum256([]byte(s))
	return h[:]
}
