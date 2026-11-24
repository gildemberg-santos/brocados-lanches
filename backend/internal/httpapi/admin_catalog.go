package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/brocados-lanches/api/internal/domain"
	"github.com/brocados-lanches/api/internal/service"
	"github.com/go-chi/chi/v5"
)

const maxAdminBodyBytes = 1 << 20

// AdminCatalogService é a dependência dos handlers de catálogo do painel.
type AdminCatalogService interface {
	ListCategories(ctx context.Context, query domain.AdminListQuery) (service.AdminPage[domain.Category], error)
	CreateCategory(ctx context.Context, actorID string, category domain.Category) (domain.Category, error)
	UpdateCategory(ctx context.Context, actorID, id string, patch domain.CategoryPatch) (domain.Category, error)

	ListProducts(ctx context.Context, query domain.AdminListQuery) (service.AdminPage[domain.Product], error)
	CreateProduct(ctx context.Context, actorID string, product domain.Product) (domain.Product, error)
	UpdateProduct(ctx context.Context, actorID, id string, patch domain.ProductPatch) (domain.Product, error)

	ListAddons(ctx context.Context, query domain.AdminListQuery) (service.AdminPage[domain.Addon], error)
	CreateAddon(ctx context.Context, actorID string, addon domain.Addon) (domain.Addon, error)
	UpdateAddon(ctx context.Context, actorID, id string, patch domain.AddonPatch) (domain.Addon, error)

	ListDeliveryZones(ctx context.Context, query domain.AdminListQuery) (service.AdminPage[domain.DeliveryZone], error)
	CreateDeliveryZone(ctx context.Context, actorID string, zone domain.DeliveryZone) (domain.DeliveryZone, error)
	UpdateDeliveryZone(ctx context.Context, actorID, id string, patch domain.DeliveryZonePatch) (domain.DeliveryZone, error)
}

type adminMeta struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"pageSize"`
	Total      int64 `json:"total"`
	TotalPages int64 `json:"totalPages"`
}

type adminListEnvelope[T any] struct {
	Data []T       `json:"data"`
	Meta adminMeta `json:"meta"`
}

type adminItemEnvelope[T any] struct {
	Data T `json:"data"`
}

type adminCategoryRequest struct {
	Name        *string `json:"name"`
	Slug        *string `json:"slug"`
	Description *string `json:"description"`
	SortOrder   *int    `json:"sortOrder"`
	Active      *bool   `json:"active"`
}

type adminAddonRequest struct {
	Name       *string `json:"name"`
	Slug       *string `json:"slug"`
	PriceCents *int64  `json:"priceCents"`
	SortOrder  *int    `json:"sortOrder"`
	Active     *bool   `json:"active"`
}

type adminZoneRequest struct {
	Name      *string `json:"name"`
	Slug      *string `json:"slug"`
	FeeCents  *int64  `json:"feeCents"`
	SortOrder *int    `json:"sortOrder"`
	Active    *bool   `json:"active"`
}

type adminProductRequest struct {
	CategoryID   *string                  `json:"categoryId"`
	Name         *string                  `json:"name"`
	Slug         *string                  `json:"slug"`
	Description  *string                  `json:"description"`
	PriceCents   *int64                   `json:"priceCents"`
	ImageURL     *string                  `json:"imageUrl"`
	ImageAlt     *string                  `json:"imageAlt"`
	Allergens    *[]string                `json:"allergens"`
	Available    *bool                    `json:"available"`
	Featured     *bool                    `json:"featured"`
	Active       *bool                    `json:"active"`
	SortOrder    *int                     `json:"sortOrder"`
	OptionGroups *[]adminOptionGroupInput `json:"optionGroups"`
	AddonIDs     *[]string                `json:"addonIds"`
}

type adminOptionGroupInput struct {
	ID            string             `json:"id"`
	Name          string             `json:"name"`
	Required      bool               `json:"required"`
	MinSelections int                `json:"minSelections"`
	MaxSelections int                `json:"maxSelections"`
	Options       []adminOptionInput `json:"options"`
}

type adminOptionInput struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	PriceCents int64  `json:"priceCents"`
	Available  bool   `json:"available"`
	SortOrder  int    `json:"sortOrder"`
}

type adminCategoryResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	SortOrder   int    `json:"sortOrder"`
	Active      bool   `json:"active"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

