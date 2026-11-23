package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/brocados-lanches/api/internal/domain"
	"github.com/brocados-lanches/api/internal/platform/password"
)

// AdminUsersRepository reúne as operações de gestão de usuários e leitura de
// auditoria do painel.
type AdminUsersRepository interface {
	CountUsers(ctx context.Context) (int64, error)
	AdminCreateUser(ctx context.Context, user domain.AdminUser, audit domain.AuditLog) error
	FindUserByEmail(ctx context.Context, email string) (domain.AdminUser, bool, error)
	FindUserByID(ctx context.Context, id string) (domain.AdminUser, error)
	ListAdminUsers(ctx context.Context, query domain.AdminListQuery) ([]domain.AdminUser, int64, error)
	CountActiveOwners(ctx context.Context) (int64, error)
	UpdateAdminUser(ctx context.Context, user domain.AdminUser, expectedUpdatedAt time.Time, audit domain.AuditLog) error
	ListAuditLogs(ctx context.Context, query domain.AuditListQuery) ([]domain.AuditLog, int64, error)
}

// AdminUsersService aplica as regras de criação/edição de usuários e expõe a
// auditoria somente leitura.
type AdminUsersService struct {
	repo         AdminUsersRepository
	now          func() time.Time
	hashPassword func(string) (string, error)
	newID        func() (string, error)
}

// NewAdminUsersService valida as dependências do serviço de usuários.
func NewAdminUsersService(repo AdminUsersRepository) (*AdminUsersService, error) {
	if repo == nil {
		return nil, errors.New("admin users repository is required")
	}
	return &AdminUsersService{
		repo:         repo,
		now:          time.Now,
		hashPassword: password.Hash,
		newID:        newObjectIDHex,
	}, nil
}

var adminUserSortFields = []string{"name", "email", "role", "active", "createdAt", "updatedAt"}

// List aplica allowlist de ordenação e devolve usuários sem passwordHash.
func (s *AdminUsersService) List(ctx context.Context, query domain.AdminListQuery) (AdminPage[domain.AdminUser], error) {
	if query.Sort == "" {
		query.Sort = "name"
	}
	if query.Order == "" {
		query.Order = "asc"
	}
	normalized, err := normalizeAdminListQuery(query, adminUserSortFields)
	if err != nil {
		return AdminPage[domain.AdminUser]{}, err
	}
	items, total, err := s.repo.ListAdminUsers(ctx, normalized)
	if err != nil {
		return AdminPage[domain.AdminUser]{}, fmt.Errorf("list admin users: %w", err)
	}
	return newAdminPage(items, normalized, total), nil
}

// Create valida o novo usuário, gera a senha temporária e registra auditoria.
func (s *AdminUsersService) Create(ctx context.Context, actorID string, input domain.AdminUser, passwordPlain string) (domain.AdminUser, error) {
	if err := validateAdminName(input.Name); err != nil {
		return domain.AdminUser{}, err
	}
	email, err := normalizeAdminUserEmail(input.Email)
	if err != nil {
		return domain.AdminUser{}, err
	}
	if err := domain.ValidateAdminRole(input.Role); err != nil {
		return domain.AdminUser{}, err
	}
	if err := domain.ValidateAdminPassword(passwordPlain); err != nil {
		return domain.AdminUser{}, err
	}

	hash, err := s.hashPassword(passwordPlain)
	if err != nil {
		return domain.AdminUser{}, fmt.Errorf("hash admin password: %w", err)
	}
	id, err := s.newID()
	if err != nil {
		return domain.AdminUser{}, fmt.Errorf("generate admin user id: %w", err)
	}
	now := s.now().UTC()
	user := domain.AdminUser{
		ID: id, Name: strings.TrimSpace(input.Name), Email: email,
		PasswordHash: hash, Role: input.Role, Active: true,
		CreatedAt: now, UpdatedAt: now,
	}
	audit := s.newAudit(actorID, domain.AuditActionCreate, id, nil, adminUserSnapshot(user), now)
	if err := s.repo.AdminCreateUser(ctx, user, audit); err != nil {
		return domain.AdminUser{}, err
	}
	return user.Public(), nil
}

