package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/brocados-lanches/api/internal/domain"
)

type fakeAdminCatalogRepository struct {
	categories map[string]domain.Category
	products   map[string]domain.Product
	addons     map[string]domain.Addon
	zones      map[string]domain.DeliveryZone
	audits     []domain.AuditLog
	lastQuery  domain.AdminListQuery
	listItems  int
	listTotal  int64
	getErr     error
	listErr    error
	writeErr   error
}

func newFakeAdminCatalogRepository() *fakeAdminCatalogRepository {
	return &fakeAdminCatalogRepository{
		categories: make(map[string]domain.Category),
		products:   make(map[string]domain.Product),
		addons:     make(map[string]domain.Addon),
		zones:      make(map[string]domain.DeliveryZone),
	}
}

func (f *fakeAdminCatalogRepository) GetCategory(_ context.Context, id string) (domain.Category, error) {
	if f.getErr != nil {
		return domain.Category{}, f.getErr
	}
	entity, ok := f.categories[id]
	if !ok {
		return domain.Category{}, domain.ErrNotFound
	}
	return entity, nil
}

func (f *fakeAdminCatalogRepository) GetProduct(_ context.Context, id string) (domain.Product, error) {
	if f.getErr != nil {
		return domain.Product{}, f.getErr
	}
	entity, ok := f.products[id]
	if !ok {
		return domain.Product{}, domain.ErrNotFound
	}
	return entity, nil
}

func (f *fakeAdminCatalogRepository) GetAddon(_ context.Context, id string) (domain.Addon, error) {
	if f.getErr != nil {
		return domain.Addon{}, f.getErr
	}
	entity, ok := f.addons[id]
	if !ok {
		return domain.Addon{}, domain.ErrNotFound
	}
	return entity, nil
}

func (f *fakeAdminCatalogRepository) GetDeliveryZone(_ context.Context, id string) (domain.DeliveryZone, error) {
	if f.getErr != nil {
		return domain.DeliveryZone{}, f.getErr
	}
	entity, ok := f.zones[id]
	if !ok {
		return domain.DeliveryZone{}, domain.ErrNotFound
	}
	return entity, nil
}

func (f *fakeAdminCatalogRepository) AdminListCategories(_ context.Context, query domain.AdminListQuery) ([]domain.Category, int64, error) {
	f.lastQuery = query
	if f.listErr != nil {
		return nil, 0, f.listErr
	}
	return make([]domain.Category, f.listItems), f.listTotal, nil
}

func (f *fakeAdminCatalogRepository) AdminListProducts(_ context.Context, query domain.AdminListQuery) ([]domain.Product, int64, error) {
	f.lastQuery = query
	if f.listErr != nil {
		return nil, 0, f.listErr
	}
	return make([]domain.Product, f.listItems), f.listTotal, nil
}

func (f *fakeAdminCatalogRepository) AdminListAddons(_ context.Context, query domain.AdminListQuery) ([]domain.Addon, int64, error) {
	f.lastQuery = query
	if f.listErr != nil {
		return nil, 0, f.listErr
	}
	return make([]domain.Addon, f.listItems), f.listTotal, nil
}

func (f *fakeAdminCatalogRepository) AdminListDeliveryZones(_ context.Context, query domain.AdminListQuery) ([]domain.DeliveryZone, int64, error) {
	f.lastQuery = query
	if f.listErr != nil {
		return nil, 0, f.listErr
	}
	return make([]domain.DeliveryZone, f.listItems), f.listTotal, nil
}

func (f *fakeAdminCatalogRepository) AdminCreateCategory(_ context.Context, category domain.Category, audit domain.AuditLog) error {
	if f.writeErr != nil {
		return f.writeErr
	}
	f.categories[category.ID] = category
	f.audits = append(f.audits, audit)
	return nil
}

func (f *fakeAdminCatalogRepository) AdminUpdateCategory(_ context.Context, category domain.Category, audit domain.AuditLog) error {
	if f.writeErr != nil {
		return f.writeErr
	}
	f.categories[category.ID] = category
	f.audits = append(f.audits, audit)
	return nil
}

func (f *fakeAdminCatalogRepository) AdminCreateProduct(_ context.Context, product domain.Product, audit domain.AuditLog) error {
	if f.writeErr != nil {
		return f.writeErr
	}
	f.products[product.ID] = product
	f.audits = append(f.audits, audit)
	return nil
}

