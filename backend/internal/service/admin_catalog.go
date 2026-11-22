package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/brocados-lanches/api/internal/domain"
)

// AdminCatalogRepository é a superfície de escrita e leitura administrativa do
// catálogo. As mesmas implementações do catálogo público são usadas para ler.
type AdminCatalogRepository interface {
	GetCategory(ctx context.Context, id string) (domain.Category, error)
	GetProduct(ctx context.Context, id string) (domain.Product, error)
	GetAddon(ctx context.Context, id string) (domain.Addon, error)
	GetDeliveryZone(ctx context.Context, id string) (domain.DeliveryZone, error)

	AdminListCategories(ctx context.Context, query domain.AdminListQuery) ([]domain.Category, int64, error)
	AdminListProducts(ctx context.Context, query domain.AdminListQuery) ([]domain.Product, int64, error)
	AdminListAddons(ctx context.Context, query domain.AdminListQuery) ([]domain.Addon, int64, error)
	AdminListDeliveryZones(ctx context.Context, query domain.AdminListQuery) ([]domain.DeliveryZone, int64, error)

	AdminCreateCategory(ctx context.Context, category domain.Category, audit domain.AuditLog) error
	AdminUpdateCategory(ctx context.Context, category domain.Category, audit domain.AuditLog) error
	AdminCreateProduct(ctx context.Context, product domain.Product, audit domain.AuditLog) error
	AdminUpdateProduct(ctx context.Context, product domain.Product, audit domain.AuditLog) error
	AdminCreateAddon(ctx context.Context, addon domain.Addon, audit domain.AuditLog) error
	AdminUpdateAddon(ctx context.Context, addon domain.Addon, audit domain.AuditLog) error
	AdminCreateDeliveryZone(ctx context.Context, zone domain.DeliveryZone, audit domain.AuditLog) error
	AdminUpdateDeliveryZone(ctx context.Context, zone domain.DeliveryZone, audit domain.AuditLog) error
}

// AdminPage é uma página de listagem com metadados para o envelope da API.
type AdminPage[T any] struct {
	Items      []T
	Page       int
	PageSize   int
	Total      int64
	TotalPages int64
}

// AdminCatalogService aplica validação e auditoria às mutações do catálogo.
type AdminCatalogService struct {
	repo  AdminCatalogRepository
	now   func() time.Time
	newID func() (string, error)
}

// NewAdminCatalogService valida as dependências do serviço administrativo.
func NewAdminCatalogService(repo AdminCatalogRepository) (*AdminCatalogService, error) {
	if repo == nil {
		return nil, errors.New("admin catalog repository is required")
	}
	return &AdminCatalogService{
		repo:  repo,
		now:   time.Now,
		newID: newObjectIDHex,
	}, nil
}

var (
	categorySortFields = []string{"name", "slug", "sortOrder", "createdAt", "updatedAt"}
	productSortFields  = []string{"name", "slug", "sortOrder", "priceCents", "createdAt", "updatedAt"}
	addonSortFields    = []string{"name", "slug", "sortOrder", "priceCents", "createdAt", "updatedAt"}
	zoneSortFields     = []string{"name", "slug", "sortOrder", "feeCents", "createdAt", "updatedAt"}
)

// ListCategories devolve uma página de categorias, inclusive inativas.
func (s *AdminCatalogService) ListCategories(ctx context.Context, query domain.AdminListQuery) (AdminPage[domain.Category], error) {
	normalized, err := normalizeAdminListQuery(query, categorySortFields)
	if err != nil {
		return AdminPage[domain.Category]{}, err
	}
	items, total, err := s.repo.AdminListCategories(ctx, normalized)
	if err != nil {
		return AdminPage[domain.Category]{}, fmt.Errorf("list categories: %w", err)
	}
	return newAdminPage(items, normalized, total), nil
}

// ListProducts devolve uma página de produtos, inclusive inativos.
func (s *AdminCatalogService) ListProducts(ctx context.Context, query domain.AdminListQuery) (AdminPage[domain.Product], error) {
	normalized, err := normalizeAdminListQuery(query, productSortFields)
	if err != nil {
		return AdminPage[domain.Product]{}, err
	}
	if query.CategoryID != "" && len(query.CategoryID) != 24 {
		return AdminPage[domain.Product]{}, invalidField("categoryId", domain.CodeInvalid, "must be a 24-character hexadecimal id")
	}
	items, total, err := s.repo.AdminListProducts(ctx, normalized)
	if err != nil {
		return AdminPage[domain.Product]{}, fmt.Errorf("list products: %w", err)
	}
	return newAdminPage(items, normalized, total), nil
}

