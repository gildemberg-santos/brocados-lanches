package token

import (
	"encoding/base64"
	"testing"
)

func TestNew(t *testing.T) {
	first, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	second, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if first == second {
		t.Error("tokens are equal; they must be random")
	}
	decoded, err := base64.RawURLEncoding.DecodeString(first)
	if err != nil {
		t.Fatalf("decode token: %v", err)
	}
	if len(decoded) != TokenLengthBytes {
		t.Errorf("token bytes = %d, want %d", len(decoded), TokenLengthBytes)
	}
}

func TestHash(t *testing.T) {
	first := Hash("token-a")
	second := Hash("token-b")
	if first == second {
		t.Error("hashes are equal for different tokens")
	}
	if len(first) != 64 {
		t.Errorf("hash length = %d, want 64", len(first))
	}
	if first != Hash("token-a") {
		t.Error("hash is not deterministic")
	}
}
