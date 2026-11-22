package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/brocados-lanches/api/internal/domain"
)

// AdminStoreRepository é a superfície necessária para ler e atualizar a
// configuração operacional da loja.
type AdminStoreRepository interface {
	GetStoreSettings(ctx context.Context) (domain.StoreSettings, error)
	AdminUpdateStoreSettings(ctx context.Context, settings domain.StoreSettings, audit domain.AuditLog) error
}

// AdminStoreService aplica validação e auditoria às mudanças do kill switch,
// horários e contato da loja.
type AdminStoreService struct {
	repo AdminStoreRepository
	now  func() time.Time
}

// NewAdminStoreService valida as dependências do serviço de configuração.
func NewAdminStoreService(repo AdminStoreRepository) (*AdminStoreService, error) {
	if repo == nil {
		return nil, errors.New("admin store repository is required")
	}
	return &AdminStoreService{repo: repo, now: time.Now}, nil
}

// Get devolve a configuração completa da loja.
func (s *AdminStoreService) Get(ctx context.Context) (domain.StoreSettings, error) {
	settings, err := s.repo.GetStoreSettings(ctx)
	if err != nil {
		return domain.StoreSettings{}, fmt.Errorf("get store settings: %w", err)
	}
	return settings, nil
}

// Update aplica um PATCH parcial, valida o resultado e grava a auditoria.
func (s *AdminStoreService) Update(ctx context.Context, actorID string, patch domain.StoreSettingsPatch) (domain.StoreSettings, error) {
	if patch.Empty() {
		return domain.StoreSettings{}, &domain.ValidationError{Field: "body", Code: domain.CodeRequired, Message: "at least one field is required"}
	}

	current, err := s.repo.GetStoreSettings(ctx)
	if err != nil {
		return domain.StoreSettings{}, fmt.Errorf("get store settings for update: %w", err)
	}
	updated := patch.Apply(current)
	if err := updated.Validate(); err != nil {
		return domain.StoreSettings{}, err
	}

	now := s.now().UTC()
	updated.UpdatedAt = now
	audit := domain.AuditLog{
		ActorID:   actorID,
		Action:    domain.AuditActionUpdate,
		Entity:    domain.AuditEntityStore,
		EntityID:  domain.StoreSettingsID,
		Before:    storeSnapshot(current),
		After:     storeSnapshot(updated),
		CreatedAt: now,
	}
	if err := s.repo.AdminUpdateStoreSettings(ctx, updated, audit); err != nil {
		return domain.StoreSettings{}, err
	}
	return updated, nil
}

// storeSnapshot exclui endereço e contato: a auditoria precisa provar quem
// alterou o kill switch, sem acumular dados desnecessários.
func storeSnapshot(settings domain.StoreSettings) map[string]any {
	return map[string]any{
		"name":         settings.Name,
		"isOpen":       settings.IsOpen,
		"deliveryTime": settings.DeliveryTime,
	}
}
