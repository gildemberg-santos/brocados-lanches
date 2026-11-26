// Package seed carrega os dados fictícios de design-system/seed.json e os
// converte em documentos prontos para o MongoDB de desenvolvimento ou teste.
package seed

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/brocados-lanches/api/internal/domain"
)

// Data representa o conteúdo lido de design-system/seed.json.
type Data struct {
	Store         Store          `json:"store"`
	Categories    []Category     `json:"categories"`
	Products      []Product      `json:"products"`
	DeliveryZones []DeliveryZone `json:"delivery_zones"`
}

// Store reúne os dados de desenvolvimento da loja.
type Store struct {
	Name         string                `json:"name"`
	IsOpen       bool                  `json:"isOpen"`
	DeliveryTime string                `json:"deliveryTime"`
	WhatsApp     string                `json:"whatsapp"`
	Instagram    string                `json:"instagram"`
	Address      string                `json:"address"`
	OpeningHours []domain.OpeningHours `json:"openingHours"`
}

// Category é uma categoria do protótipo.
type Category struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Active bool   `json:"active"`
}

// Product é um produto do protótipo. Price é lido como texto para nunca passar
// por ponto flutuante.
type Product struct {
	ID         string      `json:"id"`
	CategoryID string      `json:"catId"`
	Name       string      `json:"name"`
	Price      json.Number `json:"price"`
	Desc       string      `json:"desc"`
	Image      string      `json:"image"`
	ImageAlt   string      `json:"imageAlt"`
	Available  bool        `json:"available"`
	Featured   bool        `json:"featured"`
}

// DeliveryZone é uma zona de entrega do protótipo.
type DeliveryZone struct {
	ID     string      `json:"id"`
	Name   string      `json:"name"`
	Fee    json.Number `json:"fee"`
	Active bool        `json:"active"`
}

// Plan é o conjunto determinístico de documentos que o seed aplica.
type Plan struct {
	Store      StoreSettings
	Categories []CategoryDocument
	Products   []ProductDocument
	Zones      []ZoneDocument
}

// StoreSettings são os dados de loja que vão para o documento único.
type StoreSettings struct {
	Name         string
	IsOpen       bool
	DeliveryTime string
	WhatsApp     string
	Instagram    string
	Address      string
	OpeningHours []domain.OpeningHours
}

// CategoryDocument é uma categoria pronta para persistência.
type CategoryDocument struct {
	Slug      string
	Name      string
	Active    bool
	SortOrder int
}

// ProductDocument é um produto pronto para persistência, já em centavos.
type ProductDocument struct {
	Slug         string
	Name         string
	Description  string
	PriceCents   int64
	CategorySlug string
	ImageURL     string
	ImageAlt     string
	Available    bool
	Featured     bool
	Active       bool
	SortOrder    int
}

// ZoneDocument é uma zona pronta para persistência, já em centavos.
type ZoneDocument struct {
	Slug      string
	Name      string
	FeeCents  int64
	Active    bool
	SortOrder int
}

// Load lê e valida o arquivo de seed.
func Load(path string) (data Data, err error) {
	file, err := os.Open(path)
	if err != nil {
		return Data{}, fmt.Errorf("open seed file: %w", err)
	}
	defer func() {
		err = errors.Join(err, file.Close())
	}()

	decoder := json.NewDecoder(file)
	decoder.UseNumber()
	if err := decoder.Decode(&data); err != nil {
		return Data{}, fmt.Errorf("decode seed file: %w", err)
	}

	if err := data.Validate(); err != nil {
		return Data{}, err
	}

	return data, nil
}