// ListAddons devolve uma página de adicionais, inclusive inativos.
func (s *AdminCatalogService) ListAddons(ctx context.Context, query domain.AdminListQuery) (AdminPage[domain.Addon], error) {
	normalized, err := normalizeAdminListQuery(query, addonSortFields)
	if err != nil {
		return AdminPage[domain.Addon]{}, err
	}
	items, total, err := s.repo.AdminListAddons(ctx, normalized)
	if err != nil {
		return AdminPage[domain.Addon]{}, fmt.Errorf("list addons: %w", err)
	}
	return newAdminPage(items, normalized, total), nil
}

// ListDeliveryZones devolve uma página de zonas, inclusive inativas.
func (s *AdminCatalogService) ListDeliveryZones(ctx context.Context, query domain.AdminListQuery) (AdminPage[domain.DeliveryZone], error) {
	normalized, err := normalizeAdminListQuery(query, zoneSortFields)
	if err != nil {
		return AdminPage[domain.DeliveryZone]{}, err
	}
	items, total, err := s.repo.AdminListDeliveryZones(ctx, normalized)
	if err != nil {
		return AdminPage[domain.DeliveryZone]{}, fmt.Errorf("list delivery zones: %w", err)
	}
	return newAdminPage(items, normalized, total), nil
}

// CreateCategory valida, identifica e persiste uma categoria, gerando auditoria.
func (s *AdminCatalogService) CreateCategory(ctx context.Context, actorID string, category domain.Category) (domain.Category, error) {
	if err := category.Validate(); err != nil {
		return domain.Category{}, err
	}
	id, err := s.newID()
	if err != nil {
		return domain.Category{}, fmt.Errorf("generate category id: %w", err)
	}
	now := s.now().UTC()
	category.ID = id
	category.CreatedAt = now
	category.UpdatedAt = now

	audit := s.newAudit(actorID, domain.AuditActionCreate, domain.AuditEntityCategory, id, nil, categorySnapshot(category), now)
	if err := s.repo.AdminCreateCategory(ctx, category, audit); err != nil {
		return domain.Category{}, err
	}
	return category, nil
}

// UpdateCategory aplica um PATCH parcial e audita o resultado.
func (s *AdminCatalogService) UpdateCategory(ctx context.Context, actorID, id string, patch domain.CategoryPatch) (domain.Category, error) {
	if patch.Empty() {
		return domain.Category{}, emptyPatchError()
	}
	current, err := s.repo.GetCategory(ctx, id)
	if err != nil {
		return domain.Category{}, err
	}
	updated := patch.Apply(current)
	if err := updated.Validate(); err != nil {
		return domain.Category{}, err
	}
	now := s.now().UTC()
	updated.UpdatedAt = now

	audit := s.newAudit(actorID, domain.AuditActionUpdate, domain.AuditEntityCategory, id, categorySnapshot(current), categorySnapshot(updated), now)
	if err := s.repo.AdminUpdateCategory(ctx, updated, audit); err != nil {
		return domain.Category{}, err
	}
	return updated, nil
}

// CreateProduct valida, identifica e persiste um produto, gerando auditoria.
func (s *AdminCatalogService) CreateProduct(ctx context.Context, actorID string, product domain.Product) (domain.Product, error) {
	if err := product.Validate(); err != nil {
		return domain.Product{}, err
	}
	if err := validateProductReferences(product); err != nil {
		return domain.Product{}, err
	}
	id, err := s.newID()
	if err != nil {
		return domain.Product{}, fmt.Errorf("generate product id: %w", err)
	}
	now := s.now().UTC()
	product.ID = id
	product.CreatedAt = now
	product.UpdatedAt = now

	audit := s.newAudit(actorID, domain.AuditActionCreate, domain.AuditEntityProduct, id, nil, productSnapshot(product), now)
	if err := s.repo.AdminCreateProduct(ctx, product, audit); err != nil {
		return domain.Product{}, err
	}
	return product, nil
}

