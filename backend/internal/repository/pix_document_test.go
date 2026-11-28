package repository

import (
	"testing"
	"time"

	"github.com/brocados-lanches/api/internal/domain"
)

func TestStoreSettingsDocumentRoundTripsPix(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 11, 27, 10, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2026, 11, 27, 12, 0, 0, 0, time.UTC)
	settings := domain.StoreSettings{
		Name: "Brocados Lanches",
		Pix: domain.PixSettings{
			Active:       true,
			Key:          "123e4567-e89b-12d3-a456-426614174000",
			MerchantName: "Brocados Lanches",
			MerchantCity: "Aquiraz",
			CreatedAt:    createdAt,
			UpdatedAt:    updatedAt,
		},
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}

	document := newStoreSettingsDocument(settings)
	if document.Pix.Key != settings.Pix.Key || document.Pix.MerchantCity != settings.Pix.MerchantCity {
		t.Fatalf("pix document = %+v", document.Pix)
	}
	if !document.Pix.Active || !document.Pix.CreatedAt.Equal(createdAt) || !document.Pix.UpdatedAt.Equal(updatedAt) {
		t.Fatalf("pix timestamps = %+v", document.Pix)
	}

	restored := document.toDomain()
	if restored.Pix != settings.Pix {
		t.Errorf("pix restaurado = %+v, want %+v", restored.Pix, settings.Pix)
	}
}

func TestStoreSettingsDocumentLegacyWithoutPix(t *testing.T) {
	t.Parallel()

	settings := storeSettingsDocument{ID: domain.StoreSettingsID}.toDomain()
	if settings.Pix.Active || settings.Pix.Key != "" || settings.Pix.MerchantName != "" {
		t.Fatalf("configuração legada virou Pix ativo: %+v", settings.Pix)
	}
}
