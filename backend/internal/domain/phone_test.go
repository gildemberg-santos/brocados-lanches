package domain

import "testing"

func TestIsValidBrazilianPhone(t *testing.T) {
	tests := []struct {
		name  string
		phone string
		want  bool
	}{
		{"celular nacional", "(85) 99999-9999", true},
		{"celular com DDI", "+55 85 99999-9999", true},
		{"fixo nacional", "(85) 3232-1234", true},
		{"fixo com DDI", "+558532321234", true},
		{"DDD do Rio Grande do Sul", "+55 55 99999-9999", true},
		{"vazio", "", false},
		{"curto", "123", false},
		{"DDI duplicado com celular inválido", "+5555859913655", false},
		{"celular de 9 dígitos começando com 8", "+55 85 89999-9999", false},
		{"celular de 8 dígitos", "+55 85 9999-9999", false},
		{"DDD inexistente", "+55 00 99999-9999", false},
		{"prefixo nacional zero", "(85) 09999-9999", false},
		{"número de outro país", "+1 212 555 0123", false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := IsValidBrazilianPhone(test.phone); got != test.want {
				t.Errorf("IsValidBrazilianPhone(%q) = %v, want %v", test.phone, got, test.want)
			}
		})
	}
}
