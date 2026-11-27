package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/brocados-lanches/api/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

const adminTestActorID = "665f1c2a9b1e4a0012a3b4f9"

func adminTestAudit(action, entity, entityID string) domain.AuditLog {
	return domain.AuditLog{
		ActorID: adminTestActorID, Action: action, Entity: entity, EntityID: entityID,
		Before:    map[string]any{"active": true},
		After:     map[string]any{"active": false},
		CreatedAt: testTime(),
	}
}

func TestEnsureAuditIndexesRequiresDatabase(t *testing.T) {
	if err := EnsureAuditIndexes(context.Background(), nil); err == nil {
		t.Fatal("EnsureAuditIndexes(nil) error = nil, want error")
	}
}

func TestAdminCatalogRepositoryCreateUpdateAndAudit(t *testing.T) {
	ctx := testContext(t)
	database := openCatalogTestDatabase(t)
	if err := EnsureAuditIndexes(ctx, database); err != nil {
		t.Fatalf("EnsureAuditIndexes() error = %v", err)
	}
	repository := newCatalogRepository(t, database)

	id := bson.NewObjectID().Hex()
	category := domain.Category{
		ID: id, Name: "Hambúrgueres", Slug: "hamburgueres",
		Active: true, SortOrder: 1, CreatedAt: testTime(), UpdatedAt: testTime(),
	}
	if err := repository.AdminCreateCategory(ctx, category, adminTestAudit(domain.AuditActionCreate, domain.AuditEntityCategory, id)); err != nil {
		t.Fatalf("AdminCreateCategory() error = %v", err)
	}
	stored, err := repository.GetCategory(ctx, id)
	if err != nil || stored.Name != category.Name {
		t.Fatalf("GetCategory() = %+v, %v", stored, err)
	}

	category.Active = false
	category.UpdatedAt = testTime()
	if err := repository.AdminUpdateCategory(ctx, category, adminTestAudit(domain.AuditActionUpdate, domain.AuditEntityCategory, id)); err != nil {
		t.Fatalf("AdminUpdateCategory() error = %v", err)
	}
	stored, err = repository.GetCategory(ctx, id)
	if err != nil || stored.Active {
		t.Fatalf("GetCategory() after update = %+v, %v", stored, err)
	}

	count, err := database.Collection(CollectionAuditLogs).CountDocuments(ctx, bson.D{{Key: "entityId", Value: id}})
	if err != nil || count != 2 {
		t.Fatalf("audit count = %d, err = %v; want 2", count, err)
	}

	duplicate := domain.Category{
		ID: bson.NewObjectID().Hex(), Name: "Outra", Slug: "hamburgueres",
		Active: true, CreatedAt: testTime(), UpdatedAt: testTime(),
	}
	if err := repository.AdminCreateCategory(ctx, duplicate, adminTestAudit(domain.AuditActionCreate, domain.AuditEntityCategory, duplicate.ID)); !errors.Is(err, domain.ErrConflict) {
		t.Errorf("duplicate slug error = %v, want ErrConflict", err)
	}
}

func TestAdminCatalogRepositoryListFiltersAndPagination(t *testing.T) {
	ctx := testContext(t)
	database := openCatalogTestDatabase(t)
	repository := newCatalogRepository(t, database)

	insertDocument(t, ctx, database.Collection(CollectionCategories),
		categoryDocument{ID: bson.NewObjectID(), Name: "Hambúrgueres", Slug: "hamburgueres", Active: true, SortOrder: 2, CreatedAt: testTime(), UpdatedAt: testTime()},
		categoryDocument{ID: bson.NewObjectID(), Name: "Bebidas", Slug: "bebidas", Active: true, SortOrder: 1, CreatedAt: testTime(), UpdatedAt: testTime()},
		categoryDocument{ID: bson.NewObjectID(), Name: "Sobremesas antigas", Slug: "sobremesas", Active: false, SortOrder: 0, CreatedAt: testTime(), UpdatedAt: testTime()},
	)

	all, total, err := repository.AdminListCategories(ctx, domain.AdminListQuery{Page: 1, PageSize: 2, Sort: "sortOrder", Order: "asc"})
	if err != nil {
		t.Fatalf("AdminListCategories() error = %v", err)
	}
	if total != 3 || len(all) != 2 || all[0].Slug != "sobremesas" {
		t.Fatalf("list = %+v, total = %d", all, total)
	}

	inactive := false
	filtered, total, err := repository.AdminListCategories(ctx, domain.AdminListQuery{Page: 1, PageSize: 10, Active: &inactive, Search: "antigas"})
	if err != nil {
		t.Fatalf("AdminListCategories() filtered error = %v", err)
	}
	if total != 1 || len(filtered) != 1 || filtered[0].Slug != "sobremesas" {
		t.Fatalf("filtered = %+v, total = %d", filtered, total)
	}

	if _, _, err := repository.AdminListCategories(ctx, domain.AdminListQuery{Page: 1, PageSize: 10, Sort: "priceCents"}); err == nil {
		t.Error("unsupported sort error = nil, want validation error")
	}
	if _, _, err := repository.AdminListProducts(ctx, domain.AdminListQuery{Page: 1, PageSize: 10, CategoryID: "invalido"}); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("invalid categoryId error = %v, want ErrNotFound", err)
	}
}