// Update aplica um PATCH protegendo o último owner ativo e audita a mudança.
func (s *AdminUsersService) Update(ctx context.Context, actorID, id string, patch domain.AdminUserPatch) (domain.AdminUser, error) {
	if patch.Empty() {
		return domain.AdminUser{}, &domain.ValidationError{Field: "body", Code: domain.CodeRequired, Message: "at least one field is required"}
	}
	current, err := s.repo.FindUserByID(ctx, id)
	if err != nil {
		return domain.AdminUser{}, err
	}
	updated := current
	if patch.Name != nil {
		if err := validateAdminName(*patch.Name); err != nil {
			return domain.AdminUser{}, err
		}
		updated.Name = strings.TrimSpace(*patch.Name)
	}
	if patch.Role != nil {
		if err := domain.ValidateAdminRole(*patch.Role); err != nil {
			return domain.AdminUser{}, err
		}
		updated.Role = *patch.Role
	}
	if patch.Active != nil {
		updated.Active = *patch.Active
	}
	if current.Role == domain.AdminRoleOwner && current.Active &&
		(updated.Role != domain.AdminRoleOwner || !updated.Active) {
		owners, err := s.repo.CountActiveOwners(ctx)
		if err != nil {
			return domain.AdminUser{}, fmt.Errorf("count active owners: %w", err)
		}
		if owners <= 1 {
			return domain.AdminUser{}, domain.ErrConflict
		}
	}
	now := s.now().UTC()
	updated.UpdatedAt = now
	audit := s.newAudit(actorID, domain.AuditActionUpdate, id, adminUserSnapshot(current), adminUserSnapshot(updated), now)
	if err := s.repo.UpdateAdminUser(ctx, updated, current.UpdatedAt, audit); err != nil {
		return domain.AdminUser{}, err
	}
	return updated.Public(), nil
}

// AuditList devolve uma página de auditoria sem PII.
func (s *AdminUsersService) AuditList(ctx context.Context, query domain.AuditListQuery) (AdminPage[domain.AuditLog], error) {
	if query.Sort == "" {
		query.Sort = "createdAt"
	}
	if query.Order == "" {
		query.Order = "desc"
	}
	normalized, err := normalizeAdminListQuery(query.AdminListQuery, []string{"createdAt"})
	if err != nil {
		return AdminPage[domain.AuditLog]{}, err
	}
	query.AdminListQuery = normalized
	items, total, err := s.repo.ListAuditLogs(ctx, query)
	if err != nil {
		return AdminPage[domain.AuditLog]{}, fmt.Errorf("list audit logs: %w", err)
	}
	return newAdminPage(items, normalized, total), nil
}

func (s *AdminUsersService) newAudit(actorID, action, entityID string, before, after map[string]any, at time.Time) domain.AuditLog {
	return domain.AuditLog{
		ActorID: actorID, Action: action, Entity: domain.AuditEntityAdminUser,
		EntityID: entityID, Before: before, After: after, CreatedAt: at,
	}
}

func adminUserSnapshot(user domain.AdminUser) map[string]any {
	return map[string]any{
		"name": user.Name, "email": user.Email, "role": user.Role, "active": user.Active,
	}
}

func normalizeAdminUserEmail(email string) (string, error) {
	normalized, err := domain.NormalizeAdminEmail(email)
	if err != nil {
		return "", &domain.ValidationError{Field: "email", Code: domain.CodeInvalid, Message: "invalid email"}
	}
	return normalized, nil
}

func validateAdminName(name string) error {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" || len([]rune(trimmed)) > domain.MaxAdminNameLength {
		return &domain.ValidationError{Field: "name", Code: domain.CodeInvalid, Message: "invalid admin name"}
	}
	return nil
}
