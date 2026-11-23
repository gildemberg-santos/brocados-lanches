// Package token gera e transforma tokens opacos de sessão (ADR 0004).
package token

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

// TokenLengthBytes é o tamanho do token opaco: 256 bits de aleatoriedade.
const TokenLengthBytes = 32

// New devolve um token opaco em base64url sem padding.
func New() (string, error) {
	buffer := make([]byte, TokenLengthBytes)
	if _, err := rand.Read(buffer); err != nil {
		return "", fmt.Errorf("generate session token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buffer), nil
}

// Hash devolve o SHA-256 do token em hexadecimal. O token em claro nunca é gravado.
func Hash(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}
