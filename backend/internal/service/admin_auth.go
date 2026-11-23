package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/brocados-lanches/api/internal/domain"
	"github.com/brocados-lanches/api/internal/platform/password"
	"github.com/brocados-lanches/api/internal/platform/token"
)

// Constantes da política de sessão do ADR 0004.
const (
	AdminSessionIdleDuration     = 8 * time.Hour
	AdminSessionAbsoluteDuration = 7 * 24 * time.Hour
)

// AdminAuthRepository é a superfície necessária para autenticar usuários do painel.
type AdminAuthRepository interface {
	CountUsers(ctx context.Context) (int64, error)
	CreateUser(ctx context.Context, user domain.AdminUser) error
	FindUserByEmail(ctx context.Context, email string) (domain.AdminUser, bool, error)
	FindUserByID(ctx context.Context, id string) (domain.AdminUser, error)
	RecordLogin(ctx context.Context, userID string, at time.Time, session domain.AdminSession) error
	FindSessionByTokenHash(ctx context.Context, tokenHash string) (domain.AdminSession, error)
	DeleteSessionByTokenHash(ctx context.Context, tokenHash string) error
	TouchSession(ctx context.Context, tokenHash string, expiresAt, lastSeenAt time.Time) error
}

// AdminAuthService aplica as regras de login, sessão e bootstrap do painel.
// O token opaco em claro existe apenas no retorno de Login: o repositório recebe
// somente o hash SHA-256.
type AdminAuthService struct {
	repo           AdminAuthRepository
	now            func() time.Time
	verifyPassword func(encoded, plain string) bool
	newToken       func() (string, error)
	idle           time.Duration
	absolute       time.Duration
}

// NewAdminAuthService valida as dependências e os tempos de expiração.
func NewAdminAuthService(repo AdminAuthRepository, idle, absolute time.Duration) (*AdminAuthService, error) {
	if repo == nil {
		return nil, errors.New("admin auth repository is required")
	}
	if idle <= 0 || absolute <= idle {
		return nil, errors.New("invalid admin session durations")
	}
	return &AdminAuthService{
		repo:           repo,
		now:            time.Now,
		verifyPassword: password.Verify,
		newToken:       token.New,
		idle:           idle,
		absolute:       absolute,
	}, nil
}

// Login valida credenciais, rotaciona as sessões do usuário e devolve o token
// opaco que deve ser gravado somente no cookie HttpOnly.
func (s *AdminAuthService) Login(ctx context.Context, email, plainPassword string) (domain.AdminUser, string, error) {
	normalized, err := domain.NormalizeAdminEmail(email)
	if err != nil {
		return domain.AdminUser{}, "", domain.ErrUnauthorized
	}
	if plainPassword == "" {
		return domain.AdminUser{}, "", domain.ErrUnauthorized
	}

	user, found, err := s.repo.FindUserByEmail(ctx, normalized)
	if err != nil {
		return domain.AdminUser{}, "", fmt.Errorf("find admin user: %w", err)
	}
	if !found || !user.Active {
		// Custo similar ao de uma verificação real para não revelar existência
		// de conta ou status por tempo de resposta.
		_, _ = password.Hash(plainPassword)
		return domain.AdminUser{}, "", domain.ErrUnauthorized
	}
	if !s.verifyPassword(user.PasswordHash, plainPassword) {
		return domain.AdminUser{}, "", domain.ErrUnauthorized
	}

	rawToken, err := s.newToken()
	if err != nil {
		return domain.AdminUser{}, "", fmt.Errorf("generate admin session token: %w", err)
	}

	now := s.now().UTC()
	session := domain.AdminSession{
		TokenHash:         token.Hash(rawToken),
		UserID:            user.ID,
		CreatedAt:         now,
		ExpiresAt:         now.Add(s.idle),
		AbsoluteExpiresAt: now.Add(s.absolute),
		LastSeenAt:        now,
	}
	if err := s.repo.RecordLogin(ctx, user.ID, now, session); err != nil {
		return domain.AdminUser{}, "", fmt.Errorf("record admin login: %w", err)
	}

	user.LastLoginAt = &now
	return user.Public(), rawToken, nil
}

