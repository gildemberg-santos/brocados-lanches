package repository

import (
	"time"

	"github.com/brocados-lanches/api/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// As structs de documento mantêm as preocupações de MongoDB/BSON fora do pacote
// de domínio. Elas espelham docs/05-modelo-de-dados.md.

type categoryDocument struct {
	ID          bson.ObjectID `bson:"_id,omitempty"`
	Name        string        `bson:"name"`
	Slug        string        `bson:"slug"`
	Description string        `bson:"description,omitempty"`
	SortOrder   int           `bson:"sortOrder"`
	Active      bool          `bson:"active"`
	CreatedAt   time.Time     `bson:"createdAt"`
	UpdatedAt   time.Time     `bson:"updatedAt"`
}

type productDocument struct {
	ID           bson.ObjectID         `bson:"_id,omitempty"`
	CategoryID   bson.ObjectID         `bson:"categoryId"`
	Name         string                `bson:"name"`
	Slug         string                `bson:"slug"`
	Description  string                `bson:"description"`
	PriceCents   int64                 `bson:"priceCents"`
	ImageURL     string                `bson:"imageUrl,omitempty"`
	ImageAlt     string                `bson:"imageAlt,omitempty"`
	Allergens    []string              `bson:"allergens"`
	Available    bool                  `bson:"available"`
	Featured     bool                  `bson:"featured"`
	Active       bool                  `bson:"active"`
	SortOrder    int                   `bson:"sortOrder"`
	OptionGroups []optionGroupDocument `bson:"optionGroups"`
	AddonIDs     []bson.ObjectID       `bson:"addonIds"`
	CreatedAt    time.Time             `bson:"createdAt"`
	UpdatedAt    time.Time             `bson:"updatedAt"`
}

type optionGroupDocument struct {
	ID            string           `bson:"id"`
	Name          string           `bson:"name"`
	Required      bool             `bson:"required"`
	MinSelections int              `bson:"minSelections"`
	MaxSelections int              `bson:"maxSelections"`
	Options       []optionDocument `bson:"options"`
}

type optionDocument struct {
	ID         string `bson:"id"`
	Name       string `bson:"name"`
	PriceCents int64  `bson:"priceCents"`
	Available  bool   `bson:"available"`
	SortOrder  int    `bson:"sortOrder"`
}

type addonDocument struct {
	ID         bson.ObjectID `bson:"_id,omitempty"`
	Name       string        `bson:"name"`
	Slug       string        `bson:"slug"`
	PriceCents int64         `bson:"priceCents"`
	Active     bool          `bson:"active"`
	SortOrder  int           `bson:"sortOrder"`
	CreatedAt  time.Time     `bson:"createdAt"`
	UpdatedAt  time.Time     `bson:"updatedAt"`
}

type deliveryZoneDocument struct {
	ID        bson.ObjectID `bson:"_id,omitempty"`
	Name      string        `bson:"name"`
	Slug      string        `bson:"slug"`
	FeeCents  int64         `bson:"feeCents"`
	Active    bool          `bson:"active"`
	SortOrder int           `bson:"sortOrder"`
	CreatedAt time.Time     `bson:"createdAt"`
	UpdatedAt time.Time     `bson:"updatedAt"`
}

func (d categoryDocument) toDomain() domain.Category {
	return domain.Category{
		ID:          d.ID.Hex(),
		Name:        d.Name,
		Slug:        d.Slug,
		Description: d.Description,
		SortOrder:   d.SortOrder,
		Active:      d.Active,
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
	}
}

func (d productDocument) toDomain() domain.Product {
	return domain.Product{
		ID:           d.ID.Hex(),
		CategoryID:   d.CategoryID.Hex(),
		Name:         d.Name,
		Slug:         d.Slug,
		Description:  d.Description,
		PriceCents:   d.PriceCents,
		ImageURL:     d.ImageURL,
		ImageAlt:     d.ImageAlt,
		Allergens:    d.Allergens,
		Available:    d.Available,
		Featured:     d.Featured,
		Active:       d.Active,
		SortOrder:    d.SortOrder,
		OptionGroups: optionGroupsToDomain(d.OptionGroups),
		AddonIDs:     addonIDsToDomain(d.AddonIDs),
		CreatedAt:    d.CreatedAt,
		UpdatedAt:    d.UpdatedAt,
	}
}

func (d addonDocument) toDomain() domain.Addon {
	return domain.Addon{
		ID:         d.ID.Hex(),
		Name:       d.Name,
		Slug:       d.Slug,
		PriceCents: d.PriceCents,
		Active:     d.Active,
		SortOrder:  d.SortOrder,
		CreatedAt:  d.CreatedAt,
		UpdatedAt:  d.UpdatedAt,
	}
}

func (d deliveryZoneDocument) toDomain() domain.DeliveryZone {
	return domain.DeliveryZone{
		ID:        d.ID.Hex(),
		Name:      d.Name,
		Slug:      d.Slug,
		FeeCents:  d.FeeCents,
		Active:    d.Active,
		SortOrder: d.SortOrder,
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
	}
}

func optionGroupsToDomain(documents []optionGroupDocument) []domain.OptionGroup {
	if len(documents) == 0 {
		return nil
	}

	groups := make([]domain.OptionGroup, len(documents))
	for index, document := range documents {
		groups[index] = domain.OptionGroup{
			ID:            document.ID,
			Name:          document.Name,
			Required:      document.Required,
			MinSelections: document.MinSelections,
			MaxSelections: document.MaxSelections,
			Options:       optionsToDomain(document.Options),
		}
	}
	return groups
}

func optionsToDomain(documents []optionDocument) []domain.Option {
	if len(documents) == 0 {
		return nil
	}

	options := make([]domain.Option, len(documents))
	for index, document := range documents {
		options[index] = domain.Option{
			ID:         document.ID,
			Name:       document.Name,
			PriceCents: document.PriceCents,
			Available:  document.Available,
			SortOrder:  document.SortOrder,
		}
	}
	return options
}

func addonIDsToDomain(ids []bson.ObjectID) []string {
	if len(ids) == 0 {
		return nil
	}

	result := make([]string, len(ids))
	for index, id := range ids {
		result[index] = id.Hex()
	}
	return result
}