// UpdateProduct aplica um PATCH parcial e audita o resultado.
func (s *AdminCatalogService) UpdateProduct(ctx context.Context, actorID, id string, patch domain.ProductPatch) (domain.Product, error) {
	if patch.Empty() {
		return domain.Product{}, emptyPatchError()
	}
	current, err := s.repo.GetProduct(ctx, id)
	if err != nil {
		return domain.Product{}, err
	}
	updated := patch.Apply(current)
	if err := updated.Validate(); err != nil {
		return domain.Product{}, err
	}
	if err := validateProductReferences(updated); err != nil {
		return domain.Product{}, err
	}
	now := s.now().UTC()
	updated.UpdatedAt = now

	audit := s.newAudit(actorID, domain.AuditActionUpdate, domain.AuditEntityProduct, id, productSnapshot(current), productSnapshot(updated), now)
	if err := s.repo.AdminUpdateProduct(ctx, updated, audit); err != nil {
		return domain.Product{}, err
	}
	return updated, nil
}

// CreateAddon valida, identifica e persiste um adicional, gerando auditoria.
func (s *AdminCatalogService) CreateAddon(ctx context.Context, actorID string, addon domain.Addon) (domain.Addon, error) {
	if err := addon.Validate(); err != nil {
		return domain.Addon{}, err
	}
	id, err := s.newID()
	if err != nil {
		return domain.Addon{}, fmt.Errorf("generate addon id: %w", err)
	}
	now := s.now().UTC()
	addon.ID = id
	addon.CreatedAt = now
	addon.UpdatedAt = now

	audit := s.newAudit(actorID, domain.AuditActionCreate, domain.AuditEntityAddon, id, nil, addonSnapshot(addon), now)
	if err := s.repo.AdminCreateAddon(ctx, addon, audit); err != nil {
		return domain.Addon{}, err
	}
	return addon, nil
}

// UpdateAddon aplica um PATCH parcial e audita o resultado.
func (s *AdminCatalogService) UpdateAddon(ctx context.Context, actorID, id string, patch domain.AddonPatch) (domain.Addon, error) {
	if patch.Empty() {
		return domain.Addon{}, emptyPatchError()
	}
	current, err := s.repo.GetAddon(ctx, id)
	if err != nil {
		return domain.Addon{}, err
	}
	updated := patch.Apply(current)
	if err := updated.Validate(); err != nil {
		return domain.Addon{}, err
	}
	now := s.now().UTC()
	updated.UpdatedAt = now

	audit := s.newAudit(actorID, domain.AuditActionUpdate, domain.AuditEntityAddon, id, addonSnapshot(current), addonSnapshot(updated), now)
	if err := s.repo.AdminUpdateAddon(ctx, updated, audit); err != nil {
		return domain.Addon{}, err
	}
	return updated, nil
}

// CreateDeliveryZone valida, identifica e persiste uma zona, gerando auditoria.
func (s *AdminCatalogService) CreateDeliveryZone(ctx context.Context, actorID string, zone domain.DeliveryZone) (domain.DeliveryZone, error) {
	if err := zone.Validate(); err != nil {
		return domain.DeliveryZone{}, err
	}
	id, err := s.newID()
	if err != nil {
		return domain.DeliveryZone{}, fmt.Errorf("generate delivery zone id: %w", err)
	}
	now := s.now().UTC()
	zone.ID = id
	zone.CreatedAt = now
	zone.UpdatedAt = now

	audit := s.newAudit(actorID, domain.AuditActionCreate, domain.AuditEntityDeliveryZone, id, nil, zoneSnapshot(zone), now)
	if err := s.repo.AdminCreateDeliveryZone(ctx, zone, audit); err != nil {
		return domain.DeliveryZone{}, err
	}
	return zone, nil
}

// UpdateDeliveryZone aplica um PATCH parcial e audita o resultado.
func (s *AdminCatalogService) UpdateDeliveryZone(ctx context.Context, actorID, id string, patch domain.DeliveryZonePatch) (domain.DeliveryZone, error) {
	if patch.Empty() {
		return domain.DeliveryZone{}, emptyPatchError()
	}
	current, err := s.repo.GetDeliveryZone(ctx, id)
	if err != nil {
		return domain.DeliveryZone{}, err
	}
	updated := patch.Apply(current)
	if err := updated.Validate(); err != nil {
		return domain.DeliveryZone{}, err
	}
	now := s.now().UTC()
	updated.UpdatedAt = now

	audit := s.newAudit(actorID, domain.AuditActionUpdate, domain.AuditEntityDeliveryZone, id, zoneSnapshot(current), zoneSnapshot(updated), now)
	if err := s.repo.AdminUpdateDeliveryZone(ctx, updated, audit); err != nil {
		return domain.DeliveryZone{}, err
	}
	return updated, nil
}