// Validate confere os vínculos, os nomes e os valores monetários do seed.
func (d Data) Validate() error {
	var problems []error

	// O seed precisa ser aceito pela mesma validação da API: um whatsapp com
	// máscara, por exemplo, impede qualquer PATCH parcial em /admin/store.
	store := domain.StoreSettings{
		Name:         d.Store.Name,
		DeliveryTime: d.Store.DeliveryTime,
		WhatsApp:     d.Store.WhatsApp,
		Instagram:    d.Store.Instagram,
		Address:      d.Store.Address,
		OpeningHours: d.Store.OpeningHours,
	}
	if err := store.Validate(); err != nil {
		problems = append(problems, fmt.Errorf("store: %w", err))
	}
	if len(d.Categories) == 0 {
		problems = append(problems, errors.New("at least one category is required"))
	}

	categorySlugs := make(map[string]string, len(d.Categories))
	categoryIDs := make(map[string]string, len(d.Categories))
	for index, category := range d.Categories {
		position := fmt.Sprintf("categories[%d]", index)
		if strings.TrimSpace(category.ID) == "" {
			problems = append(problems, fmt.Errorf("%s: id is required", position))
		} else if previous, ok := categoryIDs[category.ID]; ok {
			problems = append(problems, fmt.Errorf("%s: duplicate id %q used by %q", position, category.ID, previous))
		} else {
			categoryIDs[category.ID] = category.Name
		}

		slug := Slug(category.Name)
		if slug == "" {
			problems = append(problems, fmt.Errorf("%s: name must produce a slug", position))
			continue
		}
		if previous, ok := categorySlugs[slug]; ok {
			problems = append(problems, fmt.Errorf("%s: slug %q duplicates %q", position, slug, previous))
			continue
		}
		categorySlugs[slug] = category.Name
	}

	productSlugs := make(map[string]string, len(d.Products))
	productImages := make(map[string]string, len(d.Products))
	for index, product := range d.Products {
		position := fmt.Sprintf("products[%d]", index)
		if strings.TrimSpace(product.Name) == "" {
			problems = append(problems, fmt.Errorf("%s: name is required", position))
		}
		slug := Slug(product.Name)
		if slug == "" {
			problems = append(problems, fmt.Errorf("%s: name must produce a slug", position))
		} else if previous, ok := productSlugs[slug]; ok {
			problems = append(problems, fmt.Errorf("%s: slug %q duplicates %q", position, slug, previous))
		} else {
			productSlugs[slug] = product.Name
		}

		if _, ok := categoryIDs[product.CategoryID]; !ok {
			problems = append(problems, fmt.Errorf("%s: category %q does not exist", position, product.CategoryID))
		}
		if _, err := CentsFromReais(product.Price.String()); err != nil {
			problems = append(problems, fmt.Errorf("%s: invalid price: %w", position, err))
		}
		if product.Image != "" {
			if !strings.HasPrefix(product.Image, "/products/") || strings.Contains(product.Image, "..") {
				problems = append(problems, fmt.Errorf("%s: image must be a local path under /products/", position))
			}
			if strings.TrimSpace(product.ImageAlt) == "" {
				problems = append(problems, fmt.Errorf("%s: imageAlt is required when image is set", position))
			}
			if previous, ok := productImages[product.Image]; ok {
				problems = append(problems, fmt.Errorf("%s: image %q duplicates %q", position, product.Image, previous))
			} else {
				productImages[product.Image] = product.Name
			}
		}
	}

	zoneSlugs := make(map[string]string, len(d.DeliveryZones))
	for index, zone := range d.DeliveryZones {
		position := fmt.Sprintf("delivery_zones[%d]", index)
		if strings.TrimSpace(zone.Name) == "" {
			problems = append(problems, fmt.Errorf("%s: name is required", position))
		}
		slug := Slug(zone.Name)
		if slug == "" {
			problems = append(problems, fmt.Errorf("%s: name must produce a slug", position))
		} else if previous, ok := zoneSlugs[slug]; ok {
			problems = append(problems, fmt.Errorf("%s: slug %q duplicates %q", position, slug, previous))
		} else {
			zoneSlugs[slug] = zone.Name
		}
		if _, err := CentsFromReais(zone.Fee.String()); err != nil {
			problems = append(problems, fmt.Errorf("%s: invalid fee: %w", position, err))
		}
	}

	return errors.Join(problems...)
}

// Build valida o seed e monta o plano de escrita de forma determinística.
func Build(data Data) (Plan, error) {
	if err := data.Validate(); err != nil {
		return Plan{}, err
	}

	plan := Plan{
		Store: StoreSettings{
			Name:         data.Store.Name,
			IsOpen:       data.Store.IsOpen,
			DeliveryTime: data.Store.DeliveryTime,
			WhatsApp:     data.Store.WhatsApp,
			Instagram:    data.Store.Instagram,
			Address:      data.Store.Address,
			OpeningHours: cloneOpeningHours(data.Store.OpeningHours),
		},
	}

	categorySlugByID := make(map[string]string, len(data.Categories))
	for index, category := range data.Categories {
		slug := Slug(category.Name)
		categorySlugByID[category.ID] = slug
		plan.Categories = append(plan.Categories, CategoryDocument{
			Slug:      slug,
			Name:      category.Name,
			Active:    category.Active,
			SortOrder: index,
		})
	}

	for index, product := range data.Products {
		priceCents, err := CentsFromReais(product.Price.String())
		if err != nil {
			return Plan{}, fmt.Errorf("product %q: %w", product.Name, err)
		}
		plan.Products = append(plan.Products, ProductDocument{
			Slug:         Slug(product.Name),
			Name:         product.Name,
			Description:  product.Desc,
			PriceCents:   priceCents,
			CategorySlug: categorySlugByID[product.CategoryID],
			ImageURL:     product.Image,
			ImageAlt:     product.ImageAlt,
			Available:    product.Available,
			Featured:     product.Featured,
			Active:       true,
			SortOrder:    index,
		})
	}

	for index, zone := range data.DeliveryZones {
		feeCents, err := CentsFromReais(zone.Fee.String())
		if err != nil {
			return Plan{}, fmt.Errorf("delivery zone %q: %w", zone.Name, err)
		}
		plan.Zones = append(plan.Zones, ZoneDocument{
			Slug:      Slug(zone.Name),
			Name:      zone.Name,
			FeeCents:  feeCents,
			Active:    zone.Active,
			SortOrder: index,
		})
	}

	return plan, nil
}

