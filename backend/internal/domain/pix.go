package domain

import (
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"
)

// Limites do BR Code (EMV QRCPS-MPM). Não são dados de negócio: a chave e o
// nome fantasia são informados pelo proprietário no painel.
const (
	// O subtemplate 26 do BR Code tem no máximo 99 bytes: 18 do GUI ("0014br.gov.bcb.pix")
	// mais 4 do cabeçalho da chave e, no máximo, 77 caracteres da chave.
	MaxPixKeyLength        = 77
	MaxPixMerchantNameSize = 25
	MaxPixMerchantCitySize = 15
)

// Os quatro tipos de chave aceitos pelo arranjo Pix: CPF/CNPJ (com ou sem
// máscara), e-mail, telefone em E.164 e EVP (UUID). Espaços são aparados antes
// da validação, mas não são aceitos no meio da chave.
var (
	pixDocumentKeyPattern = regexp.MustCompile(`^(?:[0-9]{11}|[0-9]{14}|[0-9]{3}\.[0-9]{3}\.[0-9]{3}-[0-9]{2}|[0-9]{2}\.[0-9]{3}\.[0-9]{3}/[0-9]{4}-[0-9]{2})$`)
	pixEmailKeyPattern    = regexp.MustCompile(`^[A-Za-z0-9._%+-]+@[A-Za-z0-9-]+(?:\.[A-Za-z0-9-]+)*\.[A-Za-z]{2,}$`)
	pixPhoneKeyPattern    = regexp.MustCompile(`^\+[1-9][0-9]{7,14}$`)
	pixEvpKeyPattern      = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
)

// pixTextReplacements espelha a higienização do gerador de BR Code: acentos
// latinos viram letras ASCII; qualquer outro caractere é descartado.
var pixTextReplacements = map[rune]rune{
	'á': 'a', 'à': 'a', 'ã': 'a', 'â': 'a', 'ä': 'a',
	'é': 'e', 'ê': 'e', 'è': 'e', 'ë': 'e',
	'í': 'i', 'ì': 'i', 'î': 'i', 'ï': 'i',
	'ó': 'o', 'ô': 'o', 'õ': 'o', 'ò': 'o', 'ö': 'o',
	'ú': 'u', 'ù': 'u', 'û': 'u', 'ü': 'u',
	'ç': 'c', 'ñ': 'n',
	'Á': 'A', 'À': 'A', 'Ã': 'A', 'Â': 'A', 'Ä': 'A',
	'É': 'E', 'Ê': 'E', 'È': 'E', 'Ë': 'E',
	'Í': 'I', 'Ì': 'I', 'Î': 'I', 'Ï': 'I',
	'Ó': 'O', 'Ô': 'O', 'Õ': 'O', 'Ò': 'O', 'Ö': 'O',
	'Ú': 'U', 'Ù': 'U', 'Û': 'U', 'Ü': 'U',
	'Ç': 'C', 'Ñ': 'N',
}

// PixSettings guarda os dados mínimos para gerar um BR Code estático local.
// A chave é pública por design, mas a auditoria registra apenas se ela está
// preenchida, nunca o valor.
type PixSettings struct {
	Active       bool
	Key          string
	MerchantName string
	MerchantCity string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Validate exige uma configuração completa somente quando o Pix está ativo.
// Assim é possível salvar um rascunho sem expor cobrança inválida ao checkout.
func (s PixSettings) Validate() error {
	if !s.Active {
		return nil
	}
	if err := validatePixKey("key", s.Key); err != nil {
		return err
	}
	if err := validatePixText("merchantName", s.MerchantName, MaxPixMerchantNameSize); err != nil {
		return err
	}
	return validatePixText("merchantCity", s.MerchantCity, MaxPixMerchantCitySize)
}

func validatePixKey(field, value string) error {
	key := strings.TrimSpace(value)
	if key == "" {
		return invalid(field, CodeRequired, "is required")
	}
	if len([]rune(key)) > MaxPixKeyLength {
		return invalid(field, CodeTooLong, fmt.Sprintf("must have at most %d characters", MaxPixKeyLength))
	}
	if !isValidPixKey(key) {
		return invalid(field, CodeInvalid, "must be a valid CPF, CNPJ, email, E.164 phone or EVP key without spaces")
	}
	return nil
}

func isValidPixKey(key string) bool {
	for _, pattern := range []*regexp.Regexp{
		pixDocumentKeyPattern,
		pixEmailKeyPattern,
		pixPhoneKeyPattern,
		pixEvpKeyPattern,
	} {
		if pattern.MatchString(key) {
			return true
		}
	}
	return false
}

// validatePixText conta os caracteres que sobrariam após a mesma higienização
// feita pelo gerador de BR Code. Assim o servidor não aceita uma configuração
// que o checkout não consegue converter em TLV válido.
func validatePixText(field, value string, maxLength int) error {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return invalid(field, CodeRequired, "is required")
	}

	sanitized := strings.Builder{}
	for _, r := range trimmed {
		if unicode.IsControl(r) {
			return invalid(field, CodeInvalid, "must not contain control characters")
		}
		if replacement, ok := pixTextReplacements[r]; ok {
			r = replacement
		}
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == ' ' {
			sanitized.WriteRune(r)
		}
	}

	result := strings.TrimSpace(sanitized.String())
	if result == "" {
		return invalid(field, CodeInvalid, "must contain letters or numbers")
	}
	if len([]rune(result)) > maxLength {
		return invalid(field, CodeTooLong, fmt.Sprintf("must have at most %d characters after sanitization", maxLength))
	}
	return nil
}

// PixSettingsPatch descreve uma atualização parcial da configuração do Pix.
type PixSettingsPatch struct {
	Active       *bool
	Key          *string
	MerchantName *string
	MerchantCity *string
}

// Empty informa se nenhum campo foi enviado.
func (p PixSettingsPatch) Empty() bool {
	return p.Active == nil && p.Key == nil && p.MerchantName == nil && p.MerchantCity == nil
}

// Apply devolve a configuração com os campos enviados substituídos. A chave é
// sempre aparada para evitar divergência entre o que o operador vê e o payload.
func (p PixSettingsPatch) Apply(current PixSettings) PixSettings {
	applyBool(p.Active, &current.Active)
	if p.Key != nil {
		current.Key = strings.TrimSpace(*p.Key)
	}
	applyString(p.MerchantName, &current.MerchantName)
	applyString(p.MerchantCity, &current.MerchantCity)
	return current
}
