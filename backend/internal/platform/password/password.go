// Package password implementa hashing de senha com Argon2id conforme ADR 0004.
package password

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Parâmetros do Argon2id. Alterá-los exige rehash das senhas existentes.
const (
	memoryKiB      = 64 * 1024
	iterations     = 3
	parallelism    = 2
	saltLength     = 16
	keyLength      = 32
	encodedVersion = "19"
)

var prefix = fmt.Sprintf("$argon2id$v=%s$m=%d,t=%d,p=%d$", encodedVersion, memoryKiB, iterations, parallelism)

// Hash gera um hashArgon2id com salt aleatório no formato PHC simplificado.
func Hash(plain string) (string, error) {
	salt := make([]byte, saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate password salt: %w", err)
	}
	key := argon2.IDKey([]byte(plain), salt, iterations, memoryKiB, parallelism, keyLength)

	encoded := prefix + base64.RawStdEncoding.EncodeToString(salt) + "$" + base64.RawStdEncoding.EncodeToString(key)
	return encoded, nil
}

// Verify confere a senha em tempo constante contra o hash armazenado.
func Verify(encoded, plain string) bool {
	salt, expected, err := parse(encoded)
	if err != nil {
		return false
	}
	actual := argon2.IDKey([]byte(plain), salt, iterations, memoryKiB, parallelism, keyLength)
	return subtle.ConstantTimeCompare(actual, expected) == 1
}

func parse(encoded string) ([]byte, []byte, error) {
	if !strings.HasPrefix(encoded, prefix) {
		return nil, nil, fmt.Errorf("unsupported password hash")
	}
	parts := strings.Split(strings.TrimPrefix(encoded, prefix), "$")
	if len(parts) != 2 {
		return nil, nil, fmt.Errorf("invalid password hash")
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[0])
	if err != nil || len(salt) != saltLength {
		return nil, nil, fmt.Errorf("invalid password salt")
	}
	key, err := base64.RawStdEncoding.DecodeString(parts[1])
	if err != nil || len(key) != keyLength {
		return nil, nil, fmt.Errorf("invalid password key")
	}
	return salt, key, nil
}