func (s *AdminCatalogService) newAudit(actorID, action, entity, entityID string, before, after map[string]any, now time.Time) domain.AuditLog {
	return domain.AuditLog{
		ActorID:   actorID,
		Action:    action,
		Entity:    entity,
		EntityID:  entityID,
		Before:    before,
		After:     after,
		CreatedAt: now,
	}
}

func normalizeAdminListQuery(query domain.AdminListQuery, allowedSort []string) (domain.AdminListQuery, error) {
	if query.Page == 0 {
		query.Page = 1
	}
	if query.Page < 1 {
		return query, invalidField("page", domain.CodeOutOfRange, "must be at least 1")
	}
	if query.PageSize == 0 {
		query.PageSize = 20
	}
	if query.PageSize < 1 || query.PageSize > 100 {
		return query, invalidField("pageSize", domain.CodeOutOfRange, "must be between 1 and 100")
	}
	if query.Sort == "" {
		query.Sort = "sortOrder"
	}
	permitted := false
	for _, candidate := range allowedSort {
		if candidate == query.Sort {
			permitted = true
			break
		}
	}
	if !permitted {
		return query, invalidField("sort", domain.CodeInvalid, "unsupported sort field")
	}
	if query.Order == "" {
		query.Order = "asc"
	}
	if query.Order != "asc" && query.Order != "desc" {
		return query, invalidField("order", domain.CodeInvalid, "must be asc or desc")
	}
	if len(query.Search) > 120 {
		return query, invalidField("q", domain.CodeTooLong, "must have at most 120 characters")
	}
	return query, nil
}

func newAdminPage[T any](items []T, query domain.AdminListQuery, total int64) AdminPage[T] {
	if items == nil {
		items = []T{}
	}
	totalPages := int64(0)
	if total > 0 {
		totalPages = int64(math.Ceil(float64(total) / float64(query.PageSize)))
	}
	return AdminPage[T]{
		Items:      items,
		Page:       query.Page,
		PageSize:   query.PageSize,
		Total:      total,
		TotalPages: totalPages,
	}
}

func emptyPatchError() error {
	return invalidField("body", domain.CodeRequired, "at least one field is required")
}

// validateProductReferences complementa domain.Product.Validate: o domínio
// confere comprimento, mas as referências precisam ser ObjectIDs válidos antes
// da conversão BSON, para que um ID malformado vire 422 e não erro interno.
func validateProductReferences(product domain.Product) error {
	if err := validateObjectID("categoryId", product.CategoryID); err != nil {
		return err
	}
	for index, id := range product.AddonIDs {
		if err := validateObjectID(fmt.Sprintf("addonIds[%d]", index), id); err != nil {
			return err
		}
	}
	return nil
}

func validateObjectID(field, value string) error {
	if len(value) != 24 {
		return invalidField(field, domain.CodeInvalid, "must be a 24-character hexadecimal id")
	}
	if _, err := hex.DecodeString(value); err != nil {
		return invalidField(field, domain.CodeInvalid, "must be a 24-character hexadecimal id")
	}
	return nil
}

func invalidField(field, code, message string) error {
	return &domain.ValidationError{Field: field, Code: code, Message: message}
}

func newObjectIDHex() (string, error) {
	var bytes [12]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes[:]), nil
}

func categorySnapshot(category domain.Category) map[string]any {
	return map[string]any{
		"name": category.Name, "slug": category.Slug,
		"sortOrder": category.SortOrder, "active": category.Active,
	}
}

func productSnapshot(product domain.Product) map[string]any {
	return map[string]any{
		"name": product.Name, "slug": product.Slug,
		"categoryId": product.CategoryID, "priceCents": product.PriceCents,
		"available": product.Available, "featured": product.Featured, "active": product.Active,
		"sortOrder": product.SortOrder,
	}
}

func addonSnapshot(addon domain.Addon) map[string]any {
	return map[string]any{
		"name": addon.Name, "slug": addon.Slug,
		"priceCents": addon.PriceCents, "active": addon.Active, "sortOrder": addon.SortOrder,
	}
}

func zoneSnapshot(zone domain.DeliveryZone) map[string]any {
	return map[string]any{
		"name": zone.Name, "slug": zone.Slug,
		"feeCents": zone.FeeCents, "active": zone.Active, "sortOrder": zone.SortOrder,
	}
}
