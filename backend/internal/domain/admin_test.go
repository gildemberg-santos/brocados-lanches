package domain

import (
	"testing"
	"time"
)

func TestNormalizeAdminEmail(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
		ok    bool
	}{
		{name: "trim and lowercase", input: "  Operador@Example.COM ", want: "operador@example.com", ok: true},
		{name: "empty", input: "   ", ok: false},
		{name: "missing at", input: "operador.example.com", ok: false},
		{name: "missing local", input: "@example.com", ok: false},
		{name: "missing domain", input: "operador@", ok: false},
		{name: "with space", input: "operador @example.com", ok: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := NormalizeAdminEmail(test.input)
			if test.ok {
				if err != nil {
					t.Fatalf("NormalizeAdminEmail() error = %v", err)
				}
				if got != test.want {
					t.Errorf("NormalizeAdminEmail() = %q, want %q", got, test.want)
				}
				return
			}
			if err == nil {
				t.Fatal("NormalizeAdminEmail() error = nil, want error")
			}
		})
	}
}

func TestValidateAdminPassword(t *testing.T) {
	if err := ValidateAdminPassword("123456789012"); err != nil {
		t.Errorf("12 characters: error = %v, want nil", err)
	}
	for _, value := range []string{"curta", "12345678901", ""} {
		if err := ValidateAdminPassword(value); err == nil {
			t.Errorf("ValidateAdminPassword(%q) error = nil, want error", value)
		}
	}
	if err := ValidateAdminPassword(string(make([]rune, MaxAdminPasswordLength+1))); err == nil {
		t.Error("password above maximum: error = nil, want error")
	}
}

func TestValidateAdminRole(t *testing.T) {
	for _, role := range []string{AdminRoleOwner, AdminRoleOperator} {
		if err := ValidateAdminRole(role); err != nil {
			t.Errorf("ValidateAdminRole(%q) error = %v", role, err)
		}
		if !IsAdminRole(role) {
			t.Errorf("IsAdminRole(%q) = false", role)
		}
	}
	if err := ValidateAdminRole("admin"); err == nil {
		t.Error("ValidateAdminRole(admin) error = nil, want error")
	}
	if IsAdminRole("admin") {
		t.Error("IsAdminRole(admin) = true, want false")
	}
}

func TestAdminUserPublicHidesPasswordHash(t *testing.T) {
	user := AdminUser{
		ID:           "665f1c2a9b1e4a0012a3b4f1",
		Name:         "Operador",
		Email:        "operador@example.com",
		PasswordHash: "hash-secreto",
		Role:         AdminRoleOperator,
		Active:       true,
		CreatedAt:    time.Date(2026, 11, 22, 12, 0, 0, 0, time.UTC),
		UpdatedAt:    time.Date(2026, 11, 22, 13, 0, 0, 0, time.UTC),
	}

	public := user.Public()
	if public.PasswordHash != "" {
		t.Error("Public() leaked password hash")
	}
	if public.ID != user.ID || public.Role != user.Role || public.Email != user.Email {
		t.Errorf("Public() dropped visible fields: %+v", public)
	}
	user.Name = "Alterado"
	if public.Name != "Operador" {
		t.Error("Public() must be a copy")
	}
}
