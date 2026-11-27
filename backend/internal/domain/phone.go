package domain

import (
	"regexp"
	"strings"
)

// DDDs em uso segundo o plano de numeração da Anatel.
const brazilianAreaCodePattern = `(1[1-9]|2[12478]|3[1-578]|4[1-9]|5[1345]|6[1-9]|7[134579]|8[1-9]|9[1-9])`

var (
	phoneCharactersPattern = regexp.MustCompile(`^\+?[0-9()\s-]+$`)
	// Celular tem 9 dígitos e começa com 9; telefone fixo tem 8 dígitos e
	// começa de 2 a 5.
	brazilianPhonePattern = regexp.MustCompile(`^\+55` + brazilianAreaCodePattern + `(9[0-9]{8}|[2-5][0-9]{7})$`)
)

// IsValidBrazilianPhone aceita número nacional (10 ou 11 dígitos) ou E.164
// com DDI 55, desde que o DDD exista e o número seja possível no Brasil.
// Números com DDI repetido ou celular de 9 dígitos começando com 8 são
// rejeitados antes de virarem destinatário de notificação.
func IsValidBrazilianPhone(value string) bool {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" || !phoneCharactersPattern.MatchString(trimmed) {
		return false
	}
	return brazilianPhonePattern.MatchString("+" + NormalizePhoneKey(trimmed))
}
