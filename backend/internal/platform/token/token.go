// Package token gera e transforma tokens opacos de sessão (ADR 0004).
package token

import (
	"crypto/hmac"
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

// DeriveOrderAccess emite o token de histórico a partir de uma chave opaca já
// conhecida pelo cliente (a chave de idempotência do primeiro pedido). O segredo
// do servidor impede que alguém derive o token apenas com essa chave.
func DeriveOrderAccess(secret, source string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte("brocados-order-access|"))
	mac.Write([]byte(source))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

// ValidateOrderAccess confere a forma do token de histórico antes de qualquer
// consulta. Tokens válidos têm 256 bits em base64url sem padding.
func ValidateOrderAccess(value string) error {
	if len(value) != 43 {
		return fmt.Errorf("invalid order access token length")
	}
	decoded, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil || len(decoded) != 32 {
		return fmt.Errorf("invalid order access token")
	}
	return nil
}
