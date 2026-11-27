package domain

import (
	"errors"
	"strings"
	"testing"
)

func TestValidateSlug(t *testing.T) {
	tests := []struct {
		name     string
		slug     string
		wantCode string
	}{
		{name: "simples", slug: "hamburgueres"},
		{name: "com números", slug: "combo-2-pessoas"},
		{name: "um caractere", slug: "a"},
		{name: "vazio", slug: "", wantCode: CodeRequired},
		{name: "maiúsculas", slug: "Hamburgueres", wantCode: CodeInvalid},
		{name: "acento", slug: "hambúrgueres", wantCode: CodeInvalid},
		{name: "underscore", slug: "hamburgueres_artesanais", wantCode: CodeInvalid},
		{name: "hífen no início", slug: "-hamburgueres", wantCode: CodeInvalid},
		{name: "hífen no fim", slug: "hamburgueres-", wantCode: CodeInvalid},
		{name: "hífen duplo", slug: "hamburgueres--artesanais", wantCode: CodeInvalid},
		{name: "espaço", slug: "hamburgueres artesanais", wantCode: CodeInvalid},
		{name: "longo demais", slug: strings.Repeat("a", MaxSlugLength+1), wantCode: CodeTooLong},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateSlug("slug", test.slug)
			if test.wantCode != "" {
				assertValidationError(t, err, "slug", test.wantCode)
				return
			}
			if err != nil {
				t.Fatalf("ValidateSlug(%q) error = %v, want nil", test.slug, err)
			}
		})
	}
}

func TestCategoryValidate(t *testing.T) {
	valid := Category{
		Name:      "Hambúrgueres",
		Slug:      "hamburgueres",
		SortOrder: 1,
		Active:    true,
	}

	tests := []struct {
		name      string
		category  Category
		wantField string
		wantCode  string
	}{
		{name: "válida", category: valid},
		{
			name:      "nome vazio",
			category:  Category{Slug: "hamburgueres"},
			wantField: "name",
			wantCode:  CodeRequired,
		},
		{
			name:      "slug inválido",
			category:  Category{Name: "Hambúrgueres", Slug: "Hamburgueres"},
			wantField: "slug",
			wantCode:  CodeInvalid,
		},
		{
			name:      "ordem negativa",
			category:  Category{Name: "Hambúrgueres", Slug: "hamburgueres", SortOrder: -1},
			wantField: "sortOrder",
			wantCode:  CodeOutOfRange,
		},
		{
			name: "descrição longa",
			category: Category{
				Name:        "Hambúrgueres",
				Slug:        "hamburgueres",
				Description: strings.Repeat("a", MaxDescriptionLength+1),
			},
			wantField: "description",
			wantCode:  CodeTooLong,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.category.Validate()
			if test.wantField == "" {
				if err != nil {
					t.Fatalf("Validate() error = %v, want nil", err)
				}
				return
			}
			assertValidationError(t, err, test.wantField, test.wantCode)
		})
	}
}

