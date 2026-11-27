package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/brocados-lanches/api/internal/domain"
)

// AdminWebhookRepository reaproveita o documento único da loja e a auditoria
// transacional já existente, sem criar outra coleção.
type AdminWebhookRepository interface {
	GetStoreSettings(ctx context.Context) (domain.StoreSettings, error)
	AdminUpdateWebhookSettings(ctx context.Context, settings domain.StoreSettings, audit domain.AuditLog) error
}

// AdminWebhookService valida e persiste a integração de notificações. A URL é
// sensível porque o corpo do webhook contém telefone do cliente: a auditoria
// registra apenas se ela está preenchida, nunca o valor.
type AdminWebhookService struct {
	repo AdminWebhookRepository
	now  func() time.Time
}

// NewAdminWebhookService valida as dependências do serviço de integrações.
func NewAdminWebhookService(repo AdminWebhookRepository) (*AdminWebhookService, error) {
	if repo == nil {
		return nil, errors.New("admin webhook repository is required")
	}
	return &AdminWebhookService{repo: repo, now: time.Now}, nil
}

// Get devolve a configuração atual do webhook.
func (s *AdminWebhookService) Get(ctx context.Context) (domain.WebhookSettings, error) {
	settings, err := s.repo.GetStoreSettings(ctx)
	if err != nil {
		return domain.WebhookSettings{}, fmt.Errorf("get webhook settings: %w", err)
	}
	return settings.Webhook, nil
}

// Update aplica um PATCH parcial, valida o resultado e registra auditoria.
func (s *AdminWebhookService) Update(ctx context.Context, actorID string, patch domain.WebhookSettingsPatch) (domain.WebhookSettings, error) {
	if patch.Empty() {
		return domain.WebhookSettings{}, &domain.ValidationError{Field: "body", Code: domain.CodeRequired, Message: "at least one field is required"}
	}

	current, err := s.repo.GetStoreSettings(ctx)
	if err != nil {
		return domain.WebhookSettings{}, fmt.Errorf("get store settings for webhook update: %w", err)
	}
	before := webhookSnapshot(current.Webhook)
	updated := patch.Apply(current.Webhook)
	if err := updated.Validate(); err != nil {
		return domain.WebhookSettings{}, err
	}

	now := s.now().UTC()
	if updated.CreatedAt.IsZero() {
		updated.CreatedAt = now
	}
	updated.UpdatedAt = now
	current.Webhook = updated
	current.UpdatedAt = now

	audit := domain.AuditLog{
		ActorID:   actorID,
		Action:    domain.AuditActionUpdate,
		Entity:    domain.AuditEntityWebhook,
		EntityID:  domain.StoreSettingsID,
		Before:    before,
		After:     webhookSnapshot(updated),
		CreatedAt: now,
	}
	if err := s.repo.AdminUpdateWebhookSettings(ctx, current, audit); err != nil {
		return domain.WebhookSettings{}, err
	}
	return updated, nil
}

// webhookSnapshot evita gravar URL, telefone ou mensagem no log de auditoria.
func webhookSnapshot(settings domain.WebhookSettings) map[string]any {
	return map[string]any{
		"active":        settings.Active,
		"urlConfigured": settings.URL != "",
	}
}
