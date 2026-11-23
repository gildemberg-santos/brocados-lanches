package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/brocados-lanches/api/internal/domain"
)

// CatalogReader é a superfície do repositório necessária ao catálogo público.
type CatalogReader interface {
	ListActiveCategories(ctx context.Context) ([]domain.Category, error)
	ListActiveProducts(ctx context.Context) ([]domain.Product, error)
	ListActiveAddons(ctx context.Context) ([]domain.Addon, error)
	ListActiveDeliveryZones(ctx context.Context) ([]domain.DeliveryZone, error)
}

// CatalogService aplica as regras de exibição do catálogo público.
type CatalogService struct {
	reader CatalogReader
}

// NewCatalogService vincula o service ao repositório de catálogo.
func NewCatalogService(reader CatalogReader) (*CatalogService, error) {
	if reader == nil {
		return nil, errors.New("catalog reader is required")
	}
	return &CatalogService{reader: reader}, nil
}

// Menu é o cardápio público já filtrado e agrupado por categoria.
type Menu struct {
	Categories []MenuCategory
	UpdatedAt  time.Time
}

// MenuCategory reúne uma categoria ativa e seus produtos ativos.
type MenuCategory struct {
	Category domain.Category
	Products []MenuProduct
}

// MenuProduct reúne o produto, os grupos de opções disponíveis e os adicionais
// ativos vinculados ao produto.
type MenuProduct struct {
	Product      domain.Product
	OptionGroups []domain.OptionGroup
	Addons       []domain.Addon
}

// Menu monta o cardápio público.
//
// Regras:
//   - categorias e produtos inativos não aparecem;
//   - produto ativo e indisponível aparece para permitir o estado "esgotado";
//   - categorias sem produtos ativos são omitidas;
//   - opções indisponíveis são removidas, mas o grupo permanece com
//     Options vazio para que o cliente bloqueie a compra;
//   - adicionais só aparecem se estiverem ativos e vinculados ao produto.
func (s *CatalogService) Menu(ctx context.Context) (Menu, error) {
	categories, err := s.reader.ListActiveCategories(ctx)
	if err != nil {
		return Menu{}, fmt.Errorf("list active categories: %w", err)
	}

	products, err := s.reader.ListActiveProducts(ctx)
	if err != nil {
		return Menu{}, fmt.Errorf("list active products: %w", err)
	}

	addons, err := s.reader.ListActiveAddons(ctx)
	if err != nil {
		return Menu{}, fmt.Errorf("list active addons: %w", err)
	}

	activeAddons := make(map[string]domain.Addon, len(addons))
	for _, addon := range addons {
		if addon.Active {
			activeAddons[addon.ID] = addon
		}
	}

	productsByCategory := make(map[string][]domain.Product)
	for _, product := range products {
		if product.Active {
			productsByCategory[product.CategoryID] = append(productsByCategory[product.CategoryID], product)
		}
	}

	activeCategories := make([]domain.Category, 0, len(categories))
	for _, category := range categories {
		if category.Active {
			activeCategories = append(activeCategories, category)
		}
	}
	sort.SliceStable(activeCategories, func(left, right int) bool {
		return categoryLess(activeCategories[left], activeCategories[right])
	})

	menu := Menu{Categories: make([]MenuCategory, 0, len(activeCategories))}
	for _, category := range activeCategories {
		categoryProducts := productsByCategory[category.ID]
		if len(categoryProducts) == 0 {
			continue
		}

		sort.SliceStable(categoryProducts, func(left, right int) bool {
			return productLess(categoryProducts[left], categoryProducts[right])
		})

		menuCategory := MenuCategory{
			Category: category,
			Products: make([]MenuProduct, 0, len(categoryProducts)),
		}
		menu.UpdatedAt = latestTime(menu.UpdatedAt, category.UpdatedAt)

		for _, product := range categoryProducts {
			selectedAddons := selectAddons(product.AddonIDs, activeAddons)

			normalized := product
			normalized.Allergens = append([]string{}, product.Allergens...)
			normalized.OptionGroups = nil
			normalized.AddonIDs = nil

			menuCategory.Products = append(menuCategory.Products, MenuProduct{
				Product:      normalized,
				OptionGroups: filterOptionGroups(product.OptionGroups),
				Addons:       selectedAddons,
			})

			menu.UpdatedAt = latestTime(menu.UpdatedAt, product.UpdatedAt)
			for _, addon := range selectedAddons {
				menu.UpdatedAt = latestTime(menu.UpdatedAt, addon.UpdatedAt)
			}
		}

		menu.Categories = append(menu.Categories, menuCategory)
	}

	return menu, nil
}

