package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/brocados-lanches/api/internal/domain"
)

type fakeCatalogReader struct {
	categories []domain.Category
	products   []domain.Product
	addons     []domain.Addon
	zones      []domain.DeliveryZone

	categoriesErr error
	productsErr   error
	addonsErr     error
	zonesErr      error
}

func (f fakeCatalogReader) ListActiveCategories(context.Context) ([]domain.Category, error) {
	return f.categories, f.categoriesErr
}

func (f fakeCatalogReader) ListActiveProducts(context.Context) ([]domain.Product, error) {
	return f.products, f.productsErr
}

func (f fakeCatalogReader) ListActiveAddons(context.Context) ([]domain.Addon, error) {
	return f.addons, f.addonsErr
}

func (f fakeCatalogReader) ListActiveDeliveryZones(context.Context) ([]domain.DeliveryZone, error) {
	return f.zones, f.zonesErr
}

func TestNewCatalogServiceRequiresReader(t *testing.T) {
	if _, err := NewCatalogService(nil); err == nil {
		t.Fatal("NewCatalogService(nil) error = nil, want error")
	}
}

func TestCatalogServiceMenuFiltersAndGroups(t *testing.T) {
	base := time.Date(2026, 11, 21, 12, 0, 0, 0, time.UTC)

	reader := fakeCatalogReader{
		categories: []domain.Category{
			{
				ID:        "cat-drinks",
				Name:      "Bebidas",
				Slug:      "bebidas",
				SortOrder: 2,
				Active:    true,
				UpdatedAt: base.Add(2 * time.Hour),
			},
			{
				ID:        "cat-burgers",
				Name:      "Hambúrgueres",
				Slug:      "hamburgueres",
				SortOrder: 1,
				Active:    true,
				UpdatedAt: base.Add(time.Hour),
			},
			{
				ID:        "cat-empty",
				Name:      "Sem produtos",
				Slug:      "sem-produtos",
				SortOrder: 3,
				Active:    true,
				UpdatedAt: base.Add(10 * time.Hour),
			},
			{
				ID:        "cat-inactive",
				Name:      "Inativa",
				Slug:      "inativa",
				SortOrder: 0,
				Active:    false,
				UpdatedAt: base.Add(20 * time.Hour),
			},
		},
		products: []domain.Product{
			{
				ID:          "prod-z",
				CategoryID:  "cat-burgers",
				Name:        "Z Burguer",
				Slug:        "z-burguer",
				Description: "Produto esgotado.",
				PriceCents:  2500,
				Available:   false,
				Active:      true,
				SortOrder:   2,
				Allergens:   []string{"leite"},
				OptionGroups: []domain.OptionGroup{
					{
						ID:            "doneness",
						Name:          "Ponto da carne",
						Required:      true,
						MinSelections: 1,
						MaxSelections: 1,
						Options: []domain.Option{
							{ID: "well-done", Name: "Bem passado", PriceCents: 0, Available: true, SortOrder: 2},
							{ID: "medium", Name: "Ao ponto", PriceCents: 0, Available: true, SortOrder: 1},
							{ID: "rare", Name: "Mal passado", PriceCents: 0, Available: false, SortOrder: 0},
						},
					},
					{
						ID:            "extras",
						Name:          "Extras",
						MinSelections: 0,
						MaxSelections: 2,
						Options: []domain.Option{
							{ID: "egg", Name: "Ovo", PriceCents: 250, Available: false, SortOrder: 1},
						},
					},
				},
				AddonIDs:  []string{"addon-bacon", "addon-inactive", "missing", "addon-cheddar"},
				UpdatedAt: base.Add(4 * time.Hour),
			},
			{
				ID:          "prod-a",
				CategoryID:  "cat-burgers",
				Name:        "A Burguer",
				Slug:        "a-burguer",
				Description: "Produto disponível.",
				PriceCents:  2200,
				Available:   true,
				Active:      true,
				SortOrder:   1,
				UpdatedAt:   base.Add(3 * time.Hour),
			},
			{
				ID:         "prod-inactive",
				CategoryID: "cat-burgers",
				Name:       "Inativo",
				Slug:       "inativo",
				Active:     false,
			},
			{
				ID:         "prod-orphan",
				CategoryID: "cat-inactive",
				Name:       "Órfão",
				Slug:       "orfao",
				Active:     true,
			},
			{
				ID:          "prod-drink",
				CategoryID:  "cat-drinks",
				Name:        "Refrigerante",
				Slug:        "refrigerante",
				Description: "Lata 350 ml.",
				PriceCents:  600,
				Available:   true,
				Active:      true,
				SortOrder:   1,
				UpdatedAt:   base.Add(2 * time.Hour),
			},
		},
		addons: []domain.Addon{
			{
				ID:         "addon-bacon",
				Name:       "Bacon",
				Slug:       "bacon",
				PriceCents: 500,
				Active:     true,
				SortOrder:  2,
				UpdatedAt:  base.Add(5 * time.Hour),
			},
			{
				ID:         "addon-cheddar",
				Name:       "Cheddar",
				Slug:       "cheddar",
				PriceCents: 300,
				Active:     true,
				SortOrder:  1,
				UpdatedAt:  base.Add(6 * time.Hour),
			},
			{
				ID:         "addon-inactive",
				Name:       "Cebola",
				Slug:       "cebola",
				PriceCents: 200,
				Active:     false,
				SortOrder:  3,
			},
			{
				ID:         "addon-unlinked",
				Name:       "Ovo",
				Slug:       "ovo",
				PriceCents: 250,
				Active:     true,
				SortOrder:  4,
			},
		},
	}

	catalog := mustCatalogService(t, reader)
	menu, err := catalog.Menu(context.Background())
	if err != nil {
		t.Fatalf("Menu() error = %v", err)
	}

	if len(menu.Categories) != 2 {
		t.Fatalf("len(Categories) = %d, want 2", len(menu.Categories))
	}
	if menu.Categories[0].Category.ID != "cat-burgers" || menu.Categories[1].Category.ID != "cat-drinks" {
		t.Fatalf("category order = %q, %q", menu.Categories[0].Category.ID, menu.Categories[1].Category.ID)
	}
	if menu.Categories[0].Category.Name != "Hambúrgueres" {
		t.Errorf("first category name = %q", menu.Categories[0].Category.Name)
	}

	burgers := menu.Categories[0].Products
	if len(burgers) != 2 {
		t.Fatalf("len(burgers) = %d, want 2", len(burgers))
	}
	if burgers[0].Product.ID != "prod-a" || burgers[1].Product.ID != "prod-z" {
		t.Fatalf("product order = %q, %q", burgers[0].Product.ID, burgers[1].Product.ID)
	}
	if burgers[1].Product.Available {
		t.Error("produto indisponível foi marcado como disponível")
	}
	if burgers[0].Product.Allergens == nil {
		t.Error("Allergens = nil, want empty slice")
	}
	if burgers[0].Addons == nil {
		t.Error("Addons = nil, want empty slice")
	}
	if burgers[0].OptionGroups == nil {
		t.Error("OptionGroups = nil, want empty slice")
	}

	optionGroups := burgers[1].OptionGroups
	if len(optionGroups) != 2 {
		t.Fatalf("len(OptionGroups) = %d, want 2", len(optionGroups))
	}
	if optionGroups[0].ID != "doneness" || optionGroups[1].ID != "extras" {
		t.Fatalf("option group order = %q, %q", optionGroups[0].ID, optionGroups[1].ID)
	}
	if len(optionGroups[0].Options) != 2 {
		t.Fatalf("len(doneness options) = %d, want 2", len(optionGroups[0].Options))
	}
	if optionGroups[0].Options[0].ID != "medium" || optionGroups[0].Options[1].ID != "well-done" {
		t.Errorf("option order = %q, %q", optionGroups[0].Options[0].ID, optionGroups[0].Options[1].ID)
	}
	if len(optionGroups[1].Options) != 0 {
		t.Errorf("len(extras options) = %d, want 0", len(optionGroups[1].Options))
	}

	addons := burgers[1].Addons
	if len(addons) != 2 {
		t.Fatalf("len(Addons) = %d, want 2", len(addons))
	}
	if addons[0].ID != "addon-cheddar" || addons[1].ID != "addon-bacon" {
		t.Errorf("addon order = %q, %q", addons[0].ID, addons[1].ID)
	}

	wantUpdatedAt := base.Add(6 * time.Hour)
	if !menu.UpdatedAt.Equal(wantUpdatedAt) {
		t.Errorf("UpdatedAt = %v, want %v", menu.UpdatedAt, wantUpdatedAt)
	}
}

