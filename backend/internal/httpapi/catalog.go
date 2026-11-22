package httpapi

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/brocados-lanches/api/internal/domain"
	"github.com/brocados-lanches/api/internal/service"
)

// CatalogService é a dependência do catálogo público usada pelos handlers.
type CatalogService interface {
	Menu(ctx context.Context) (service.Menu, error)
	DeliveryZones(ctx context.Context) ([]domain.DeliveryZone, error)
}

type menuEnvelope struct {
	Data menuData `json:"data"`
}

type menuData struct {
	Categories []menuCategoryResponse `json:"categories"`
	UpdatedAt  *string                `json:"updatedAt,omitempty"`
}

type menuCategoryResponse struct {
	ID        string                `json:"id"`
	Name      string                `json:"name"`
	Slug      string                `json:"slug"`
	SortOrder int                   `json:"sortOrder"`
	Products  []menuProductResponse `json:"products"`
}

type menuProductResponse struct {
	ID           string                    `json:"id"`
	Name         string                    `json:"name"`
	Slug         string                    `json:"slug"`
	Description  string                    `json:"description"`
	PriceCents   int64                     `json:"priceCents"`
	ImageURL     string                    `json:"imageUrl"`
	ImageAlt     string                    `json:"imageAlt"`
	Allergens    []string                  `json:"allergens"`
	Available    bool                      `json:"available"`
	Featured     bool                      `json:"featured"`
	SortOrder    int                       `json:"sortOrder"`
	OptionGroups []menuOptionGroupResponse `json:"optionGroups"`
	Addons       []menuAddonResponse       `json:"addons"`
}

type menuOptionGroupResponse struct {
	ID            string               `json:"id"`
	Name          string               `json:"name"`
	Required      bool                 `json:"required"`
	MinSelections int                  `json:"minSelections"`
	MaxSelections int                  `json:"maxSelections"`
	Options       []menuOptionResponse `json:"options"`
}

type menuOptionResponse struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	PriceCents int64  `json:"priceCents"`
	Available  bool   `json:"available"`
}

type menuAddonResponse struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	PriceCents int64  `json:"priceCents"`
	Available  bool   `json:"available"`
}

type deliveryZonesEnvelope struct {
	Data []deliveryZoneResponse `json:"data"`
}

type deliveryZoneResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	FeeCents  int64  `json:"feeCents"`
	SortOrder int    `json:"sortOrder"`
}

func handleMenu(catalog CatalogService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if catalog == nil {
			writeError(w, r, logger, errors.New("catalog service is not configured"))
			return
		}

		menu, err := catalog.Menu(r.Context())
		if err != nil {
			writeError(w, r, logger, err)
			return
		}

		writeJSON(w, http.StatusOK, menuEnvelope{Data: newMenuData(menu)})
	}
}

func handleDeliveryZones(catalog CatalogService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if catalog == nil {
			writeError(w, r, logger, errors.New("catalog service is not configured"))
			return
		}

		zones, err := catalog.DeliveryZones(r.Context())
		if err != nil {
			writeError(w, r, logger, err)
			return
		}

		writeJSON(w, http.StatusOK, deliveryZonesEnvelope{Data: newDeliveryZonesData(zones)})
	}
}

func newMenuData(menu service.Menu) menuData {
	data := menuData{
		Categories: make([]menuCategoryResponse, 0, len(menu.Categories)),
	}

	if !menu.UpdatedAt.IsZero() {
		updatedAt := menu.UpdatedAt.UTC().Format(time.RFC3339)
		data.UpdatedAt = &updatedAt
	}

	for _, category := range menu.Categories {
		products := make([]menuProductResponse, 0, len(category.Products))
		for _, product := range category.Products {
			products = append(products, newMenuProductData(product))
		}

		data.Categories = append(data.Categories, menuCategoryResponse{
			ID:        category.Category.ID,
			Name:      category.Category.Name,
			Slug:      category.Category.Slug,
			SortOrder: category.Category.SortOrder,
			Products:  products,
		})
	}

	return data
}

func newMenuProductData(item service.MenuProduct) menuProductResponse {
	product := item.Product

	allergens := make([]string, len(product.Allergens))
	copy(allergens, product.Allergens)

	optionGroups := make([]menuOptionGroupResponse, 0, len(item.OptionGroups))
	for _, group := range item.OptionGroups {
		options := make([]menuOptionResponse, 0, len(group.Options))
		for _, option := range group.Options {
			options = append(options, menuOptionResponse{
				ID:         option.ID,
				Name:       option.Name,
				PriceCents: option.PriceCents,
				Available:  option.Available,
			})
		}

		optionGroups = append(optionGroups, menuOptionGroupResponse{
			ID:            group.ID,
			Name:          group.Name,
			Required:      group.Required,
			MinSelections: group.MinSelections,
			MaxSelections: group.MaxSelections,
			Options:       options,
		})
	}

	addons := make([]menuAddonResponse, 0, len(item.Addons))
	for _, addon := range item.Addons {
		addons = append(addons, menuAddonResponse{
			ID:         addon.ID,
			Name:       addon.Name,
			PriceCents: addon.PriceCents,
			Available:  addon.Active,
		})
	}

	return menuProductResponse{
		ID:           product.ID,
		Name:         product.Name,
		Slug:         product.Slug,
		Description:  product.Description,
		PriceCents:   product.PriceCents,
		ImageURL:     product.ImageURL,
		ImageAlt:     product.ImageAlt,
		Allergens:    allergens,
		Available:    product.Available,
		Featured:     product.Featured,
		SortOrder:    product.SortOrder,
		OptionGroups: optionGroups,
		Addons:       addons,
	}
}

func newDeliveryZonesData(zones []domain.DeliveryZone) []deliveryZoneResponse {
	data := make([]deliveryZoneResponse, 0, len(zones))
	for _, zone := range zones {
		data = append(data, deliveryZoneResponse{
			ID:        zone.ID,
			Name:      zone.Name,
			FeeCents:  zone.FeeCents,
			SortOrder: zone.SortOrder,
		})
	}
	return data
}
