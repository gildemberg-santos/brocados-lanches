package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/brocados-lanches/api/internal/domain"
)

type fakeAdminWebhookRepository struct {
	settings  domain.StoreSettings
	getErr    error
	updateErr error
	received  domain.StoreSettings
	audit     domain.AuditLog
	updates   int
	reads     int
}

func (f *fakeAdminWebhookRepository) GetStoreSettings(context.Context) (domain.StoreSettings, error) {
	f.reads++
	return f.settings, f.getErr
}

func (f *fakeAdminWebhookRepository) AdminUpdateWebhookSettings(_ context.Context, settings domain.StoreSettings, audit domain.AuditLog) error {
	if f.updateErr != nil {
		return f.updateErr
	}
	f.received = settings
	f.audit = audit
	f.updates++
	return nil
}

func TestAdminWebhookGetNotificationReadsStoreOnce(t *testing.T) {
	repo := &fakeAdminWebhookRepository{settings: domain.StoreSettings{
		WhatsApp: "+5585992358708",
		Webhook:  domain.WebhookSettings{URL: "http://127.0.0.1:5678/hook", Active: true},
	}}
	service, err := NewAdminWebhookService(repo)
	if err != nil {
		t.Fatalf("NewAdminWebhookService: %v", err)
	}

	notification, err := service.GetNotification(context.Background())
	if err != nil {
		t.Fatalf("GetNotification: %v", err)
	}
	if notification.Settings != repo.settings.Webhook {
		t.Fatalf("configuração inesperada: %+v", notification.Settings)
	}
	if notification.StoreNumber != "+5585992358708" {
		t.Fatalf("storeNumber = %q", notification.StoreNumber)
	}
	if repo.reads != 1 {
		t.Fatalf("leituras do documento = %d, want 1", repo.reads)
	}
}

func TestAdminWebhookGetNotificationPropagatesRepositoryError(t *testing.T) {
	repo := &fakeAdminWebhookRepository{getErr: errors.New("indisponível")}
	service, err := NewAdminWebhookService(repo)
	if err != nil {
		t.Fatalf("NewAdminWebhookService: %v", err)
	}
	if _, err := service.GetNotification(context.Background()); err == nil {
		t.Fatal("esperava erro ao ler configuração de notificação")
	}
}

func TestAdminWebhookUpdatePersistsWithoutLeakingURL(t *testing.T) {
	now := time.Date(2026, 11, 26, 12, 0, 0, 0, time.UTC)
	repo := &fakeAdminWebhookRepository{settings: domain.StoreSettings{
		Name:    "Brocados",
		IsOpen:  true,
		Address: "Rua Principal, 100",
		Webhook: domain.WebhookSettings{URL: "https://antigo.invalid/hook", Active: true},
	}}
	service, err := NewAdminWebhookService(repo)
	if err != nil {
		t.Fatalf("NewAdminWebhookService: %v", err)
	}
	service.now = func() time.Time { return now }

	url := "http://127.0.0.1:5678/hook"
	active := true
	updated, err := service.Update(context.Background(), "owner-1", domain.WebhookSettingsPatch{
		URL: &url, Active: &active,
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.URL != url || !updated.Active {
		t.Fatalf("configuração inesperada: %+v", updated)
	}
	if repo.updates != 1 {
		t.Fatalf("updates = %d", repo.updates)
	}
	if repo.received.Webhook.URL != url {
		t.Fatal("webhook não foi gravado no documento da loja")
	}
	if repo.audit.Entity != domain.AuditEntityWebhook || repo.audit.ActorID != "owner-1" {
		t.Fatalf("auditoria inesperada: %+v", repo.audit)
	}
	if !repo.audit.After["urlConfigured"].(bool) {
		t.Fatal("auditoria deveria registrar que a URL está preenchida")
	}
	for side, snapshot := range map[string]map[string]any{
		"before": repo.audit.Before,
		"after":  repo.audit.After,
	} {
		for key, value := range snapshot {
			if text, ok := value.(string); ok && text == url {
				t.Fatalf("URL vazou na auditoria (%s.%s)", side, key)
			}
		}
	}
}

func TestAdminWebhookValidationErrors(t *testing.T) {
	service, err := NewAdminWebhookService(&fakeAdminWebhookRepository{})
	if err != nil {
		t.Fatalf("NewAdminWebhookService: %v", err)
	}

	if _, err := service.Update(context.Background(), "owner-1", domain.WebhookSettingsPatch{}); err == nil {
		t.Fatal("esperava erro para patch vazio")
	}
	active := true
	if _, err := service.Update(context.Background(), "owner-1", domain.WebhookSettingsPatch{Active: &active}); err == nil {
		t.Fatal("esperava erro ao ativar sem URL")
	}
	if _, err := service.Update(context.Background(), "owner-1", domain.WebhookSettingsPatch{}); err == nil {
		t.Fatal("esperava erro ao atualizar sem alterações")
	}
}

func TestAdminWebhookGetPropagatesRepositoryError(t *testing.T) {
	service, err := NewAdminWebhookService(&fakeAdminWebhookRepository{getErr: errors.New("indisponível")})
	if err != nil {
		t.Fatalf("NewAdminWebhookService: %v", err)
	}
	if _, err := service.Get(context.Background()); err == nil {
		t.Fatal("esperava erro do repositório")
	}
}
