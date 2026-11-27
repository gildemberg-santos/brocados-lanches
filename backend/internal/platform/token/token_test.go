package token

import (
	"encoding/base64"
	"strings"
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

func TestDeriveOrderAccess(t *testing.T) {
	first := DeriveOrderAccess("segredo", "chave-1")
	second := DeriveOrderAccess("segredo", "chave-1")
	differentKey := DeriveOrderAccess("segredo", "chave-2")
	differentSecret := DeriveOrderAccess("outro-segredo", "chave-1")

	if first != second {
		t.Error("derived token is not deterministic")
	}
	if first == differentKey || first == differentSecret {
		t.Error("derived token must depend on secret and source")
	}
	if err := ValidateOrderAccess(first); err != nil {
		t.Fatalf("ValidateOrderAccess() error = %v", err)
	}
}

func TestValidateOrderAccessRejectsInvalidTokens(t *testing.T) {
	for _, value := range []string{"", "curto", strings.Repeat("a", 44), strings.Repeat("*", 43)} {
		if err := ValidateOrderAccess(value); err == nil {
			t.Errorf("ValidateOrderAccess(%q) error = nil, want error", value)
		}
	}
}
