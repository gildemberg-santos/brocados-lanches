package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/brocados-lanches/api/internal/domain"
)

// AdminPixRepository reaproveita o documento único da loja e a auditoria
// transacional existente, sem criar outra coleção.
type AdminPixRepository interface {
	GetStoreSettings(ctx context.Context) (domain.StoreSettings, error)
	AdminUpdatePixSettings(ctx context.Context, settings domain.StoreSettings, audit domain.AuditLog) error
}

// AdminPixService valida e persiste a configuração do BR Code estático. A
// auditoria registra apenas se a chave está preenchida, nunca o valor.
type AdminPixService struct {
	repo AdminPixRepository
	now  func() time.Time
}

// NewAdminPixService valida as dependências do serviço de configuração do Pix.
func NewAdminPixService(repo AdminPixRepository) (*AdminPixService, error) {
	if repo == nil {
		return nil, errors.New("admin pix repository is required")
	}
	return &AdminPixService{repo: repo, now: time.Now}, nil
}

// Get devolve a configuração atual do Pix.
func (s *AdminPixService) Get(ctx context.Context) (domain.PixSettings, error) {
	settings, err := s.repo.GetStoreSettings(ctx)
	if err != nil {
		return domain.PixSettings{}, fmt.Errorf("get pix settings: %w", err)
	}
	return settings.Pix, nil
}

// Update aplica um PATCH parcial, valida o resultado e grava a auditoria.
func (s *AdminPixService) Update(ctx context.Context, actorID string, patch domain.PixSettingsPatch) (domain.PixSettings, error) {
	if patch.Empty() {
		return domain.PixSettings{}, &domain.ValidationError{Field: "body", Code: domain.CodeRequired, Message: "at least one field is required"}
	}

	current, err := s.repo.GetStoreSettings(ctx)
	if err != nil {
		return domain.PixSettings{}, fmt.Errorf("get store settings for pix update: %w", err)
	}
	before := pixSnapshot(current.Pix)
	updated := patch.Apply(current.Pix)
	if err := updated.Validate(); err != nil {
		return domain.PixSettings{}, err
	}

	now := s.now().UTC()
	if updated.CreatedAt.IsZero() {
		updated.CreatedAt = now
	}
	updated.UpdatedAt = now
	current.Pix = updated
	current.UpdatedAt = now

	audit := domain.AuditLog{
		ActorID:   actorID,
		Action:    domain.AuditActionUpdate,
		Entity:    domain.AuditEntityPix,
		EntityID:  domain.StoreSettingsID,
		Before:    before,
		After:     pixSnapshot(updated),
		CreatedAt: now,
	}
	if err := s.repo.AdminUpdatePixSettings(ctx, current, audit); err != nil {
		return domain.PixSettings{}, err
	}
	return updated, nil
}

// pixSnapshot evita gravar a chave, o nome e a cidade na auditoria. A chave é
// pública no BR Code, mas não precisa ser acumulada no histórico de mutações.
func pixSnapshot(settings domain.PixSettings) map[string]any {
	return map[string]any{
		"active":         settings.Active,
		"keyConfigured":  settings.Key != "",
		"nameConfigured": settings.MerchantName != "",
		"cityConfigured": settings.MerchantCity != "",
	}
}