func (f *fakeAdminCatalogRepository) AdminUpdateProduct(_ context.Context, product domain.Product, audit domain.AuditLog) error {
	if f.writeErr != nil {
		return f.writeErr
	}
	f.products[product.ID] = product
	f.audits = append(f.audits, audit)
	return nil
}

func (f *fakeAdminCatalogRepository) AdminCreateAddon(_ context.Context, addon domain.Addon, audit domain.AuditLog) error {
	if f.writeErr != nil {
		return f.writeErr
	}
	f.addons[addon.ID] = addon
	f.audits = append(f.audits, audit)
	return nil
}

func (f *fakeAdminCatalogRepository) AdminUpdateAddon(_ context.Context, addon domain.Addon, audit domain.AuditLog) error {
	if f.writeErr != nil {
		return f.writeErr
	}
	f.addons[addon.ID] = addon
	f.audits = append(f.audits, audit)
	return nil
}

func (f *fakeAdminCatalogRepository) AdminCreateDeliveryZone(_ context.Context, zone domain.DeliveryZone, audit domain.AuditLog) error {
	if f.writeErr != nil {
		return f.writeErr
	}
	f.zones[zone.ID] = zone
	f.audits = append(f.audits, audit)
	return nil
}

func (f *fakeAdminCatalogRepository) AdminUpdateDeliveryZone(_ context.Context, zone domain.DeliveryZone, audit domain.AuditLog) error {
	if f.writeErr != nil {
		return f.writeErr
	}
	f.zones[zone.ID] = zone
	f.audits = append(f.audits, audit)
	return nil
}

func newTestAdminCatalogService(repo AdminCatalogRepository) *AdminCatalogService {
	service, err := NewAdminCatalogService(repo)
	if err != nil {
		panic(err)
	}
	service.now = func() time.Time { return time.Date(2026, 11, 22, 12, 0, 0, 0, time.UTC) }
	service.newID = func() (string, error) { return "665f1c2a9b1e4a0012a3b4f1", nil }
	return service
}

const testActorID = "665f1c2a9b1e4a0012a3b4f9"

func TestNewAdminCatalogServiceRequiresRepository(t *testing.T) {
	if _, err := NewAdminCatalogService(nil); err == nil {
		t.Fatal("NewAdminCatalogService(nil) error = nil, want error")
	}
}

func TestAdminCatalogCreateAuditsAndSetsTimestamps(t *testing.T) {
	repo := newFakeAdminCatalogRepository()
	service := newTestAdminCatalogService(repo)

	created, err := service.CreateCategory(context.Background(), testActorID, domain.Category{Name: "Hambúrgueres", Slug: "hamburgueres", Active: true})
	if err != nil {
		t.Fatalf("CreateCategory() error = %v", err)
	}
	if created.ID != "665f1c2a9b1e4a0012a3b4f1" || created.CreatedAt.IsZero() || created.UpdatedAt.IsZero() {
		t.Errorf("created category = %+v", created)
	}
	if len(repo.audits) != 1 {
		t.Fatalf("audits = %d, want 1", len(repo.audits))
	}
	audit := repo.audits[0]
	if audit.ActorID != testActorID || audit.Action != domain.AuditActionCreate ||
		audit.Entity != domain.AuditEntityCategory || audit.EntityID != created.ID || audit.Before != nil {
		t.Errorf("audit = %+v", audit)
	}
	if audit.After["slug"] != "hamburgueres" {
		t.Errorf("audit after = %+v", audit.After)
	}

	if _, err := service.CreateCategory(context.Background(), testActorID, domain.Category{Name: "", Slug: " invalido "}); err == nil {
		t.Error("CreateCategory() with invalid category error = nil, want validation error")
	}
	if len(repo.audits) != 1 {
		t.Error("invalid create must not be audited")
	}

	repo.writeErr = domain.ErrConflict
	if _, err := service.CreateCategory(context.Background(), testActorID, domain.Category{Name: "Bebidas", Slug: "bebidas", Active: true}); !errors.Is(err, domain.ErrConflict) {
		t.Errorf("CreateCategory() error = %v, want ErrConflict", err)
	}
}