func TestProductValidate(t *testing.T) {
	valid := validProduct()

	tests := []struct {
		name      string
		mutate    func(*Product)
		wantField string
		wantCode  string
	}{
		{name: "válido"},
		{
			name:      "categoria vazia",
			mutate:    func(product *Product) { product.CategoryID = "" },
			wantField: "categoryId",
			wantCode:  CodeRequired,
		},
		{
			name:      "nome vazio",
			mutate:    func(product *Product) { product.Name = "" },
			wantField: "name",
			wantCode:  CodeRequired,
		},
		{
			name:      "slug inválido",
			mutate:    func(product *Product) { product.Slug = "Brocado" },
			wantField: "slug",
			wantCode:  CodeInvalid,
		},
		{
			name:      "descrição vazia",
			mutate:    func(product *Product) { product.Description = "" },
			wantField: "description",
			wantCode:  CodeRequired,
		},
		{
			name:      "preço negativo",
			mutate:    func(product *Product) { product.PriceCents = -1 },
			wantField: "priceCents",
			wantCode:  CodeOutOfRange,
		},
		{
			name: "imagem com texto alternativo",
			mutate: func(product *Product) {
				product.ImageURL = "https://cdn.example.com/burger.webp"
				product.ImageAlt = "Hambúrguer com cheddar"
			},
		},
		{
			name: "imagem com caminho relativo do seed",
			mutate: func(product *Product) {
				product.ImageURL = "/products/burger.png"
				product.ImageAlt = "Hambúrguer com cheddar"
			},
		},
		{
			name: "imagem com caminho de upload",
			mutate: func(product *Product) {
				product.ImageURL = "/api/v1/uploads/665f1c2a9b1e4a0012a3b4f1"
				product.ImageAlt = "Hambúrguer com cheddar"
			},
		},
		{
			name: "imagem sem texto alternativo",
			mutate: func(product *Product) {
				product.ImageURL = "https://cdn.example.com/burger.webp"
			},
			wantField: "imageAlt",
			wantCode:  CodeRequired,
		},
		{
			name:      "ordem negativa",
			mutate:    func(product *Product) { product.SortOrder = -1 },
			wantField: "sortOrder",
			wantCode:  CodeOutOfRange,
		},
		{
			name: "alérgeno vazio",
			mutate: func(product *Product) {
				product.Allergens = []string{"glúten", ""}
			},
			wantField: "allergens[1]",
			wantCode:  CodeRequired,
		},
		{
			name: "alérgeno duplicado",
			mutate: func(product *Product) {
				product.Allergens = []string{"glúten", "Glúten"}
			},
			wantField: "allergens[1]",
			wantCode:  CodeDuplicate,
		},
		{
			name: "grupo de opções inválido",
			mutate: func(product *Product) {
				product.OptionGroups = []OptionGroup{{
					ID:            "doneness",
					Name:          "Ponto da carne",
					Required:      true,
					MinSelections: 0,
					MaxSelections: 1,
					Options:       []Option{{ID: "medium", Name: "Ao ponto"}},
				}}
			},
			wantField: "optionGroups[0].minSelections",
			wantCode:  CodeInvalid,
		},
		{
			name: "grupo de opções duplicado",
			mutate: func(product *Product) {
				group := validOptionGroup()
				product.OptionGroups = []OptionGroup{group, group}
			},
			wantField: "optionGroups[1].id",
			wantCode:  CodeDuplicate,
		},
		{
			name: "adicional vazio",
			mutate: func(product *Product) {
				product.AddonIDs = []string{""}
			},
			wantField: "addonIds[0]",
			wantCode:  CodeRequired,
		},
		{
			name: "adicional duplicado",
			mutate: func(product *Product) {
				product.AddonIDs = []string{"665f1c2a9b1e4a0012a3b4d1", "665f1c2a9b1e4a0012a3b4d1"}
			},
			wantField: "addonIds[1]",
			wantCode:  CodeDuplicate,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			product := valid
			if test.mutate != nil {
				test.mutate(&product)
			}

			err := product.Validate()
			if test.wantField == "" {
				if err != nil {
					t.Fatalf("Validate() error = %v, want nil", err)
				}
				return
			}
			assertValidationError(t, err, test.wantField, test.wantCode)
		})
	}
}

func TestOptionGroupValidate(t *testing.T) {
	valid := validOptionGroup()

	tests := []struct {
		name      string
		mutate    func(*OptionGroup)
		wantField string
		wantCode  string
	}{
		{name: "válido"},
		{
			name:      "id vazio",
			mutate:    func(group *OptionGroup) { group.ID = "" },
			wantField: "id",
			wantCode:  CodeRequired,
		},
		{
			name:      "nome vazio",
			mutate:    func(group *OptionGroup) { group.Name = "" },
			wantField: "name",
			wantCode:  CodeRequired,
		},
		{
			name:      "mínimo negativo",
			mutate:    func(group *OptionGroup) { group.MinSelections = -1 },
			wantField: "minSelections",
			wantCode:  CodeOutOfRange,
		},
		{
			name:      "máximo zero",
			mutate:    func(group *OptionGroup) { group.MaxSelections = 0 },
			wantField: "maxSelections",
			wantCode:  CodeOutOfRange,
		},
		{
			name:      "mínimo maior que máximo",
			mutate:    func(group *OptionGroup) { group.MinSelections = 2; group.MaxSelections = 1 },
			wantField: "minSelections",
			wantCode:  CodeInvalid,
		},
		{
			name:      "obrigatório sem mínimo",
			mutate:    func(group *OptionGroup) { group.Required = true; group.MinSelections = 0 },
			wantField: "minSelections",
			wantCode:  CodeInvalid,
		},
		{
			name:      "sem opções",
			mutate:    func(group *OptionGroup) { group.Options = nil },
			wantField: "options",
			wantCode:  CodeRequired,
		},
		{
			name:      "máximo maior que opções",
			mutate:    func(group *OptionGroup) { group.MaxSelections = 3 },
			wantField: "maxSelections",
			wantCode:  CodeOutOfRange,
		},
		{
			name: "opção duplicada",
			mutate: func(group *OptionGroup) {
				group.Options = append(group.Options, group.Options[0])
			},
			wantField: "options[2].id",
			wantCode:  CodeDuplicate,
		},
		{
			name: "opção com preço negativo",
			mutate: func(group *OptionGroup) {
				group.Options[0].PriceCents = -1
			},
			wantField: "options[0].priceCents",
			wantCode:  CodeOutOfRange,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			group := valid
			if test.mutate != nil {
				test.mutate(&group)
			}

			err := group.Validate()
			if test.wantField == "" {
				if err != nil {
					t.Fatalf("Validate() error = %v, want nil", err)
				}
				return
			}
			assertValidationError(t, err, test.wantField, test.wantCode)
		})
	}
}

