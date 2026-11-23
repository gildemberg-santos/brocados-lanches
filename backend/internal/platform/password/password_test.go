package password

import (
	"strings"
	"testing"
)

func TestHashAndVerify(t *testing.T) {
	hash, err := Hash("senha-forte-com-12-caracteres")
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}
	if !strings.HasPrefix(hash, prefix) {
		t.Fatalf("hash = %q, want Argon2id prefix %q", hash, prefix)
	}
	if !Verify(hash, "senha-forte-com-12-caracteres") {
		t.Error("Verify() = false for the original password")
	}
	if Verify(hash, "outra-senha-forte-123") {
		t.Error("Verify() = true for a wrong password")
	}
}

func TestHashUsesRandomSalt(t *testing.T) {
	first, err := Hash("senha-forte-com-12-caracteres")
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}
	second, err := Hash("senha-forte-com-12-caracteres")
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}
	if first == second {
		t.Error("hashes are equal; salt must be random")
	}
}

func TestVerifyRejectsMalformedHash(t *testing.T) {
	for _, hash := range []string{"", "plain", "$argon2id$v=19$m=1,t=1,p=1$c2FsdA$aGFzaA"} {
		if Verify(hash, "senha-forte-com-12-caracteres") {
			t.Errorf("Verify(%q) = true, want false", hash)
		}
	}
}