func TestAdminStoreRepositoryUpdateAndAudit(t *testing.T) {
	ctx := testContext(t)
	database := openCatalogTestDatabase(t)
	repository := newCatalogRepository(t, database)

	settings := domain.StoreSettings{
		Name: "Brocados Lanches", IsOpen: true, DeliveryTime: "30-90 min",
		WhatsApp: "+5585992358708", Address: "Rua das Palmeiras, 123, Aquiraz-CE",
		Webhook: domain.WebhookSettings{
			URL: "https://hook.invalid/pedidos", Active: true,
			CreatedAt: testTime(), UpdatedAt: testTime(),
		},
		CreatedAt: testTime(), UpdatedAt: testTime(),
	}
	audit := adminTestAudit(domain.AuditActionUpdate, domain.AuditEntityStore, domain.StoreSettingsID)
	audit.Before = map[string]any{"isOpen": true}
	audit.After = map[string]any{"isOpen": false}
	if err := repository.AdminUpdateStoreSettings(ctx, settings, audit); err != nil {
		t.Fatalf("AdminUpdateStoreSettings() error = %v", err)
	}

	settings.IsOpen = false
	settings.UpdatedAt = testTime()
	if err := repository.AdminUpdateStoreSettings(ctx, settings, audit); err != nil {
		t.Fatalf("AdminUpdateStoreSettings() second call error = %v", err)
	}
	stored, err := repository.GetStoreSettings(ctx)
	if err != nil || stored.IsOpen {
		t.Fatalf("GetStoreSettings() = %+v, %v", stored, err)
	}
	if !stored.CreatedAt.Equal(testTime()) {
		t.Errorf("createdAt = %s, want preserved", stored.CreatedAt)
	}
	// O PATCH da loja não possui o subdocumento da integração.
	if stored.Webhook.URL != "" || stored.Webhook.Active {
		t.Fatalf("webhook after store update = %+v, want zero value", stored.Webhook)
	}

	webhookAudit := adminTestAudit(domain.AuditActionUpdate, domain.AuditEntityWebhook, domain.StoreSettingsID)
	if err := repository.AdminUpdateWebhookSettings(ctx, settings, webhookAudit); err != nil {
		t.Fatalf("AdminUpdateWebhookSettings() error = %v", err)
	}
	stored, err = repository.GetStoreSettings(ctx)
	if err != nil {
		t.Fatalf("GetStoreSettings() after webhook update: %v", err)
	}
	if stored.Webhook.URL != settings.Webhook.URL || !stored.Webhook.Active ||
		!stored.Webhook.CreatedAt.Equal(testTime()) || !stored.Webhook.UpdatedAt.Equal(testTime()) {
		t.Fatalf("webhook round-trip = %+v, want %+v", stored.Webhook, settings.Webhook)
	}
	if stored.IsOpen {
		t.Fatal("atualização do webhook restaurou campo operacional da loja")
	}

	// Um PATCH da loja com valor obsoleto do webhook não pode sobrescrever a
	// configuração gravada pelo owner.
	settings.Webhook = domain.WebhookSettings{URL: "https://antigo.invalid/hook"}
	settings.UpdatedAt = testTime()
	if err := repository.AdminUpdateStoreSettings(ctx, settings, audit); err != nil {
		t.Fatalf("AdminUpdateStoreSettings() with stale webhook error = %v", err)
	}
	stored, err = repository.GetStoreSettings(ctx)
	if err != nil || stored.Webhook.URL != "https://hook.invalid/pedidos" || !stored.Webhook.Active {
		t.Fatalf("webhook sobrescrito por PATCH da loja: %+v, %v", stored.Webhook, err)
	}

	count, err := database.Collection(CollectionAuditLogs).CountDocuments(ctx, bson.D{{Key: "entity", Value: domain.AuditEntityStore}})
	if err != nil || count != 3 {
		t.Fatalf("store audit count = %d, err = %v; want 3", count, err)
	}
	count, err = database.Collection(CollectionAuditLogs).CountDocuments(ctx, bson.D{{Key: "entity", Value: domain.AuditEntityWebhook}})
	if err != nil || count != 1 {
		t.Fatalf("webhook audit count = %d, err = %v; want 1", count, err)
	}
}