func TestAddonValidate(t *testing.T) {
	valid := Addon{
		Name:       "Bacon Crocante",
		Slug:       "bacon-crocante",
		PriceCents: 500,
		Active:     true,
		SortOrder:  1,
	}

	tests := []struct {
		name      string
		addon     Addon
		wantField string
		wantCode  string
	}{
		{name: "válido", addon: valid},
		{
			name:      "nome vazio",
			addon:     Addon{Slug: "bacon-crocante"},
			wantField: "name",
			wantCode:  CodeRequired,
		},
		{
			name:      "slug inválido",
			addon:     Addon{Name: "Bacon", Slug: "Bacon Crocante"},
			wantField: "slug",
			wantCode:  CodeInvalid,
		},
		{
			name:      "preço negativo",
			addon:     Addon{Name: "Bacon", Slug: "bacon", PriceCents: -1},
			wantField: "priceCents",
			wantCode:  CodeOutOfRange,
		},
		{
			name:      "ordem negativa",
			addon:     Addon{Name: "Bacon", Slug: "bacon", SortOrder: -1},
			wantField: "sortOrder",
			wantCode:  CodeOutOfRange,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.addon.Validate()
			if test.wantField == "" {
				if err != nil {
					t.Fatalf("Validate() error = %v, want nil", err)
				}
				return
			}
			assertValidationError(t, err, test.wantField, test.wantCode)
		})
	}
}

func TestDeliveryZoneValidate(t *testing.T) {
	valid := DeliveryZone{
		Name:      "Centro",
		Slug:      "centro",
		FeeCents:  500,
		Active:    true,
		SortOrder: 1,
	}

	tests := []struct {
		name      string
		zone      DeliveryZone
		wantField string
		wantCode  string
	}{
		{name: "válida", zone: valid},
		{
			name:      "nome vazio",
			zone:      DeliveryZone{Slug: "centro"},
			wantField: "name",
			wantCode:  CodeRequired,
		},
		{
			name:      "slug inválido",
			zone:      DeliveryZone{Name: "Centro", Slug: "Centro"},
			wantField: "slug",
			wantCode:  CodeInvalid,
		},
		{
			name:      "taxa negativa",
			zone:      DeliveryZone{Name: "Centro", Slug: "centro", FeeCents: -1},
			wantField: "feeCents",
			wantCode:  CodeOutOfRange,
		},
		{
			name:      "ordem negativa",
			zone:      DeliveryZone{Name: "Centro", Slug: "centro", SortOrder: -1},
			wantField: "sortOrder",
			wantCode:  CodeOutOfRange,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.zone.Validate()
			if test.wantField == "" {
				if err != nil {
					t.Fatalf("Validate() error = %v, want nil", err)
				}
				return
			}
			assertValidationError(t, err, test.wantField, test.wantCode)
		})
	}
}

func validProduct() Product {
	return Product{
		CategoryID:  "665f1c2a9b1e4a0012a3b4c1",
		Name:        "Brocado's Burguer Cheddar",
		Slug:        "brocados-burguer-cheddar",
		Description: "Blend artesanal 160g, cheddar e pão brioche.",
		PriceCents:  2300,
		Available:   true,
		Featured:    true,
		Active:      true,
		SortOrder:   1,
		OptionGroups: []OptionGroup{
			validOptionGroup(),
		},
		AddonIDs: []string{"665f1c2a9b1e4a0012a3b4d1"},
	}
}

func validOptionGroup() OptionGroup {
	return OptionGroup{
		ID:            "doneness",
		Name:          "Ponto da carne",
		Required:      true,
		MinSelections: 1,
		MaxSelections: 1,
		Options: []Option{
			{ID: "medium", Name: "Ao ponto", SortOrder: 1},
			{ID: "well_done", Name: "Bem passado", SortOrder: 2},
		},
	}
}

func assertValidationError(t *testing.T, err error, wantField, wantCode string) {
	t.Helper()

	if err == nil {
		t.Fatal("error = nil, want validation error")
	}
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("errors.Is(err, ErrValidation) = false; err = %v", err)
	}

	var validationError *ValidationError
	if !errors.As(err, &validationError) {
		t.Fatalf("errors.As(err, *ValidationError) = false; err = %v", err)
	}
	if validationError.Field != wantField {
		t.Errorf("field = %q, want %q", validationError.Field, wantField)
	}
	if validationError.Code != wantCode {
		t.Errorf("code = %q, want %q", validationError.Code, wantCode)
	}
}