// DeliveryZones devolve as zonas ativas ordenadas para exibição.
func (s *CatalogService) DeliveryZones(ctx context.Context) ([]domain.DeliveryZone, error) {
	zones, err := s.reader.ListActiveDeliveryZones(ctx)
	if err != nil {
		return nil, fmt.Errorf("list active delivery zones: %w", err)
	}

	active := make([]domain.DeliveryZone, 0, len(zones))
	for _, zone := range zones {
		if zone.Active {
			active = append(active, zone)
		}
	}
	sort.SliceStable(active, func(left, right int) bool {
		return zoneLess(active[left], active[right])
	})

	return active, nil
}

func filterOptionGroups(groups []domain.OptionGroup) []domain.OptionGroup {
	filtered := make([]domain.OptionGroup, 0, len(groups))
	for _, group := range groups {
		options := make([]domain.Option, 0, len(group.Options))
		for _, option := range group.Options {
			if option.Available {
				options = append(options, option)
			}
		}
		sort.SliceStable(options, func(left, right int) bool {
			return optionLess(options[left], options[right])
		})
		group.Options = options
		filtered = append(filtered, group)
	}
	return filtered
}

func selectAddons(ids []string, active map[string]domain.Addon) []domain.Addon {
	selected := make([]domain.Addon, 0, len(ids))
	seen := make(map[string]struct{}, len(ids))

	for _, id := range ids {
		if _, duplicate := seen[id]; duplicate {
			continue
		}
		seen[id] = struct{}{}

		addon, ok := active[id]
		if !ok {
			continue
		}
		selected = append(selected, addon)
	}

	sort.SliceStable(selected, func(left, right int) bool {
		return addonLess(selected[left], selected[right])
	})
	return selected
}

func categoryLess(left, right domain.Category) bool {
	if left.SortOrder != right.SortOrder {
		return left.SortOrder < right.SortOrder
	}
	if left.Name != right.Name {
		return left.Name < right.Name
	}
	return left.ID < right.ID
}

func productLess(left, right domain.Product) bool {
	if left.SortOrder != right.SortOrder {
		return left.SortOrder < right.SortOrder
	}
	if left.Name != right.Name {
		return left.Name < right.Name
	}
	return left.ID < right.ID
}

func optionLess(left, right domain.Option) bool {
	if left.SortOrder != right.SortOrder {
		return left.SortOrder < right.SortOrder
	}
	if left.Name != right.Name {
		return left.Name < right.Name
	}
	return left.ID < right.ID
}

func addonLess(left, right domain.Addon) bool {
	if left.SortOrder != right.SortOrder {
		return left.SortOrder < right.SortOrder
	}
	if left.Name != right.Name {
		return left.Name < right.Name
	}
	return left.ID < right.ID
}

func zoneLess(left, right domain.DeliveryZone) bool {
	if left.SortOrder != right.SortOrder {
		return left.SortOrder < right.SortOrder
	}
	if left.Name != right.Name {
		return left.Name < right.Name
	}
	return left.ID < right.ID
}

func latestTime(current, candidate time.Time) time.Time {
	if candidate.After(current) {
		return candidate
	}
	return current
}
