package domain

import (
	"strings"
	"time"
	"unicode/utf8"
)

// Papéis do painel administrativo. A autorização é sempre validada no servidor.
const (
	AdminRoleOwner    = "owner"
	AdminRoleOperator = "operator"
)

// Limites de entrada da autenticação. Não são dados de negócio.
const (
	MaxAdminNameLength     = 120
	MaxAdminEmailLength    = 254
	MinAdminPasswordLength = 12
	MaxAdminPasswordLength = 128
)

// AdminUser é um usuário do painel. PasswordHash nunca sai da camada de serviço.
type AdminUser struct {
	ID           string
	Name         string
	Email        string
	PasswordHash string
	Role         string
	Active       bool
	LastLoginAt  *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Public devolve uma cópia segura para respostas HTTP e logs.
func (u AdminUser) Public() AdminUser {
	return AdminUser{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		Role:      u.Role,
		Active:    u.Active,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

// AdminSession guarda apenas o hash do token opaco; o token em claro vive no cookie.
type AdminSession struct {
	TokenHash         string
	UserID            string
	CreatedAt         time.Time
	ExpiresAt         time.Time
	AbsoluteExpiresAt time.Time
	LastSeenAt        time.Time
	UserAgentHash     string
}

// NormalizeAdminEmail normaliza o e-mail usado como identificador de login.
func NormalizeAdminEmail(email string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(email))
	if normalized == "" || utf8.RuneCountInString(normalized) > MaxAdminEmailLength {
		return "", ErrUnauthorized
	}
	if !strings.Contains(normalized, "@") || strings.HasPrefix(normalized, "@") ||
		strings.HasSuffix(normalized, "@") || strings.ContainsAny(normalized, " \t\r\n") {
		return "", ErrUnauthorized
	}
	return normalized, nil
}

// ValidateAdminPassword aplica a política de senha do ADR 0004.
func ValidateAdminPassword(password string) error {
	length := utf8.RuneCountInString(password)
	if length < MinAdminPasswordLength || length > MaxAdminPasswordLength {
		return &ValidationError{
			Field:   "password",
			Code:    CodeOutOfRange,
			Message: "must have between 12 and 128 characters",
		}
	}
	return nil
}

// ValidateAdminRole aceita apenas os papéis definidos para a V1.
func ValidateAdminRole(role string) error {
	if role != AdminRoleOwner && role != AdminRoleOperator {
		return &ValidationError{Field: "role", Code: CodeInvalid, Message: "invalid admin role"}
	}
	return nil
}

// IsAdminRole informa se o papel é reconhecido pelo sistema.
func IsAdminRole(role string) bool {
	return role == AdminRoleOwner || role == AdminRoleOperator
}
