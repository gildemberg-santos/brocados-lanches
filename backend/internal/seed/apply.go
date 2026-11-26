package seed

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/brocados-lanches/api/internal/domain"
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

// ApplyOptions controla a escrita do plano sem alterar a validação de alvo do
// chamador. MarkerKey e Marker substituem o marcador padrão de desenvolvimento
// quando a migração versionada reaproveita o mesmo plano.
type ApplyOptions struct {
	Reset     bool
	MarkerKey string
	Marker    bson.D
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
	return ApplyWithOptions(ctx, database, plan, ApplyOptions{Reset: reset})
}

// ApplyWithOptions grava o plano sem validar o alvo: a validação é
// responsabilidade da migração que o chama. É usada apenas para reaproveitar o
// plano determinístico do seed em migrações versionadas, com outro marcador.
func ApplyWithOptions(ctx context.Context, database *mongo.Database, plan Plan, options ApplyOptions) (Result, error) {
	if database == nil {
		return Result{}, errors.New("mongo database is required")
	}
	if plan.Store.Name == "" {
		return Result{}, errors.New("seed plan has no store name")
	}

	markerKey := options.MarkerKey
	if markerKey == "" {
		markerKey = "seed"
	}
	marker := options.Marker
	if len(marker) == 0 {
		marker = seedMarker
	}

	now := time.Now().UTC()
	result := Result{Database: database.Name(), Reset: options.Reset}

	if options.Reset {
		for _, collection := range seededCollections {
			if _, err := database.Collection(collection).DeleteMany(ctx, bson.D{}); err != nil {
				return Result{}, fmt.Errorf("reset collection %s: %w", collection, err)
			}
		}
	}

	storeCount, err := applyStoreSettings(ctx, database, plan.Store, now, markerKey, marker)
	if err != nil {
		return Result{}, err
	}
	result.StoreSettings = storeCount

	for _, category := range plan.Categories {
		count, err := applyCategory(ctx, database, category, now, markerKey, marker)
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
		count, err := applyProduct(ctx, database, product, categoryID, now, markerKey, marker)
		if err != nil {
			return Result{}, err
		}
		result.Products += count
	}

	for _, zone := range plan.Zones {
		count, err := applyZone(ctx, database, zone, now, markerKey, marker)
		if err != nil {
			return Result{}, err
		}
		result.DeliveryZones += count
	}

	return result, nil
}

func applyStoreSettings(ctx context.Context, database *mongo.Database, store StoreSettings, now time.Time, markerKey string, marker bson.D) (int, error) {
	fields := bson.D{
		{Key: "name", Value: store.Name},
		{Key: "isOpen", Value: store.IsOpen},
		{Key: "deliveryTime", Value: store.DeliveryTime},
		{Key: "whatsapp", Value: store.WhatsApp},
		{Key: "instagram", Value: store.Instagram},
		{Key: "address", Value: store.Address},
		{Key: "openingHours", Value: openingHoursBSON(store.OpeningHours)},
	}

	count, err := applyDocument(ctx, database, CollectionStoreSettings, bson.D{{Key: "_id", Value: "main"}}, fields, now, markerKey, marker)
	if err != nil {
		return 0, fmt.Errorf("seed store settings: %w", err)
	}
	return count, nil
}

func openingHoursBSON(hours []domain.OpeningHours) bson.A {
	result := make(bson.A, 0, len(hours))
	for _, day := range hours {
		intervals := make(bson.A, 0, len(day.Intervals))
		for _, interval := range day.Intervals {
			intervals = append(intervals, bson.D{
				{Key: "open", Value: interval.Open},
				{Key: "close", Value: interval.Close},
			})
		}
		result = append(result, bson.D{
			{Key: "day", Value: day.Day},
			{Key: "closed", Value: day.Closed},
			{Key: "intervals", Value: intervals},
		})
	}
	return result
}

func applyCategory(ctx context.Context, database *mongo.Database, category CategoryDocument, now time.Time, markerKey string, marker bson.D) (int, error) {
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
		markerKey,
		marker,
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
	markerKey string,
	marker bson.D,
) (int, error) {
	fields := bson.D{
		{Key: "slug", Value: product.Slug},
		{Key: "name", Value: product.Name},
		{Key: "description", Value: product.Description},
		{Key: "priceCents", Value: product.PriceCents},
		{Key: "categoryId", Value: categoryID},
		{Key: "imageUrl", Value: product.ImageURL},
		{Key: "imageAlt", Value: product.ImageAlt},
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
		markerKey,
		marker,
	)
	if err != nil {
		return 0, fmt.Errorf("seed product %q: %w", product.Slug, err)
	}
	return count, nil
}

func applyZone(ctx context.Context, database *mongo.Database, zone ZoneDocument, now time.Time, markerKey string, marker bson.D) (int, error) {
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
		markerKey,
		marker,
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
	markerKey string,
	marker bson.D,
) (int, error) {
	set := make(bson.D, 0, len(fields)+2)
	set = append(set, fields...)
	set = append(set, bson.E{Key: markerKey, Value: marker}, bson.E{Key: "updatedAt", Value: now})

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

func loadCategoryIDs(ctx context.Context, database *mongo.Database) (rows map[string]bson.ObjectID, err error) {
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
	defer func() {
		err = errors.Join(err, cursor.Close(ctx))
	}()

	rows = make(map[string]bson.ObjectID)
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
