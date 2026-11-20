package seed

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Coleções tocadas pelo seed. O reset apaga apenas estas coleções, nunca
// usuários, sessões, pedidos ou auditoria.
const (
	CollectionStoreSettings = "store_settings"
	CollectionCategories    = "categories"
	CollectionProducts      = "products"
	CollectionDeliveryZones = "delivery_zones"
)

// Result resume o que foi aplicado.
type Result struct {
	Database      string
	Reset         bool
	StoreSettings int
	Categories    int
	Products      int
	DeliveryZones int
}

var seedMarker = bson.D{
	{Key: "fictitious", Value: true},
	{Key: "source", Value: "design-system/seed.json"},
}

var seededCollections = []string{
	CollectionCategories,
	CollectionProducts,
	CollectionDeliveryZones,
	CollectionStoreSettings,
}

// Apply grava o plano no banco informado. A operação é idempotente: rodar duas
// vezes sem reset mantém as mesmas quantidades de documentos. Com reset, as
// coleções semeadas são esvaziadas antes da escrita.
func Apply(ctx context.Context, database *mongo.Database, plan Plan, reset bool) (Result, error) {
	if database == nil {
		return Result{}, errors.New("mongo database is required")
	}
	if err := ValidateTarget(database.Name(), ""); err != nil {
		return Result{}, err
	}
	if plan.Store.Name == "" {
		return Result{}, errors.New("seed plan has no store name")
	}

	now := time.Now().UTC()
	result := Result{Database: database.Name(), Reset: reset}

	if reset {
		for _, collection := range seededCollections {
			if _, err := database.Collection(collection).DeleteMany(ctx, bson.D{}); err != nil {
				return Result{}, fmt.Errorf("reset collection %s: %w", collection, err)
			}
		}
	}

	storeCount, err := applyStoreSettings(ctx, database, plan.Store, now)
	if err != nil {
		return Result{}, err
	}
	result.StoreSettings = storeCount

	for _, category := range plan.Categories {
		count, err := applyCategory(ctx, database, category, now)
		if err != nil {
			return Result{}, err
		}
		result.Categories += count
	}

	categoryIDs, err := loadCategoryIDs(ctx, database)
	if err != nil {
		return Result{}, err
	}

	for _, product := range plan.Products {
		categoryID, ok := categoryIDs[product.CategorySlug]
		if !ok {
			return Result{}, fmt.Errorf("category %q is missing after seed", product.CategorySlug)
		}
		count, err := applyProduct(ctx, database, product, categoryID, now)
		if err != nil {
			return Result{}, err
		}
		result.Products += count
	}

	for _, zone := range plan.Zones {
		count, err := applyZone(ctx, database, zone, now)
		if err != nil {
			return Result{}, err
		}
		result.DeliveryZones += count
	}

	return result, nil
}

func applyStoreSettings(ctx context.Context, database *mongo.Database, store StoreSettings, now time.Time) (int, error) {
	fields := bson.D{
		{Key: "name", Value: store.Name},
		{Key: "isOpen", Value: store.IsOpen},
		{Key: "deliveryTime", Value: store.DeliveryTime},
		{Key: "whatsapp", Value: store.WhatsApp},
		{Key: "instagram", Value: store.Instagram},
		{Key: "address", Value: store.Address},
	}

	count, err := applyDocument(ctx, database, CollectionStoreSettings, bson.D{{Key: "_id", Value: "main"}}, fields, now)
	if err != nil {
		return 0, fmt.Errorf("seed store settings: %w", err)
	}
	return count, nil
}

func applyCategory(ctx context.Context, database *mongo.Database, category CategoryDocument, now time.Time) (int, error) {
	fields := bson.D{
		{Key: "slug", Value: category.Slug},
		{Key: "name", Value: category.Name},
		{Key: "active", Value: category.Active},
		{Key: "sortOrder", Value: category.SortOrder},
	}

	count, err := applyDocument(
		ctx,
		database,
		CollectionCategories,
		bson.D{{Key: "slug", Value: category.Slug}},
		fields,
		now,
	)
	if err != nil {
		return 0, fmt.Errorf("seed category %q: %w", category.Slug, err)
	}
	return count, nil
}