// Authenticate valida o token da sessão, renova a expiração ociosa e devolve o
// usuário autorizado. Falhas sempre viram ErrUnauthorized, sem detalhes internos.
func (s *AdminAuthService) Authenticate(ctx context.Context, rawToken string) (domain.AdminUser, error) {
	if strings.TrimSpace(rawToken) == "" {
		return domain.AdminUser{}, domain.ErrUnauthorized
	}

	tokenHash := token.Hash(rawToken)
	session, err := s.repo.FindSessionByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.AdminUser{}, domain.ErrUnauthorized
		}
		return domain.AdminUser{}, fmt.Errorf("find admin session: %w", err)
	}

	now := s.now().UTC()
	if !now.Before(session.ExpiresAt) || !now.Before(session.AbsoluteExpiresAt) {
		return domain.AdminUser{}, domain.ErrUnauthorized
	}

	user, err := s.repo.FindUserByID(ctx, session.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.AdminUser{}, domain.ErrUnauthorized
		}
		return domain.AdminUser{}, fmt.Errorf("find admin user by session: %w", err)
	}
	if !user.Active {
		return domain.AdminUser{}, domain.ErrUnauthorized
	}

	expiresAt := now.Add(s.idle)
	if expiresAt.After(session.AbsoluteExpiresAt) {
		expiresAt = session.AbsoluteExpiresAt
	}
	if err := s.repo.TouchSession(ctx, tokenHash, expiresAt, now); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.AdminUser{}, domain.ErrUnauthorized
		}
		return domain.AdminUser{}, fmt.Errorf("touch admin session: %w", err)
	}

	return user.Public(), nil
}

// Logout remove a sessão atual. É idempotente: sessão inexistente também é sucesso.
func (s *AdminAuthService) Logout(ctx context.Context, rawToken string) error {
	if strings.TrimSpace(rawToken) == "" {
		return nil
	}
	if err := s.repo.DeleteSessionByTokenHash(ctx, token.Hash(rawToken)); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil
		}
		return fmt.Errorf("delete admin session: %w", err)
	}
	return nil
}

// BootstrapOwner cria o primeiro usuário owner quando não existe nenhum usuário.
// É idempotente e usado apenas na inicialização, nunca em rota HTTP pública.
func (s *AdminAuthService) BootstrapOwner(ctx context.Context, name, email, plainPassword string) (bool, error) {
	name = strings.TrimSpace(name)
	if name == "" || len([]rune(name)) > domain.MaxAdminNameLength {
		return false, &domain.ValidationError{
			Field:   "name",
			Code:    domain.CodeOutOfRange,
			Message: "nome do administrador inválido",
		}
	}
	normalized, err := domain.NormalizeAdminEmail(email)
	if err != nil {
		return false, &domain.ValidationError{
			Field:   "email",
			Code:    domain.CodeInvalid,
			Message: "e-mail do administrador inválido",
		}
	}
	if err := domain.ValidateAdminPassword(plainPassword); err != nil {
		return false, err
	}

	count, err := s.repo.CountUsers(ctx)
	if err != nil {
		return false, fmt.Errorf("count admin users: %w", err)
	}
	if count > 0 {
		return false, nil
	}

	hash, err := password.Hash(plainPassword)
	if err != nil {
		return false, fmt.Errorf("hash bootstrap password: %w", err)
	}
	now := s.now().UTC()
	user := domain.AdminUser{
		Name:         name,
		Email:        normalized,
		PasswordHash: hash,
		Role:         domain.AdminRoleOwner,
		Active:       true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.repo.CreateUser(ctx, user); err != nil {
		if errors.Is(err, domain.ErrConflict) {
			// Outra instância venceu a corrida de bootstrap.
			return false, nil
		}
		return false, fmt.Errorf("create bootstrap owner: %w", err)
	}
	return true, nil
}