func TestAdminCatalogUpdatePatchesAndAudits(t *testing.T) {
	repo := newFakeAdminCatalogRepository()
	service := newTestAdminCatalogService(repo)
	id := "665f1c2a9b1e4a0012a3b4f1"
	repo.categories[id] = domain.Category{ID: id, Name: "Hambúrgueres", Slug: "hamburgueres", Active: true, SortOrder: 2}

	active := false
	updated, err := service.UpdateCategory(context.Background(), testActorID, id, domain.CategoryPatch{Active: &active})
	if err != nil {
		t.Fatalf("UpdateCategory() error = %v", err)
	}
	if updated.Active || updated.Name != "Hambúrgueres" || updated.SortOrder != 2 {
		t.Errorf("updated category = %+v", updated)
	}
	if len(repo.audits) != 1 {
		t.Fatalf("audits = %d, want 1", len(repo.audits))
	}
	audit := repo.audits[0]
	if audit.Action != domain.AuditActionUpdate || audit.Before["active"] != true || audit.After["active"] != false {
		t.Errorf("audit = %+v", audit)
	}

	if _, err := service.UpdateCategory(context.Background(), testActorID, id, domain.CategoryPatch{}); err == nil {
		t.Error("empty patch error = nil, want validation error")
	}
	if _, err := service.UpdateCategory(context.Background(), testActorID, "missing", domain.CategoryPatch{Active: &active}); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("missing category error = %v, want ErrNotFound", err)
	}

	name := ""
	if _, err := service.UpdateCategory(context.Background(), testActorID, id, domain.CategoryPatch{Name: &name}); err == nil {
		t.Error("invalid merged category error = nil, want validation error")
	}
	if len(repo.audits) != 1 {
		t.Error("failed update must not be audited")
	}
}

func TestAdminCatalogCreateAndUpdateAllEntities(t *testing.T) {
	ctx := context.Background()
	repo := newFakeAdminCatalogRepository()
	service := newTestAdminCatalogService(repo)

	product, err := service.CreateProduct(ctx, testActorID, domain.Product{
		CategoryID: "665f1c2a9b1e4a0012a3b4f1", Name: "Hambúrguer", Slug: "hamburguer",
		Description: "Descrição", PriceCents: 2300, Available: true, Active: true,
	})
	if err != nil {
		t.Fatalf("CreateProduct() error = %v", err)
	}
	price := int64(2500)
	updatedProduct, err := service.UpdateProduct(ctx, testActorID, product.ID, domain.ProductPatch{PriceCents: &price})
	if err != nil {
		t.Fatalf("UpdateProduct() error = %v", err)
	}
	if updatedProduct.PriceCents != price {
		t.Errorf("product price = %d, want %d", updatedProduct.PriceCents, price)
	}

	addon, err := service.CreateAddon(ctx, testActorID, domain.Addon{Name: "Queijo", Slug: "queijo", PriceCents: 300, Active: true})
	if err != nil {
		t.Fatalf("CreateAddon() error = %v", err)
	}
	addonActive := false
	if _, err := service.UpdateAddon(ctx, testActorID, addon.ID, domain.AddonPatch{Active: &addonActive}); err != nil {
		t.Fatalf("UpdateAddon() error = %v", err)
	}

	zone, err := service.CreateDeliveryZone(ctx, testActorID, domain.DeliveryZone{Name: "Aquiraz", Slug: "aquiraz", FeeCents: 500, Active: true})
	if err != nil {
		t.Fatalf("CreateDeliveryZone() error = %v", err)
	}
	fee := int64(700)
	if _, err := service.UpdateDeliveryZone(ctx, testActorID, zone.ID, domain.DeliveryZonePatch{FeeCents: &fee}); err != nil {
		t.Fatalf("UpdateDeliveryZone() error = %v", err)
	}

	if len(repo.audits) != 6 {
		t.Fatalf("audits = %d, want 6", len(repo.audits))
	}
	entities := map[string]int{}
	for _, audit := range repo.audits {
		entities[audit.Entity]++
	}
	for _, entity := range []string{domain.AuditEntityProduct, domain.AuditEntityAddon, domain.AuditEntityDeliveryZone} {
		if entities[entity] != 2 {
			t.Errorf("audits for %s = %d, want 2", entity, entities[entity])
		}
	}
}