func applyProduct(
	ctx context.Context,
	database *mongo.Database,
	product ProductDocument,
	categoryID bson.ObjectID,
	now time.Time,
) (int, error) {
	fields := bson.D{
		{Key: "slug", Value: product.Slug},
		{Key: "name", Value: product.Name},
		{Key: "description", Value: product.Description},
		{Key: "priceCents", Value: product.PriceCents},
		{Key: "categoryId", Value: categoryID},
		{Key: "imageUrl", Value: ""},
		{Key: "imageAlt", Value: ""},
		{Key: "allergens", Value: bson.A{}},
		{Key: "available", Value: product.Available},
		{Key: "featured", Value: product.Featured},
		{Key: "active", Value: product.Active},
		{Key: "sortOrder", Value: product.SortOrder},
		{Key: "optionGroups", Value: bson.A{}},
		{Key: "addonIds", Value: bson.A{}},
	}

	count, err := applyDocument(
		ctx,
		database,
		CollectionProducts,
		bson.D{{Key: "slug", Value: product.Slug}},
		fields,
		now,
	)
	if err != nil {
		return 0, fmt.Errorf("seed product %q: %w", product.Slug, err)
	}
	return count, nil
}

func applyZone(ctx context.Context, database *mongo.Database, zone ZoneDocument, now time.Time) (int, error) {
	fields := bson.D{
		{Key: "slug", Value: zone.Slug},
		{Key: "name", Value: zone.Name},
		{Key: "feeCents", Value: zone.FeeCents},
		{Key: "active", Value: zone.Active},
		{Key: "sortOrder", Value: zone.SortOrder},
	}

	count, err := applyDocument(
		ctx,
		database,
		CollectionDeliveryZones,
		bson.D{{Key: "slug", Value: zone.Slug}},
		fields,
		now,
	)
	if err != nil {
		return 0, fmt.Errorf("seed delivery zone %q: %w", zone.Slug, err)
	}
	return count, nil
}

func applyDocument(
	ctx context.Context,
	database *mongo.Database,
	collection string,
	filter bson.D,
	fields bson.D,
	now time.Time,
) (int, error) {
	set := make(bson.D, 0, len(fields)+2)
	set = append(set, fields...)
	set = append(set, bson.E{Key: "seed", Value: seedMarker}, bson.E{Key: "updatedAt", Value: now})

	update := bson.D{
		{Key: "$set", Value: set},
		{Key: "$setOnInsert", Value: bson.D{{Key: "createdAt", Value: now}}},
	}

	updateResult, err := database.Collection(collection).UpdateOne(
		ctx,
		filter,
		update,
		options.UpdateOne().SetUpsert(true),
	)
	if err != nil {
		return 0, err
	}

	// ModifiedCount é subconjunto de MatchedCount; somar os dois contaria duas vezes.
	return int(updateResult.MatchedCount + updateResult.UpsertedCount), nil
}

func loadCategoryIDs(ctx context.Context, database *mongo.Database) (map[string]bson.ObjectID, error) {
	cursor, err := database.Collection(CollectionCategories).Find(
		ctx,
		bson.D{},
		options.Find().SetProjection(bson.D{
			{Key: "_id", Value: 1},
			{Key: "slug", Value: 1},
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("load categories: %w", err)
	}
	defer cursor.Close(ctx)

	rows := make(map[string]bson.ObjectID)
	for cursor.Next(ctx) {
		var row struct {
			ID   bson.ObjectID `bson:"_id"`
			Slug string        `bson:"slug"`
		}
		if err := cursor.Decode(&row); err != nil {
			return nil, fmt.Errorf("decode category: %w", err)
		}
		rows[row.Slug] = row.ID
	}
	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("iterate categories: %w", err)
	}

	return rows, nil
}
