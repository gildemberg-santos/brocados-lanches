package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/brocados-lanches/api/internal/domain"
)

type fakeStoreRepository struct {
	settings domain.StoreSettings
	err      error
}

func (f *fakeStoreRepository) GetStoreSettings(context.Context) (domain.StoreSettings, error) {
	return f.settings, f.err
}

func newTestStoreService(repo *fakeStoreRepository, location *time.Location) *StoreService {
	service, err := NewStoreService(repo, location)
	if err != nil {
		panic(err)
	}
	service.now = func() time.Time { return time.Date(2026, 11, 22, 22, 30, 0, 0, time.UTC) }
	return service
}

func TestNewStoreServiceValidatesDependencies(t *testing.T) {
	location := time.FixedZone("UTC-3", -3*60*60)
	if _, err := NewStoreService(nil, location); err == nil {
		t.Error("NewStoreService(nil, location) error = nil, want error")
	}
	if _, err := NewStoreService(&fakeStoreRepository{}, nil); err == nil {
		t.Error("NewStoreService(repo, nil) error = nil, want error")
	}
}

func TestStoreServiceGetCalculatesEffectiveStatus(t *testing.T) {
	location := time.FixedZone("UTC-3", -3*60*60)
	repo := &fakeStoreRepository{settings: domain.StoreSettings{
		Name: "Brocados Lanches", IsOpen: true, DeliveryTime: "30-90 min",
		WhatsApp: "+5585992358708", Address: "Rua das Palmeiras, 123, Aquiraz-CE",
		OpeningHours: []domain.OpeningHours{{
			Day: "sunday", Intervals: []domain.OpeningInterval{{Open: "18:00", Close: "23:30"}},
		}},
	}}
	service := newTestStoreService(repo, location)

	view, err := service.Get(context.Background())
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if view.Status != "open" {
		t.Errorf("Status = %q, want open", view.Status)
	}
	if view.NextOpeningAt == nil {
		t.Fatal("NextOpeningAt = nil, want next opening even while open")
	}
	if got := view.NextOpeningAt.Format(time.RFC3339); got != "2026-11-29T21:00:00Z" {
		t.Errorf("NextOpeningAt = %s, want 2026-11-29T21:00:00Z", got)
	}

	repo.settings.IsOpen = false
	view, err = service.Get(context.Background())
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if view.Status != "closed" || view.NextOpeningAt != nil {
		t.Errorf("view = %+v, want closed without next opening", view)
	}

	repo.settings.IsOpen = true
	repo.settings.OpeningHours = []domain.OpeningHours{{
		Day: "sunday", Intervals: []domain.OpeningInterval{{Open: "23:00", Close: "23:30"}},
	}}
	view, err = service.Get(context.Background())
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if view.Status != "closed" || view.NextOpeningAt == nil {
		t.Fatalf("view = %+v, want closed with next opening", view)
	}
	if got := view.NextOpeningAt.Format(time.RFC3339); got != "2026-11-23T02:00:00Z" {
		t.Errorf("NextOpeningAt = %s, want 2026-11-23T02:00:00Z", got)
	}
}

func TestStoreServiceGetWrapsRepositoryError(t *testing.T) {
	location := time.FixedZone("UTC-3", -3*60*60)
	repo := &fakeStoreRepository{err: errors.New("mongo down")}
	service := newTestStoreService(repo, location)

	if _, err := service.Get(context.Background()); err == nil {
		t.Error("Get() error = nil, want repository error")
	}
}