func TestAdminCatalogProductRejectsMalformedReferences(t *testing.T) {
	ctx := context.Background()
	repo := newFakeAdminCatalogRepository()
	service := newTestAdminCatalogService(repo)
	validID := "665f1c2a9b1e4a0012a3b4f1"

	tests := []struct {
		name    string
		product domain.Product
		field   string
	}{
		{
			name:    "categoryId curto",
			product: domain.Product{CategoryID: "curto", Name: "Hambúrguer", Slug: "hamburguer", Description: "Descrição", PriceCents: 2300, Available: true, Active: true},
			field:   "categoryId",
		},
		{
			name:    "categoryId não hexadecimal",
			product: domain.Product{CategoryID: "zzzzzzzzzzzzzzzzzzzzzzzz", Name: "Hambúrguer", Slug: "hamburguer", Description: "Descrição", PriceCents: 2300, Available: true, Active: true},
			field:   "categoryId",
		},
		{
			name: "addonId malformado",
			product: domain.Product{
				CategoryID: validID, Name: "Hambúrguer", Slug: "hamburguer",
				Description: "Descrição", PriceCents: 2300, Available: true, Active: true,
				AddonIDs: []string{"curto"},
			},
			field: "addonIds[0]",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := service.CreateProduct(ctx, testActorID, test.product)
			var validationError *domain.ValidationError
			if !errors.As(err, &validationError) || validationError.Field != test.field {
				t.Fatalf("error = %v, want validation error for %q", err, test.field)
			}
		})
	}
	if len(repo.audits) != 0 {
		t.Errorf("audits = %d, want 0 after validation failures", len(repo.audits))
	}

	// O PATCH precisa validar o resultado depois do merge.
	repo.products[validID] = domain.Product{
		ID: validID, CategoryID: validID, Name: "Hambúrguer", Slug: "hamburguer",
		Description: "Descrição", PriceCents: 2300, Available: true, Active: true,
	}
	badCategory := "curto"
	_, err := service.UpdateProduct(ctx, testActorID, validID, domain.ProductPatch{CategoryID: &badCategory})
	var validationError *domain.ValidationError
	if !errors.As(err, &validationError) || validationError.Field != "categoryId" {
		t.Fatalf("UpdateProduct() error = %v, want validation error for categoryId", err)
	}
	if len(repo.audits) != 0 {
		t.Errorf("audits = %d, want 0 after failed update", len(repo.audits))
	}
}

func TestAdminCatalogListNormalizesQueryAndPaginates(t *testing.T) {
	repo := newFakeAdminCatalogRepository()
	repo.listTotal = 42
	service := newTestAdminCatalogService(repo)

	page, err := service.ListCategories(context.Background(), domain.AdminListQuery{})
	if err != nil {
		t.Fatalf("ListCategories() error = %v", err)
	}
	if page.Page != 1 || page.PageSize != 20 || page.TotalPages != 3 || page.Total != 42 {
		t.Errorf("page metadata = %+v", page)
	}
	if repo.lastQuery.Sort != "sortOrder" || repo.lastQuery.Order != "asc" {
		t.Errorf("normalized query = %+v", repo.lastQuery)
	}
	if page.Items == nil {
		t.Error("empty items must be a non-nil slice")
	}

	active := true
	query := domain.AdminListQuery{Page: 2, PageSize: 10, Sort: "name", Order: "desc", Active: &active, Search: "  hambúrguer "}
	if _, err := service.ListCategories(context.Background(), query); err != nil {
		t.Fatalf("ListCategories() with filters error = %v", err)
	}
	if repo.lastQuery.Page != 2 || repo.lastQuery.PageSize != 10 || repo.lastQuery.Search != "  hambúrguer " || repo.lastQuery.Active == nil || !*repo.lastQuery.Active {
		t.Errorf("query passed to repository = %+v", repo.lastQuery)
	}
}

func TestAdminCatalogListRejectsInvalidQuery(t *testing.T) {
	service := newTestAdminCatalogService(newFakeAdminCatalogRepository())
	tests := []struct {
		name  string
		query domain.AdminListQuery
		field string
	}{
		{name: "página zero negativa", query: domain.AdminListQuery{Page: -1}, field: "page"},
		{name: "pageSize acima do limite", query: domain.AdminListQuery{PageSize: 101}, field: "pageSize"},
		{name: "ordenação inválida", query: domain.AdminListQuery{Sort: "priceCents"}, field: "sort"},
		{name: "direção inválida", query: domain.AdminListQuery{Order: "sideways"}, field: "order"},
		{name: "busca longa", query: domain.AdminListQuery{Search: strings.Repeat("a", 121)}, field: "q"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := service.ListCategories(context.Background(), test.query)
			var validationError *domain.ValidationError
			if !errors.As(err, &validationError) || validationError.Field != test.field {
				t.Fatalf("error = %v, want validation error for %q", err, test.field)
			}
		})
	}

	if _, err := service.ListProducts(context.Background(), domain.AdminListQuery{CategoryID: "curto"}); err == nil {
		t.Error("invalid categoryId error = nil, want validation error")
	}
}
