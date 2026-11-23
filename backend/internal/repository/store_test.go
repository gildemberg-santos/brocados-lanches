package repository

import (
	"strings"
	"testing"

	"github.com/brocados-lanches/api/internal/domain"
)

func TestCatalogRepositoryGetStoreSettings(t *testing.T) {
	ctx := testContext(t)
	database := openCatalogTestDatabase(t)
	repository := newCatalogRepository(t, database)

	document := storeSettingsDocument{
		ID:     domain.StoreSettingsID,
		Name:   "Brocados Lanches",
		IsOpen: true,
		OpeningHours: []openingHoursDocument{{
			Day:       "sunday",
			Intervals: []openingIntervalDocument{{Open: "18:00", Close: "23:30"}},
		}},
		DeliveryTime: "35-50 min",
		CreatedAt:    testTime(),
		UpdatedAt:    testTime(),
	}
	insertDocument(t, ctx, database.Collection(CollectionStoreSettings), document)

	settings, err := repository.GetStoreSettings(ctx)
	if err != nil {
		t.Fatalf("GetStoreSettings() error = %v", err)
	}
	if settings.Name != "Brocados Lanches" || !settings.IsOpen || settings.DeliveryTime != "35-50 min" {
		t.Errorf("settings = %+v", settings)
	}
	if len(settings.OpeningHours) != 1 || len(settings.OpeningHours[0].Intervals) != 1 {
		t.Fatalf("opening hours = %+v", settings.OpeningHours)
	}
	if settings.OpeningHours[0].Day != "sunday" || settings.OpeningHours[0].Intervals[0].Open != "18:00" {
		t.Errorf("opening hours = %+v", settings.OpeningHours[0])
	}
}

func TestCatalogRepositoryGetStoreSettingsMissingDocument(t *testing.T) {
	ctx := testContext(t)
	database := openCatalogTestDatabase(t)
	repository := newCatalogRepository(t, database)

	_, err := repository.GetStoreSettings(ctx)
	if err == nil {
		t.Fatal("GetStoreSettings() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "store settings document is missing") {
		t.Errorf("error = %v, want missing document context", err)
	}
}