// ForEnvironment ajusta o plano para testes E2E determinísticos: sem horários e
// com o kill switch ligado, o status da loja não depende da hora da execução.
func ForEnvironment(plan Plan, environment string) Plan {
	if !strings.EqualFold(strings.TrimSpace(environment), "test") {
		return plan
	}
	plan.Store.IsOpen = true
	plan.Store.OpeningHours = nil
	return plan
}

func cloneOpeningHours(hours []domain.OpeningHours) []domain.OpeningHours {
	if hours == nil {
		return nil
	}
	copied := make([]domain.OpeningHours, len(hours))
	for index, day := range hours {
		copied[index] = day
		if day.Intervals != nil {
			copied[index].Intervals = append([]domain.OpeningInterval(nil), day.Intervals...)
		}
	}
	return copied
}

// CentsFromReais converte um valor em reais, sempre em texto, para centavos.
// A conversão nunca usa ponto flutuante.
func CentsFromReais(raw string) (int64, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return 0, errors.New("money value is empty")
	}
	if strings.HasPrefix(value, "-") {
		return 0, fmt.Errorf("money value %q is negative", value)
	}

	parts := strings.Split(value, ".")
	if len(parts) > 2 {
		return 0, fmt.Errorf("money value %q has more than one decimal separator", value)
	}
	if parts[0] == "" || !onlyDigits(parts[0]) {
		return 0, fmt.Errorf("money value %q must start with digits", value)
	}

	var cents int64
	if len(parts) == 2 {
		fraction := parts[1]
		if fraction == "" || len(fraction) > 2 || !onlyDigits(fraction) {
			return 0, fmt.Errorf("money value %q must have at most two decimal digits", value)
		}
		if len(fraction) == 1 {
			fraction += "0"
		}
		parsed, err := strconv.ParseInt(fraction, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("parse cents of %q: %w", value, err)
		}
		cents = parsed
	}

	units, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse reais of %q: %w", value, err)
	}
	if units > (math.MaxInt64-cents)/100 {
		return 0, fmt.Errorf("money value %q overflows int64 cents", value)
	}

	return units*100 + cents, nil
}

// Slug gera um identificador estável em minúsculas, sem acentos e separado por
// hífen. Apóstrofos são removidos para preservar nomes como "Brocado's".
func Slug(value string) string {
	lowered := accentReplacer.Replace(strings.ToLower(strings.TrimSpace(value)))

	var builder strings.Builder
	lastWasSeparator := false
	for _, character := range lowered {
		switch {
		case character >= 'a' && character <= 'z', character >= '0' && character <= '9':
			builder.WriteRune(character)
			lastWasSeparator = false
		case character == '\'' || character == '’':
			// Apóstrofo não separa palavras.
		default:
			if builder.Len() > 0 && !lastWasSeparator {
				builder.WriteByte('-')
				lastWasSeparator = true
			}
		}
	}

	return strings.Trim(builder.String(), "-")
}

// ValidateTarget impede que o seed rode contra banco de produção.
func ValidateTarget(database, environment string) error {
	if strings.TrimSpace(database) == "" {
		return errors.New("database name is required")
	}
	if environment == "production" {
		return errors.New("refusing to seed with APP_ENV=production")
	}
	if database == "brocados" {
		return fmt.Errorf("refusing to seed production database %q", database)
	}
	if !strings.HasSuffix(database, "_dev") && !strings.HasSuffix(database, "_test") {
		return fmt.Errorf("refusing to seed database %q: name must end with _dev or _test", database)
	}
	return nil
}

func onlyDigits(value string) bool {
	for _, character := range value {
		if character < '0' || character > '9' {
			return false
		}
	}
	return true
}

var accentReplacer = strings.NewReplacer(
	"á", "a", "à", "a", "â", "a", "ã", "a", "ä", "a",
	"é", "e", "è", "e", "ê", "e", "ë", "e",
	"í", "i", "ì", "i", "î", "i", "ï", "i",
	"ó", "o", "ò", "o", "ô", "o", "õ", "o", "ö", "o",
	"ú", "u", "ù", "u", "û", "u", "ü", "u",
	"ç", "c", "ñ", "n", "ý", "y", "ÿ", "y",
)
