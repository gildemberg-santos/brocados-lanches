package domain

import (
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

// Limites de entrada das entidades de catálogo. Protegem a API e o banco
// sem codificar nenhum valor de negócio.
const (
	MaxSlugLength          = 80
	MaxNameLength          = 120
	MaxDescriptionLength   = 2000
	MaxImageURLLength      = 500
	MaxImageAltLength      = 200
	MaxAllergenLength      = 60
	MaxIDLength            = 64
	MaxOptionIDLength      = MaxIDLength
	MaxOptionGroupIDLength = MaxIDLength
)

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// Category é uma seção do cardápio.
type Category struct {
	ID          string
	Name        string
	Slug        string
	Description string
	SortOrder   int
	Active      bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Product é um item vendável do cardápio. Os grupos de opções são embutidos
// porque são pequenos e pertencem ao produto.
type Product struct {
	ID           string
	CategoryID   string
	Name         string
	Slug         string
	Description  string
	PriceCents   int64
	ImageURL     string
	ImageAlt     string
	Allergens    []string
	Available    bool
	Featured     bool
	Active       bool
	SortOrder    int
	OptionGroups []OptionGroup
	AddonIDs     []string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// OptionGroup é um conjunto de escolhas obrigatórias ou opcionais.
type OptionGroup struct {
	ID            string
	Name          string
	Required      bool
	MinSelections int
	MaxSelections int
	Options       []Option
}

// Option é uma escolha dentro de um grupo de opções.
type Option struct {
	ID         string
	Name       string
	PriceCents int64
	Available  bool
	SortOrder  int
}

// Addon é um adicional reutilizável que os produtos vinculam por AddonIDs.
type Addon struct {
	ID         string
	Name       string
	Slug       string
	PriceCents int64
	Active     bool
	SortOrder  int
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// DeliveryZone é um bairro e sua taxa de entrega calculada no servidor.
type DeliveryZone struct {
	ID        string
	Name      string
	Slug      string
	FeeCents  int64
	Active    bool
	SortOrder int
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Validate confere as invariantes da categoria.
func (c Category) Validate() error {
	if err := validateRequiredString("name", c.Name, MaxNameLength); err != nil {
		return err
	}
	if err := ValidateSlug("slug", c.Slug); err != nil {
		return err
	}
	if strings.TrimSpace(c.Description) != "" {
		if err := validateMaxLength("description", c.Description, MaxDescriptionLength); err != nil {
			return err
		}
	}
	return validateSortOrder("sortOrder", c.SortOrder)
}

// Validate confere as invariantes do produto, incluindo grupos de opções
// embutidos e o formato dos vínculos com adicionais. A existência do adicional
// é uma regra do service/repositório.
func (p Product) Validate() error {
	if err := validateRequiredString("categoryId", p.CategoryID, MaxIDLength); err != nil {
		return err
	}
	if err := validateRequiredString("name", p.Name, MaxNameLength); err != nil {
		return err
	}
	if err := ValidateSlug("slug", p.Slug); err != nil {
		return err
	}
	if err := validateRequiredString("description", p.Description, MaxDescriptionLength); err != nil {
		return err
	}
	if err := validateMoneyCents("priceCents", p.PriceCents); err != nil {
		return err
	}
	if strings.TrimSpace(p.ImageURL) != "" {
		if err := validateMaxLength("imageUrl", p.ImageURL, MaxImageURLLength); err != nil {
			return err
		}
		if err := validateRequiredString("imageAlt", p.ImageAlt, MaxImageAltLength); err != nil {
			return err
		}
	} else if strings.TrimSpace(p.ImageAlt) != "" {
		if err := validateMaxLength("imageAlt", p.ImageAlt, MaxImageAltLength); err != nil {
			return err
		}
	}
	if err := validateSortOrder("sortOrder", p.SortOrder); err != nil {
		return err
	}
	if err := validateAllergens(p.Allergens); err != nil {
		return err
	}
	if err := validateOptionGroups(p.OptionGroups); err != nil {
		return err
	}
	return validateAddonIDs(p.AddonIDs)
}

// Validate confere as invariantes do grupo de opções.
func (g OptionGroup) Validate() error {
	if err := validateRequiredString("id", g.ID, MaxOptionGroupIDLength); err != nil {
		return err
	}
	if err := validateRequiredString("name", g.Name, MaxNameLength); err != nil {
		return err
	}
	if g.MinSelections < 0 {
		return invalid("minSelections", CodeOutOfRange, "must be greater than or equal to zero")
	}
	if g.MaxSelections < 1 {
		return invalid("maxSelections", CodeOutOfRange, "must be greater than or equal to one")
	}
	if g.MinSelections > g.MaxSelections {
		return invalid("minSelections", CodeInvalid, "must be less than or equal to maxSelections")
	}
	if g.Required && g.MinSelections < 1 {
		return invalid("minSelections", CodeInvalid, "must be at least one when the group is required")
	}
	if len(g.Options) == 0 {
		return invalid("options", CodeRequired, "at least one option is required")
	}
	if g.MaxSelections > len(g.Options) {
		return invalid("maxSelections", CodeOutOfRange, "cannot exceed the number of options")
	}

	seen := make(map[string]struct{}, len(g.Options))
	for index, option := range g.Options {
		if err := option.Validate(); err != nil {
			return prefixValidationError(fmt.Sprintf("options[%d]", index), err)
		}
		if _, exists := seen[option.ID]; exists {
			return invalid(fmt.Sprintf("options[%d].id", index), CodeDuplicate, "duplicates another option")
		}
		seen[option.ID] = struct{}{}
	}

	return nil
}

// Validate confere as invariantes da opção.
func (o Option) Validate() error {
	if err := validateRequiredString("id", o.ID, MaxOptionIDLength); err != nil {
		return err
	}
	if err := validateRequiredString("name", o.Name, MaxNameLength); err != nil {
		return err
	}
	if err := validateMoneyCents("priceCents", o.PriceCents); err != nil {
		return err
	}
	return validateSortOrder("sortOrder", o.SortOrder)
}

// Validate confere as invariantes do adicional.
func (a Addon) Validate() error {
	if err := validateRequiredString("name", a.Name, MaxNameLength); err != nil {
		return err
	}
	if err := ValidateSlug("slug", a.Slug); err != nil {
		return err
	}
	if err := validateMoneyCents("priceCents", a.PriceCents); err != nil {
		return err
	}
	return validateSortOrder("sortOrder", a.SortOrder)
}

// Validate confere as invariantes da zona de entrega.
func (z DeliveryZone) Validate() error {
	if err := validateRequiredString("name", z.Name, MaxNameLength); err != nil {
		return err
	}
	if err := ValidateSlug("slug", z.Slug); err != nil {
		return err
	}
	if err := validateMoneyCents("feeCents", z.FeeCents); err != nil {
		return err
	}
	return validateSortOrder("sortOrder", z.SortOrder)
}

// ValidateSlug confere se o slug é seguro para URL, minúsculo e sem segmentos
// vazios. A mesma regra vale para todas as entidades de catálogo.
func ValidateSlug(field, slug string) error {
	if err := validateRequiredString(field, slug, MaxSlugLength); err != nil {
		return err
	}
	if !slugPattern.MatchString(slug) {
		return invalid(field, CodeInvalid, "must contain only lowercase letters, digits and single hyphens")
	}
	return nil
}

func validateRequiredString(field, value string, maxLength int) error {
	if strings.TrimSpace(value) == "" {
		return invalid(field, CodeRequired, "is required")
	}
	return validateMaxLength(field, value, maxLength)
}

func validateMaxLength(field, value string, maxLength int) error {
	if utf8.RuneCountInString(value) > maxLength {
		return invalid(field, CodeTooLong, fmt.Sprintf("must have at most %d characters", maxLength))
	}
	return nil
}

func validateMoneyCents(field string, value int64) error {
	if value < 0 {
		return invalid(field, CodeOutOfRange, "must be greater than or equal to zero")
	}
	return nil
}

func validateSortOrder(field string, value int) error {
	if value < 0 {
		return invalid(field, CodeOutOfRange, "must be greater than or equal to zero")
	}
	return nil
}

func validateAllergens(allergens []string) error {
	seen := make(map[string]struct{}, len(allergens))
	for index, allergen := range allergens {
		field := fmt.Sprintf("allergens[%d]", index)
		if err := validateRequiredString(field, allergen, MaxAllergenLength); err != nil {
			return err
		}

		normalized := strings.ToLower(strings.TrimSpace(allergen))
		if _, exists := seen[normalized]; exists {
			return invalid(field, CodeDuplicate, "duplicates another allergen")
		}
		seen[normalized] = struct{}{}
	}
	return nil
}

func validateOptionGroups(groups []OptionGroup) error {
	seen := make(map[string]struct{}, len(groups))
	for index, group := range groups {
		if err := group.Validate(); err != nil {
			return prefixValidationError(fmt.Sprintf("optionGroups[%d]", index), err)
		}
		if _, exists := seen[group.ID]; exists {
			return invalid(fmt.Sprintf("optionGroups[%d].id", index), CodeDuplicate, "duplicates another option group")
		}
		seen[group.ID] = struct{}{}
	}
	return nil
}

func validateAddonIDs(addonIDs []string) error {
	seen := make(map[string]struct{}, len(addonIDs))
	for index, addonID := range addonIDs {
		field := fmt.Sprintf("addonIds[%d]", index)
		if err := validateRequiredString(field, addonID, MaxIDLength); err != nil {
			return err
		}
		if _, exists := seen[addonID]; exists {
			return invalid(field, CodeDuplicate, "duplicates another addon")
		}
		seen[addonID] = struct{}{}
	}
	return nil
}
