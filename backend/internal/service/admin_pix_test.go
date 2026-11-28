package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/brocados-lanches/api/internal/domain"
)

type fakeAdminPixRepository struct {
	settings  domain.StoreSettings
	getErr    error
	updateErr error
	received  domain.StoreSettings
	audit     domain.AuditLog
	updates   int
}

func (f *fakeAdminPixRepository) GetStoreSettings(context.Context) (domain.StoreSettings, error) {
	return f.settings, f.getErr
}

func (f *fakeAdminPixRepository) AdminUpdatePixSettings(_ context.Context, settings domain.StoreSettings, audit domain.AuditLog) error {
	if f.updateErr != nil {
		return f.updateErr
	}
	f.received = settings
	f.audit = audit
	f.updates++
	return nil
}

func TestAdminPixUpdatePersistsWithoutLeakingKey(t *testing.T) {
	now := time.Date(2026, 11, 27, 12, 0, 0, 0, time.UTC)
	repo := &fakeAdminPixRepository{settings: domain.StoreSettings{
		Name: "Brocados Lanches",
		Pix: domain.PixSettings{
			Key:          "chave-antiga@example.com",
			MerchantName: "Nome antigo",
			MerchantCity: "Aquiraz",
		},
	}}
	service, err := NewAdminPixService(repo)
	if err != nil {
		t.Fatalf("NewAdminPixService: %v", err)
	}
	service.now = func() time.Time { return now }

	active := true
	key := " 123e4567-e89b-12d3-a456-426614174000 "
	updated, err := service.Update(context.Background(), "owner-1", domain.PixSettingsPatch{
		Active: &active, Key: &key,
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if !updated.Active || updated.Key != "123e4567-e89b-12d3-a456-426614174000" {
		t.Fatalf("configuração inesperada: %+v", updated)
	}
	if updated.MerchantName != "Nome antigo" || updated.MerchantCity != "Aquiraz" {
		t.Fatal("campos não enviados foram alterados")
	}
	if repo.received.Pix.Key != updated.Key || repo.received.UpdatedAt != now {
		t.Fatal("configuração não foi gravada no documento da loja")
	}
	if repo.audit.Entity != domain.AuditEntityPix || repo.audit.ActorID != "owner-1" {
		t.Fatalf("auditoria inesperada: %+v", repo.audit)
	}
	if repo.audit.Before["keyConfigured"] != true || repo.audit.After["keyConfigured"] != true {
		t.Fatal("auditoria deveria registrar apenas se a chave está preenchida")
	}
	for side, snapshot := range map[string]map[string]any{
		"before": repo.audit.Before,
		"after":  repo.audit.After,
	} {
		for key, value := range snapshot {
			if text, ok := value.(string); ok && text == updated.Key {
				t.Fatalf("chave Pix vazou na auditoria (%s.%s)", side, key)
			}
		}
	}
}

func TestAdminPixValidationErrors(t *testing.T) {
	service, err := NewAdminPixService(&fakeAdminPixRepository{})
	if err != nil {
		t.Fatalf("NewAdminPixService: %v", err)
	}

	if _, err := service.Update(context.Background(), "owner-1", domain.PixSettingsPatch{}); err == nil {
		t.Fatal("esperava erro para patch vazio")
	}
	active := true
	if _, err := service.Update(context.Background(), "owner-1", domain.PixSettingsPatch{Active: &active}); err == nil {
		t.Fatal("esperava erro ao ativar sem chave e dados do recebedor")
	}
}

func TestAdminPixGetPropagatesRepositoryError(t *testing.T) {
	service, err := NewAdminPixService(&fakeAdminPixRepository{getErr: errors.New("indisponível")})
	if err != nil {
		t.Fatalf("NewAdminPixService: %v", err)
	}
	if _, err := service.Get(context.Background()); err == nil {
		t.Fatal("esperava erro do repositório")
	}
}

func TestAdminPixUpdatePropagatesRepositoryError(t *testing.T) {
	repo := &fakeAdminPixRepository{
		settings: domain.StoreSettings{Pix: domain.PixSettings{
			Active: true, Key: "123e4567-e89b-12d3-a456-426614174000",
			MerchantName: "Brocados", MerchantCity: "Aquiraz",
		}},
		updateErr: errors.New("indisponível"),
	}
	service, err := NewAdminPixService(repo)
	if err != nil {
		t.Fatalf("NewAdminPixService: %v", err)
	}
	city := "Aquiraz"
	if _, err := service.Update(context.Background(), "owner-1", domain.PixSettingsPatch{MerchantCity: &city}); err == nil {
		t.Fatal("esperava erro do repositório")
	}
}