func TestCatalogServiceMenuEmpty(t *testing.T) {
	catalog := mustCatalogService(t, fakeCatalogReader{})

	menu, err := catalog.Menu(context.Background())
	if err != nil {
		t.Fatalf("Menu() error = %v", err)
	}
	if menu.Categories == nil {
		t.Error("Categories = nil, want empty slice")
	}
	if len(menu.Categories) != 0 {
		t.Errorf("len(Categories) = %d, want 0", len(menu.Categories))
	}
	if !menu.UpdatedAt.IsZero() {
		t.Errorf("UpdatedAt = %v, want zero", menu.UpdatedAt)
	}
}

func TestCatalogServiceMenuErrors(t *testing.T) {
	tests := []struct {
		name   string
		reader fakeCatalogReader
		want   string
	}{
		{
			name:   "categorias",
			reader: fakeCatalogReader{categoriesErr: errors.New("mongo down")},
			want:   "list active categories",
		},
		{
			name:   "produtos",
			reader: fakeCatalogReader{productsErr: errors.New("mongo down")},
			want:   "list active products",
		},
		{
			name:   "adicionais",
			reader: fakeCatalogReader{addonsErr: errors.New("mongo down")},
			want:   "list active addons",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			catalog := mustCatalogService(t, test.reader)

			_, err := catalog.Menu(context.Background())
			if err == nil {
				t.Fatal("Menu() error = nil, want error")
			}
			if !strings.Contains(err.Error(), test.want) {
				t.Errorf("Menu() error = %q, want substring %q", err, test.want)
			}
		})
	}
}

