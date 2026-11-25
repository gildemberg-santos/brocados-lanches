package repository

import (
	"context"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func TestEnsureCatalogIndexesRequiresDatabase(t *testing.T) {
	if err := EnsureCatalogIndexes(context.Background(), nil); err == nil {
		t.Fatal("EnsureCatalogIndexes(nil) error = nil, want error")
	}
}

func TestEnsureCatalogIndexesIsIdempotent(t *testing.T) {
	ctx := testContext(t)
	database := openCatalogTestDatabase(t)

	if err := EnsureCatalogIndexes(ctx, database); err != nil {
		t.Fatalf("second EnsureCatalogIndexes() error = %v", err)
	}

	expected := map[string][]string{
		CollectionCategories: {
			"categories_slug_unique",
			"categories_active_sort",
		},
		CollectionProducts: {
			"products_slug_unique",
			"products_category_active_sort",
			"products_active_featured",
		},
		CollectionAddons: {
			"addons_slug_unique",
			"addons_active_sort",
		},
		CollectionDeliveryZones: {
			"delivery_zones_slug_unique",
			"delivery_zones_active_name",
		},
	}

	for collection, names := range expected {
		indexes := listIndexes(t, ctx, database.Collection(collection))
		for _, name := range names {
			if _, ok := indexes[name]; !ok {
				t.Errorf("collection %s is missing index %q", collection, name)
			}
		}
	}

	unique := map[string]string{
		CollectionCategories:    "categories_slug_unique",
		CollectionProducts:      "products_slug_unique",
		CollectionAddons:        "addons_slug_unique",
		CollectionDeliveryZones: "delivery_zones_slug_unique",
	}
	for collection, name := range unique {
		index, ok := listIndexes(t, ctx, database.Collection(collection))[name]
		if !ok {
			continue
		}
		isUnique, _ := index["unique"].(bool)
		if !isUnique {
			t.Errorf("index %s.%s unique = false, want true", collection, name)
		}
	}
}

func TestCatalogUniqueSlugIndexes(t *testing.T) {
	ctx := testContext(t)
	database := openCatalogTestDatabase(t)

	tests := []struct {
		name       string
		collection string
		document   func() any
	}{
		{
			name:       "categories",
			collection: CollectionCategories,
			document: func() any {
				return categoryDocument{Name: "Categoria", Slug: "duplicada", Active: true}
			},
		},
		{
			name:       "products",
			collection: CollectionProducts,
			document: func() any {
				return productDocument{
					CategoryID:  bson.NewObjectID(),
					Name:        "Produto",
					Slug:        "duplicado",
					Description: "Descrição",
					PriceCents:  1000,
					Active:      true,
				}
			},
		},
		{
			name:       "addons",
			collection: CollectionAddons,
			document: func() any {
				return addonDocument{Name: "Adicional", Slug: "duplicado", PriceCents: 500, Active: true}
			},
		},
		{
			name:       "delivery zones",
			collection: CollectionDeliveryZones,
			document: func() any {
				return deliveryZoneDocument{Name: "Zona", Slug: "duplicada", FeeCents: 500, Active: true}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			collection := database.Collection(test.collection)

			if _, err := collection.InsertOne(ctx, test.document()); err != nil {
				t.Fatalf("first InsertOne() error = %v", err)
			}
			if _, err := collection.InsertOne(ctx, test.document()); !mongo.IsDuplicateKeyError(err) {
				t.Fatalf("second InsertOne() error = %v, want duplicate key error", err)
			}
		})
	}
}

func listIndexes(t *testing.T, ctx context.Context, collection *mongo.Collection) map[string]bson.M {
	t.Helper()

	cursor, err := collection.Indexes().List(ctx)
	if err != nil {
		t.Fatalf("Indexes().List() error = %v", err)
	}
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			t.Errorf("close index cursor: %v", err)
		}
	}()

	var indexes []bson.M
	if err := cursor.All(ctx, &indexes); err != nil {
		t.Fatalf("decode indexes: %v", err)
	}

	result := make(map[string]bson.M, len(indexes))
	for _, index := range indexes {
		name, _ := index["name"].(string)
		result[name] = index
	}
	return result
}
