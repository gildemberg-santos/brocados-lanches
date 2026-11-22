package repository

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type catalogIndexSet struct {
	collection string
	indexes    []mongo.IndexModel
}

// EnsureCatalogIndexes cria os índices de catálogo de docs/05-modelo-de-dados.md
// de forma idempotente. É chamado na inicialização da API e nos testes de integração.
func EnsureCatalogIndexes(ctx context.Context, database *mongo.Database) error {
	if database == nil {
		return errors.New("mongo database is required")
	}

	sets := []catalogIndexSet{
		{
			collection: CollectionCategories,
			indexes: []mongo.IndexModel{
				{
					Keys:    bson.D{{Key: "slug", Value: 1}},
					Options: options.Index().SetName("categories_slug_unique").SetUnique(true),
				},
				{
					Keys:    bson.D{{Key: "active", Value: 1}, {Key: "sortOrder", Value: 1}},
					Options: options.Index().SetName("categories_active_sort"),
				},
			},
		},
		{
			collection: CollectionProducts,
			indexes: []mongo.IndexModel{
				{
					Keys:    bson.D{{Key: "slug", Value: 1}},
					Options: options.Index().SetName("products_slug_unique").SetUnique(true),
				},
				{
					Keys: bson.D{
						{Key: "categoryId", Value: 1},
						{Key: "active", Value: 1},
						{Key: "sortOrder", Value: 1},
					},
					Options: options.Index().SetName("products_category_active_sort"),
				},
				{
					Keys:    bson.D{{Key: "active", Value: 1}, {Key: "featured", Value: 1}},
					Options: options.Index().SetName("products_active_featured"),
				},
			},
		},
		{
			collection: CollectionAddons,
			indexes: []mongo.IndexModel{
				{
					Keys:    bson.D{{Key: "slug", Value: 1}},
					Options: options.Index().SetName("addons_slug_unique").SetUnique(true),
				},
				{
					Keys:    bson.D{{Key: "active", Value: 1}, {Key: "sortOrder", Value: 1}},
					Options: options.Index().SetName("addons_active_sort"),
				},
			},
		},
		{
			collection: CollectionDeliveryZones,
			indexes: []mongo.IndexModel{
				{
					Keys:    bson.D{{Key: "slug", Value: 1}},
					Options: options.Index().SetName("delivery_zones_slug_unique").SetUnique(true),
				},
				{
					Keys:    bson.D{{Key: "active", Value: 1}, {Key: "name", Value: 1}},
					Options: options.Index().SetName("delivery_zones_active_name"),
				},
			},
		},
	}

	for _, set := range sets {
		if _, err := database.Collection(set.collection).Indexes().CreateMany(ctx, set.indexes); err != nil {
			return fmt.Errorf("ensure indexes on %s: %w", set.collection, err)
		}
	}

	return nil
}