func TestCatalogServiceDeliveryZones(t *testing.T) {
	reader := fakeCatalogReader{
		zones: []domain.DeliveryZone{
			{ID: "zone-centro", Name: "Centro", FeeCents: 1000, Active: true, SortOrder: 2},
			{ID: "zone-inactive", Name: "Inativa", FeeCents: 1500, Active: false},
			{ID: "zone-tapera", Name: "Tapera", FeeCents: 800, Active: true, SortOrder: 1},
		},
	}

	catalog := mustCatalogService(t, reader)
	zones, err := catalog.DeliveryZones(context.Background())
	if err != nil {
		t.Fatalf("DeliveryZones() error = %v", err)
	}
	if len(zones) != 2 {
		t.Fatalf("len(zones) = %d, want 2", len(zones))
	}
	if zones[0].ID != "zone-tapera" || zones[1].ID != "zone-centro" {
		t.Errorf("zone order = %q, %q", zones[0].ID, zones[1].ID)
	}
	if zones[0].FeeCents != 800 || zones[1].FeeCents != 1000 {
		t.Errorf("zone fees = %d, %d", zones[0].FeeCents, zones[1].FeeCents)
	}

	emptyCatalog := mustCatalogService(t, fakeCatalogReader{})
	emptyZones, err := emptyCatalog.DeliveryZones(context.Background())
	if err != nil {
		t.Fatalf("DeliveryZones() empty error = %v", err)
	}
	if emptyZones == nil {
		t.Error("empty zones = nil, want empty slice")
	}
}

func TestCatalogServiceDeliveryZonesError(t *testing.T) {
	catalog := mustCatalogService(t, fakeCatalogReader{zonesErr: errors.New("mongo down")})

	_, err := catalog.DeliveryZones(context.Background())
	if err == nil {
		t.Fatal("DeliveryZones() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "list active delivery zones") {
		t.Errorf("DeliveryZones() error = %q", err)
	}
}

func mustCatalogService(t *testing.T, reader CatalogReader) *CatalogService {
	t.Helper()

	catalog, err := NewCatalogService(reader)
	if err != nil {
		t.Fatalf("NewCatalogService() error = %v", err)
	}
	return catalog
}
