package repository

import (
	"fmt"

	"github.com/brocados-lanches/api/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func newCategoryDocument(category domain.Category) (categoryDocument, error) {
	id, err := bson.ObjectIDFromHex(category.ID)
	if err != nil {
		return categoryDocument{}, fmt.Errorf("invalid category id: %w", err)
	}
	return categoryDocument{
		ID:          id,
		Name:        category.Name,
		Slug:        category.Slug,
		Description: category.Description,
		SortOrder:   category.SortOrder,
		Active:      category.Active,
		CreatedAt:   category.CreatedAt.UTC(),
		UpdatedAt:   category.UpdatedAt.UTC(),
	}, nil
}

func newProductDocument(product domain.Product) (productDocument, error) {
	id, err := bson.ObjectIDFromHex(product.ID)
	if err != nil {
		return productDocument{}, fmt.Errorf("invalid product id: %w", err)
	}
	categoryID, err := bson.ObjectIDFromHex(product.CategoryID)
	if err != nil {
		return productDocument{}, fmt.Errorf("invalid product category id: %w", err)
	}
	addonIDs := make([]bson.ObjectID, 0, len(product.AddonIDs))
	for _, raw := range product.AddonIDs {
		addonID, err := bson.ObjectIDFromHex(raw)
		if err != nil {
			return productDocument{}, fmt.Errorf("invalid product addon id: %w", err)
		}
		addonIDs = append(addonIDs, addonID)
	}

	return productDocument{
		ID:           id,
		CategoryID:   categoryID,
		Name:         product.Name,
		Slug:         product.Slug,
		Description:  product.Description,
		PriceCents:   product.PriceCents,
		ImageURL:     product.ImageURL,
		ImageAlt:     product.ImageAlt,
		Allergens:    append([]string(nil), product.Allergens...),
		Available:    product.Available,
		Featured:     product.Featured,
		Active:       product.Active,
		SortOrder:    product.SortOrder,
		OptionGroups: newOptionGroupDocuments(product.OptionGroups),
		AddonIDs:     addonIDs,
		CreatedAt:    product.CreatedAt.UTC(),
		UpdatedAt:    product.UpdatedAt.UTC(),
	}, nil
}

func newOptionGroupDocuments(groups []domain.OptionGroup) []optionGroupDocument {
	if len(groups) == 0 {
		return []optionGroupDocument{}
	}
	documents := make([]optionGroupDocument, len(groups))
	for index, group := range groups {
		options := make([]optionDocument, len(group.Options))
		for optionIndex, option := range group.Options {
			options[optionIndex] = optionDocument{
				ID:         option.ID,
				Name:       option.Name,
				PriceCents: option.PriceCents,
				Available:  option.Available,
				SortOrder:  option.SortOrder,
			}
		}
		documents[index] = optionGroupDocument{
			ID:            group.ID,
			Name:          group.Name,
			Required:      group.Required,
			MinSelections: group.MinSelections,
			MaxSelections: group.MaxSelections,
			Options:       options,
		}
	}
	return documents
}

func newAddonDocument(addon domain.Addon) (addonDocument, error) {
	id, err := bson.ObjectIDFromHex(addon.ID)
	if err != nil {
		return addonDocument{}, fmt.Errorf("invalid addon id: %w", err)
	}
	return addonDocument{
		ID:         id,
		Name:       addon.Name,
		Slug:       addon.Slug,
		PriceCents: addon.PriceCents,
		Active:     addon.Active,
		SortOrder:  addon.SortOrder,
		CreatedAt:  addon.CreatedAt.UTC(),
		UpdatedAt:  addon.UpdatedAt.UTC(),
	}, nil
}

func newDeliveryZoneDocument(zone domain.DeliveryZone) (deliveryZoneDocument, error) {
	id, err := bson.ObjectIDFromHex(zone.ID)
	if err != nil {
		return deliveryZoneDocument{}, fmt.Errorf("invalid delivery zone id: %w", err)
	}
	return deliveryZoneDocument{
		ID:        id,
		Name:      zone.Name,
		Slug:      zone.Slug,
		FeeCents:  zone.FeeCents,
		Active:    zone.Active,
		SortOrder: zone.SortOrder,
		CreatedAt: zone.CreatedAt.UTC(),
		UpdatedAt: zone.UpdatedAt.UTC(),
	}, nil
}