type adminAddonResponse struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Slug       string `json:"slug"`
	PriceCents int64  `json:"priceCents"`
	Active     bool   `json:"active"`
	SortOrder  int    `json:"sortOrder"`
	CreatedAt  string `json:"createdAt"`
	UpdatedAt  string `json:"updatedAt"`
}

type adminZoneResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	FeeCents  int64  `json:"feeCents"`
	Active    bool   `json:"active"`
	SortOrder int    `json:"sortOrder"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type adminProductResponse struct {
	ID           string                     `json:"id"`
	CategoryID   string                     `json:"categoryId"`
	Name         string                     `json:"name"`
	Slug         string                     `json:"slug"`
	Description  string                     `json:"description"`
	PriceCents   int64                      `json:"priceCents"`
	ImageURL     string                     `json:"imageUrl"`
	ImageAlt     string                     `json:"imageAlt"`
	Allergens    []string                   `json:"allergens"`
	Available    bool                       `json:"available"`
	Featured     bool                       `json:"featured"`
	Active       bool                       `json:"active"`
	SortOrder    int                        `json:"sortOrder"`
	OptionGroups []adminOptionGroupResponse `json:"optionGroups"`
	AddonIDs     []string                   `json:"addonIds"`
	CreatedAt    string                     `json:"createdAt"`
	UpdatedAt    string                     `json:"updatedAt"`
}

type adminOptionGroupResponse struct {
	ID            string                `json:"id"`
	Name          string                `json:"name"`
	Required      bool                  `json:"required"`
	MinSelections int                   `json:"minSelections"`
	MaxSelections int                   `json:"maxSelections"`
	Options       []adminOptionResponse `json:"options"`
}

type adminOptionResponse struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	PriceCents int64  `json:"priceCents"`
	Available  bool   `json:"available"`
	SortOrder  int    `json:"sortOrder"`
}

func handleAdminListCategories(catalog AdminCatalogService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if catalog == nil {
			writeUnconfigured(w, r, logger, "admin catalog service")
			return
		}
		query, err := parseAdminListQuery(r)
		if err != nil {
			writeError(w, r, logger, err)
			return
		}
		page, err := catalog.ListCategories(r.Context(), query)
		if err != nil {
			writeError(w, r, logger, err)
			return
		}
		items := make([]adminCategoryResponse, 0, len(page.Items))
		for _, category := range page.Items {
			items = append(items, newAdminCategoryResponse(category))
		}
		writeJSON(w, http.StatusOK, adminListEnvelope[adminCategoryResponse]{Data: items, Meta: newAdminMeta(page.Page, page.PageSize, page.Total, page.TotalPages)})
	}
}

func handleAdminCreateCategory(catalog AdminCatalogService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if catalog == nil {
			writeUnconfigured(w, r, logger, "admin catalog service")
			return
		}
		var request adminCategoryRequest
		if err := decodeAdminRequest(w, r, &request); err != nil {
			writeError(w, r, logger, err)
			return
		}
		actor, ok := adminFromContext(r.Context())
		if !ok {
			writeError(w, r, logger, domain.ErrUnauthorized)
			return
		}
		category, err := catalog.CreateCategory(r.Context(), actor.ID, request.category())
		if err != nil {
			writeError(w, r, logger, err)
			return
		}
		writeJSON(w, http.StatusCreated, adminItemEnvelope[adminCategoryResponse]{Data: newAdminCategoryResponse(category)})
	}
}

func handleAdminUpdateCategory(catalog AdminCatalogService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if catalog == nil {
			writeUnconfigured(w, r, logger, "admin catalog service")
			return
		}
		var request adminCategoryRequest
		if err := decodeAdminRequest(w, r, &request); err != nil {
			writeError(w, r, logger, err)
			return
		}
		actor, ok := adminFromContext(r.Context())
		if !ok {
			writeError(w, r, logger, domain.ErrUnauthorized)
			return
		}
		category, err := catalog.UpdateCategory(r.Context(), actor.ID, chi.URLParam(r, "id"), request.patch())
		if err != nil {
			writeError(w, r, logger, err)
			return
		}
		writeJSON(w, http.StatusOK, adminItemEnvelope[adminCategoryResponse]{Data: newAdminCategoryResponse(category)})
	}
}

func handleAdminListAddons(catalog AdminCatalogService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if catalog == nil {
			writeUnconfigured(w, r, logger, "admin catalog service")
			return
		}
		query, err := parseAdminListQuery(r)
		if err != nil {
			writeError(w, r, logger, err)
			return
		}
		page, err := catalog.ListAddons(r.Context(), query)
		if err != nil {
			writeError(w, r, logger, err)
			return
		}
		items := make([]adminAddonResponse, 0, len(page.Items))
		for _, addon := range page.Items {
			items = append(items, newAdminAddonResponse(addon))
		}
		writeJSON(w, http.StatusOK, adminListEnvelope[adminAddonResponse]{Data: items, Meta: newAdminMeta(page.Page, page.PageSize, page.Total, page.TotalPages)})
	}
}

func handleAdminCreateAddon(catalog AdminCatalogService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if catalog == nil {
			writeUnconfigured(w, r, logger, "admin catalog service")
			return
		}
		var request adminAddonRequest
		if err := decodeAdminRequest(w, r, &request); err != nil {
			writeError(w, r, logger, err)
			return
		}
		actor, ok := adminFromContext(r.Context())
		if !ok {
			writeError(w, r, logger, domain.ErrUnauthorized)
			return
		}
		addon, err := catalog.CreateAddon(r.Context(), actor.ID, request.addon())
		if err != nil {
			writeError(w, r, logger, err)
			return
		}
		writeJSON(w, http.StatusCreated, adminItemEnvelope[adminAddonResponse]{Data: newAdminAddonResponse(addon)})
	}
}

func handleAdminUpdateAddon(catalog AdminCatalogService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if catalog == nil {
			writeUnconfigured(w, r, logger, "admin catalog service")
			return
		}
		var request adminAddonRequest
		if err := decodeAdminRequest(w, r, &request); err != nil {
			writeError(w, r, logger, err)
			return
		}
		actor, ok := adminFromContext(r.Context())
		if !ok {
			writeError(w, r, logger, domain.ErrUnauthorized)
			return
		}
		addon, err := catalog.UpdateAddon(r.Context(), actor.ID, chi.URLParam(r, "id"), request.patch())
		if err != nil {
			writeError(w, r, logger, err)
			return
		}
		writeJSON(w, http.StatusOK, adminItemEnvelope[adminAddonResponse]{Data: newAdminAddonResponse(addon)})
	}
}

func handleAdminListZones(catalog AdminCatalogService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if catalog == nil {
			writeUnconfigured(w, r, logger, "admin catalog service")
			return
		}
		query, err := parseAdminListQuery(r)
		if err != nil {
			writeError(w, r, logger, err)
			return
		}
		page, err := catalog.ListDeliveryZones(r.Context(), query)
		if err != nil {
			writeError(w, r, logger, err)
			return
		}
		items := make([]adminZoneResponse, 0, len(page.Items))
		for _, zone := range page.Items {
			items = append(items, newAdminZoneResponse(zone))
		}
		writeJSON(w, http.StatusOK, adminListEnvelope[adminZoneResponse]{Data: items, Meta: newAdminMeta(page.Page, page.PageSize, page.Total, page.TotalPages)})
	}
}

func handleAdminCreateZone(catalog AdminCatalogService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if catalog == nil {
			writeUnconfigured(w, r, logger, "admin catalog service")
			return
		}
		var request adminZoneRequest
		if err := decodeAdminRequest(w, r, &request); err != nil {
			writeError(w, r, logger, err)
			return
		}
		actor, ok := adminFromContext(r.Context())
		if !ok {
			writeError(w, r, logger, domain.ErrUnauthorized)
			return
		}
		zone, err := catalog.CreateDeliveryZone(r.Context(), actor.ID, request.zone())
		if err != nil {
			writeError(w, r, logger, err)
			return
		}
		writeJSON(w, http.StatusCreated, adminItemEnvelope[adminZoneResponse]{Data: newAdminZoneResponse(zone)})
	}
}

func handleAdminUpdateZone(catalog AdminCatalogService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if catalog == nil {
			writeUnconfigured(w, r, logger, "admin catalog service")
			return
		}
		var request adminZoneRequest
		if err := decodeAdminRequest(w, r, &request); err != nil {
			writeError(w, r, logger, err)
			return
		}
		actor, ok := adminFromContext(r.Context())
		if !ok {
			writeError(w, r, logger, domain.ErrUnauthorized)
			return
		}
		zone, err := catalog.UpdateDeliveryZone(r.Context(), actor.ID, chi.URLParam(r, "id"), request.patch())
		if err != nil {
			writeError(w, r, logger, err)
			return
		}
		writeJSON(w, http.StatusOK, adminItemEnvelope[adminZoneResponse]{Data: newAdminZoneResponse(zone)})
	}
}

func handleAdminListProducts(catalog AdminCatalogService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if catalog == nil {
			writeUnconfigured(w, r, logger, "admin catalog service")
			return
		}
		query, err := parseAdminListQuery(r)
		if err != nil {
			writeError(w, r, logger, err)
			return
		}
		query.CategoryID = r.URL.Query().Get("categoryId")
		page, err := catalog.ListProducts(r.Context(), query)
		if err != nil {
			writeError(w, r, logger, err)
			return
		}
		items := make([]adminProductResponse, 0, len(page.Items))
		for _, product := range page.Items {
			items = append(items, newAdminProductResponse(product))
		}
		writeJSON(w, http.StatusOK, adminListEnvelope[adminProductResponse]{Data: items, Meta: newAdminMeta(page.Page, page.PageSize, page.Total, page.TotalPages)})
	}
}

func handleAdminCreateProduct(catalog AdminCatalogService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if catalog == nil {
			writeUnconfigured(w, r, logger, "admin catalog service")
			return
		}
		var request adminProductRequest
		if err := decodeAdminRequest(w, r, &request); err != nil {
			writeError(w, r, logger, err)
			return
		}
		actor, ok := adminFromContext(r.Context())
		if !ok {
			writeError(w, r, logger, domain.ErrUnauthorized)
			return
		}
		product, err := catalog.CreateProduct(r.Context(), actor.ID, request.product())
		if err != nil {
			writeError(w, r, logger, err)
			return
		}
		writeJSON(w, http.StatusCreated, adminItemEnvelope[adminProductResponse]{Data: newAdminProductResponse(product)})
	}
}

func handleAdminUpdateProduct(catalog AdminCatalogService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if catalog == nil {
			writeUnconfigured(w, r, logger, "admin catalog service")
			return
		}
		var request adminProductRequest
		if err := decodeAdminRequest(w, r, &request); err != nil {
			writeError(w, r, logger, err)
			return
		}
		actor, ok := adminFromContext(r.Context())
		if !ok {
			writeError(w, r, logger, domain.ErrUnauthorized)
			return
		}
		product, err := catalog.UpdateProduct(r.Context(), actor.ID, chi.URLParam(r, "id"), request.patch())
		if err != nil {
			writeError(w, r, logger, err)
			return
		}
		writeJSON(w, http.StatusOK, adminItemEnvelope[adminProductResponse]{Data: newAdminProductResponse(product)})
	}
}

func (r adminCategoryRequest) category() domain.Category {
	return domain.Category{
		Name:        stringValue(r.Name),
		Slug:        stringValue(r.Slug),
		Description: stringValue(r.Description),
		SortOrder:   intValue(r.SortOrder),
		Active:      boolValue(r.Active, true),
	}
}

func (r adminCategoryRequest) patch() domain.CategoryPatch {
	return domain.CategoryPatch{
		Name: r.Name, Slug: r.Slug, Description: r.Description,
		SortOrder: r.SortOrder, Active: r.Active,
	}
}

func (r adminAddonRequest) addon() domain.Addon {
	return domain.Addon{
		Name:       stringValue(r.Name),
		Slug:       stringValue(r.Slug),
		PriceCents: int64Value(r.PriceCents),
		SortOrder:  intValue(r.SortOrder),
		Active:     boolValue(r.Active, true),
	}
}

func (r adminAddonRequest) patch() domain.AddonPatch {
	return domain.AddonPatch{
		Name: r.Name, Slug: r.Slug, PriceCents: r.PriceCents,
		SortOrder: r.SortOrder, Active: r.Active,
	}
}

func (r adminZoneRequest) zone() domain.DeliveryZone {
	return domain.DeliveryZone{
		Name:      stringValue(r.Name),
		Slug:      stringValue(r.Slug),
		FeeCents:  int64Value(r.FeeCents),
		SortOrder: intValue(r.SortOrder),
		Active:    boolValue(r.Active, true),
	}
}

func (r adminZoneRequest) patch() domain.DeliveryZonePatch {
	return domain.DeliveryZonePatch{
		Name: r.Name, Slug: r.Slug, FeeCents: r.FeeCents,
		SortOrder: r.SortOrder, Active: r.Active,
	}
}

func (r adminProductRequest) product() domain.Product {
	return domain.Product{
		CategoryID:   stringValue(r.CategoryID),
		Name:         stringValue(r.Name),
		Slug:         stringValue(r.Slug),
		Description:  stringValue(r.Description),
		PriceCents:   int64Value(r.PriceCents),
		ImageURL:     stringValue(r.ImageURL),
		ImageAlt:     stringValue(r.ImageAlt),
		Allergens:    sliceValue(r.Allergens),
		Available:    boolValue(r.Available, true),
		Featured:     boolValue(r.Featured, false),
		Active:       boolValue(r.Active, true),
		SortOrder:    intValue(r.SortOrder),
		OptionGroups: newOptionGroups(r.OptionGroups),
		AddonIDs:     sliceValue(r.AddonIDs),
	}
}

func (r adminProductRequest) patch() domain.ProductPatch {
	var groups *[]domain.OptionGroup
	if r.OptionGroups != nil {
		converted := newOptionGroups(r.OptionGroups)
		groups = &converted
	}
	return domain.ProductPatch{
		CategoryID: r.CategoryID, Name: r.Name, Slug: r.Slug, Description: r.Description,
		PriceCents: r.PriceCents, ImageURL: r.ImageURL, ImageAlt: r.ImageAlt,
		Allergens: r.Allergens, Available: r.Available, Featured: r.Featured,
		Active: r.Active, SortOrder: r.SortOrder, OptionGroups: groups, AddonIDs: r.AddonIDs,
	}
}

func newOptionGroups(inputs *[]adminOptionGroupInput) []domain.OptionGroup {
	if inputs == nil {
		return nil
	}
	groups := make([]domain.OptionGroup, 0, len(*inputs))
	for _, input := range *inputs {
		options := make([]domain.Option, 0, len(input.Options))
		for _, option := range input.Options {
			options = append(options, domain.Option{
				ID: option.ID, Name: option.Name, PriceCents: option.PriceCents,
				Available: option.Available, SortOrder: option.SortOrder,
			})
		}
		groups = append(groups, domain.OptionGroup{
			ID: input.ID, Name: input.Name, Required: input.Required,
			MinSelections: input.MinSelections, MaxSelections: input.MaxSelections,
			Options: options,
		})
	}
	return groups
}

func decodeAdminRequest(w http.ResponseWriter, r *http.Request, target any) error {
	body := http.MaxBytesReader(w, r.Body, maxAdminBodyBytes)
	decoder := json.NewDecoder(body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return &malformedRequestError{field: "body", code: domain.CodeInvalid, message: "invalid JSON body"}
	}
	return nil
}

func parseAdminListQuery(r *http.Request) (domain.AdminListQuery, error) {
	query := r.URL.Query()
	page, err := parseOptionalInt(query.Get("page"), "page")
	if err != nil {
		return domain.AdminListQuery{}, err
	}
	pageSize, err := parseOptionalInt(query.Get("pageSize"), "pageSize")
	if err != nil {
		return domain.AdminListQuery{}, err
	}
	result := domain.AdminListQuery{
		Page:       page,
		PageSize:   pageSize,
		Sort:       query.Get("sort"),
		Order:      query.Get("order"),
		Search:     strings.TrimSpace(query.Get("q")),
		CategoryID: strings.TrimSpace(query.Get("categoryId")),
	}
	if raw := query.Get("active"); raw != "" {
		active, err := strconv.ParseBool(raw)
		if err != nil {
			return result, &domain.ValidationError{Field: "active", Code: domain.CodeInvalid, Message: "must be true or false"}
		}
		result.Active = &active
	}
	return result, nil
}

func parseOptionalInt(raw, field string) (int, error) {
	if raw == "" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, &domain.ValidationError{Field: field, Code: domain.CodeInvalid, Message: "must be an integer"}
	}
	return value, nil
}

func newAdminMeta(page, pageSize int, total, totalPages int64) adminMeta {
	return adminMeta{Page: page, PageSize: pageSize, Total: total, TotalPages: totalPages}
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func intValue(value *int) int {
	if value == nil {
		return 0
	}
	return *value
}

func int64Value(value *int64) int64 {
	if value == nil {
		return 0
	}
	return *value
}

func boolValue(value *bool, fallback bool) bool {
	if value == nil {
		return fallback
	}
	return *value
}

func sliceValue[T any](value *[]T) []T {
	if value == nil {
		return nil
	}
	return append([]T(nil), (*value)...)
}

// sliceOrEmpty preserves the response contract: nullable slices become [] instead of null.
// Mutations must keep using sliceValue so an omitted PATCH field remains distinguishable.
func sliceOrEmpty[T any](value []T) []T {
	result := make([]T, len(value))
	copy(result, value)
	return result
}

func formatAdminTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
}

func newAdminCategoryResponse(category domain.Category) adminCategoryResponse {
	return adminCategoryResponse{
		ID: category.ID, Name: category.Name, Slug: category.Slug,
		Description: category.Description, SortOrder: category.SortOrder, Active: category.Active,
		CreatedAt: formatAdminTime(category.CreatedAt), UpdatedAt: formatAdminTime(category.UpdatedAt),
	}
}

func newAdminAddonResponse(addon domain.Addon) adminAddonResponse {
	return adminAddonResponse{
		ID: addon.ID, Name: addon.Name, Slug: addon.Slug, PriceCents: addon.PriceCents,
		Active: addon.Active, SortOrder: addon.SortOrder,
		CreatedAt: formatAdminTime(addon.CreatedAt), UpdatedAt: formatAdminTime(addon.UpdatedAt),
	}
}

func newAdminZoneResponse(zone domain.DeliveryZone) adminZoneResponse {
	return adminZoneResponse{
		ID: zone.ID, Name: zone.Name, Slug: zone.Slug, FeeCents: zone.FeeCents,
		Active: zone.Active, SortOrder: zone.SortOrder,
		CreatedAt: formatAdminTime(zone.CreatedAt), UpdatedAt: formatAdminTime(zone.UpdatedAt),
	}
}

func newAdminProductResponse(product domain.Product) adminProductResponse {
	groups := make([]adminOptionGroupResponse, 0, len(product.OptionGroups))
	for _, group := range product.OptionGroups {
		options := make([]adminOptionResponse, 0, len(group.Options))
		for _, option := range group.Options {
			options = append(options, adminOptionResponse{
				ID: option.ID, Name: option.Name, PriceCents: option.PriceCents,
				Available: option.Available, SortOrder: option.SortOrder,
			})
		}
		groups = append(groups, adminOptionGroupResponse{
			ID: group.ID, Name: group.Name, Required: group.Required,
			MinSelections: group.MinSelections, MaxSelections: group.MaxSelections,
			Options: options,
		})
	}

	return adminProductResponse{
		ID: product.ID, CategoryID: product.CategoryID, Name: product.Name, Slug: product.Slug,
		Description: product.Description, PriceCents: product.PriceCents,
		ImageURL: product.ImageURL, ImageAlt: product.ImageAlt,
		Allergens: sliceOrEmpty(product.Allergens), Available: product.Available,
		Featured: product.Featured, Active: product.Active, SortOrder: product.SortOrder,
		OptionGroups: groups, AddonIDs: sliceOrEmpty(product.AddonIDs),
		CreatedAt: formatAdminTime(product.CreatedAt), UpdatedAt: formatAdminTime(product.UpdatedAt),
	}
}

func writeUnconfigured(w http.ResponseWriter, r *http.Request, logger *slog.Logger, name string) {
	writeError(w, r, logger, errors.New(name+" is not configured"))
}
