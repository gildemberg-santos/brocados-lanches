package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/brocados-lanches/api/internal/domain"
)

type fakeAdminStoreRepository struct {
	settings  domain.StoreSettings
	audit     domain.AuditLog
	getErr    error
	updateErr error
	updates   int
}

func (f *fakeAdminStoreRepository) GetStoreSettings(context.Context) (domain.StoreSettings, error) {
	if f.getErr != nil {
		return domain.StoreSettings{}, f.getErr
	}
	return f.settings, nil
}

func (f *fakeAdminStoreRepository) AdminUpdateStoreSettings(_ context.Context, settings domain.StoreSettings, audit domain.AuditLog) error {
	if f.updateErr != nil {
		return f.updateErr
	}
	f.settings = settings
	f.audit = audit
	f.updates++
	return nil
}

func validStoreForService() domain.StoreSettings {
	return domain.StoreSettings{
		Name: "Brocados Lanches", IsOpen: true, DeliveryTime: "30-90 min",
		WhatsApp: "+5585992358708", Address: "Rua das Palmeiras, 123, Aquiraz-CE",
		CreatedAt: time.Date(2026, 11, 20, 12, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 11, 20, 12, 0, 0, 0, time.UTC),
	}
}

func newTestAdminStoreService(repo AdminStoreRepository) *AdminStoreService {
	service, err := NewAdminStoreService(repo)
	if err != nil {
		panic(err)
	}
	service.now = func() time.Time { return time.Date(2026, 11, 22, 12, 0, 0, 0, time.UTC) }
	return service
}

func TestNewAdminStoreServiceRequiresRepository(t *testing.T) {
	if _, err := NewAdminStoreService(nil); err == nil {
		t.Fatal("NewAdminStoreService(nil) error = nil, want error")
	}
}

func TestAdminStoreGet(t *testing.T) {
	repo := &fakeAdminStoreRepository{settings: validStoreForService()}
	service := newTestAdminStoreService(repo)

	settings, err := service.Get(context.Background())
	if err != nil || settings.Name != repo.settings.Name {
		t.Fatalf("Get() = %+v, %v", settings, err)
	}

	repo.getErr = errors.New("mongo down")
	if _, err := service.Get(context.Background()); err == nil {
		t.Error("Get() error = nil, want wrapped repository error")
	}
}

func TestAdminStoreUpdateAuditsKillSwitch(t *testing.T) {
	repo := &fakeAdminStoreRepository{settings: validStoreForService()}
	service := newTestAdminStoreService(repo)

	isOpen := false
	updated, err := service.Update(context.Background(), testActorID, domain.StoreSettingsPatch{IsOpen: &isOpen})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if updated.IsOpen || updated.Name != "Brocados Lanches" || updated.DeliveryTime != "30-90 min" {
		t.Errorf("updated settings = %+v", updated)
	}
	if !updated.CreatedAt.Equal(repo.settings.CreatedAt) || !updated.UpdatedAt.Equal(time.Date(2026, 11, 22, 12, 0, 0, 0, time.UTC)) {
		t.Errorf("timestamps = created %s, updated %s", updated.CreatedAt, updated.UpdatedAt)
	}
	if repo.updates != 1 {
		t.Fatalf("updates = %d, want 1", repo.updates)
	}
	audit := repo.audit
	if audit.ActorID != testActorID || audit.Action != domain.AuditActionUpdate ||
		audit.Entity != domain.AuditEntityStore || audit.EntityID != domain.StoreSettingsID {
		t.Errorf("audit = %+v", audit)
	}
	if audit.Before["isOpen"] != true || audit.After["isOpen"] != false {
		t.Errorf("audit snapshots = %+v / %+v", audit.Before, audit.After)
	}
	if _, found := audit.After["whatsapp"]; found {
		t.Error("audit snapshot leaked contact data")
	}
}

func TestAdminStoreUpdateRejectsInvalidDataAndEmptyPatch(t *testing.T) {
	repo := &fakeAdminStoreRepository{settings: validStoreForService()}
	service := newTestAdminStoreService(repo)

	if _, err := service.Update(context.Background(), testActorID, domain.StoreSettingsPatch{}); err == nil {
		t.Error("empty patch error = nil, want validation error")
	}
	badPhone := "85992358708"
	if _, err := service.Update(context.Background(), testActorID, domain.StoreSettingsPatch{WhatsApp: &badPhone}); err == nil {
		t.Error("invalid whatsapp error = nil, want validation error")
	}
	if repo.updates != 0 {
		t.Errorf("updates = %d, want 0 after validation failures", repo.updates)
	}

	repo.getErr = errors.New("mongo down")
	if _, err := service.Update(context.Background(), testActorID, domain.StoreSettingsPatch{IsOpen: boolPointer(false)}); err == nil {
		t.Error("Update() error = nil, want repository error")
	}

	repo.getErr = nil
	repo.updateErr = errors.New("write failed")
	if _, err := service.Update(context.Background(), testActorID, domain.StoreSettingsPatch{IsOpen: boolPointer(false)}); err == nil {
		t.Error("Update() error = nil, want write error")
	}
}

func boolPointer(value bool) *bool {
	return &value
}
