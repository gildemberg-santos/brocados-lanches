package repository

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/brocados-lanches/api/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

const catalogTestDatabase = "brocados_catalog_test"

func TestNewCatalogRepositoryRequiresDatabase(t *testing.T) {
	if _, err := NewCatalogRepository(nil); err == nil {
		t.Fatal("NewCatalogRepository(nil) error = nil, want error")
	}
}

func TestCatalogRepositoryListActiveCategories(t *testing.T) {
	ctx := testContext(t)
	database := openCatalogTestDatabase(t)
	repository := newCatalogRepository(t, database)

	first := categoryDocument{
		ID:        bson.NewObjectID(),
		Name:      "Hambúrgueres",
		Slug:      "hamburgueres",
		SortOrder: 1,
		Active:    true,
		CreatedAt: testTime(),
		UpdatedAt: testTime(),
	}
	second := categoryDocument{
		ID:        bson.NewObjectID(),
		Name:      "Bebidas",
		Slug:      "bebidas",
		SortOrder: 2,
		Active:    true,
		CreatedAt: testTime(),
		UpdatedAt: testTime(),
	}
	inactive := categoryDocument{
		ID:        bson.NewObjectID(),
		Name:      "Combos",
		Slug:      "combos",
		SortOrder: 0,
		Active:    false,
		CreatedAt: testTime(),
		UpdatedAt: testTime(),
	}
	insertDocument(t, ctx, database.Collection(CollectionCategories), first, second, inactive)

	categories, err := repository.ListActiveCategories(ctx)
	if err != nil {
		t.Fatalf("ListActiveCategories() error = %v", err)
	}
	if len(categories) != 2 {
		t.Fatalf("len(categories) = %d, want 2", len(categories))
	}
	if categories[0].ID != first.ID.Hex() || categories[1].ID != second.ID.Hex() {
		t.Fatalf("categories order = %q, %q", categories[0].Name, categories[1].Name)
	}
	if categories[0].Name != "Hambúrgueres" || categories[1].Name != "Bebidas" {
		t.Errorf("categories names = %q, %q", categories[0].Name, categories[1].Name)
	}
}

func TestCatalogRepositoryListActiveProducts(t *testing.T) {
	ctx := testContext(t)
	database := openCatalogTestDatabase(t)
	repository := newCatalogRepository(t, database)

	categoryID := bson.NewObjectID()
	addonID := bson.NewObjectID()
	product := productDocument{
		ID:          bson.NewObjectID(),
		CategoryID:  categoryID,
		Name:        "Brocado's Burguer Cheddar",
		Slug:        "brocados-burguer-cheddar",
		Description: "Blend artesanal 160g, cheddar e pão brioche.",
		PriceCents:  2300,
		ImageURL:    "https://cdn.example.com/burger.webp",
		ImageAlt:    "Hambúrguer com cheddar",
		Allergens:   []string{"glúten", "leite"},
		Available:   false,
		Featured:    true,
		Active:      true,
		SortOrder:   1,
		OptionGroups: []optionGroupDocument{
			{
				ID:            "doneness",
				Name:          "Ponto da carne",
				Required:      true,
				MinSelections: 1,
				MaxSelections: 1,
				Options: []optionDocument{
					{ID: "medium", Name: "Ao ponto", PriceCents: 0, Available: true, SortOrder: 1},
				},
			},
		},
		AddonIDs:  []bson.ObjectID{addonID},
		CreatedAt: testTime(),
		UpdatedAt: testTime(),
	}
	inactive := productDocument{
		ID:          bson.NewObjectID(),
		CategoryID:  categoryID,
		Name:        "Produto inativo",
		Slug:        "produto-inativo",
		Description: "Não deve aparecer.",
		PriceCents:  1000,
		Active:      false,
		CreatedAt:   testTime(),
		UpdatedAt:   testTime(),
	}
	insertDocument(t, ctx, database.Collection(CollectionProducts), product, inactive)

	products, err := repository.ListActiveProducts(ctx)
	if err != nil {
		t.Fatalf("ListActiveProducts() error = %v", err)
	}
	if len(products) != 1 {
		t.Fatalf("len(products) = %d, want 1", len(products))
	}

	got := products[0]
	if got.ID != product.ID.Hex() {
		t.Errorf("ID = %q, want %q", got.ID, product.ID.Hex())
	}
	if got.CategoryID != categoryID.Hex() {
		t.Errorf("CategoryID = %q, want %q", got.CategoryID, categoryID.Hex())
	}
	if got.PriceCents != 2300 {
		t.Errorf("PriceCents = %d, want 2300", got.PriceCents)
	}
	if got.Available {
		t.Error("Available = true, want false (sold out products must still be listed)")
	}
	if len(got.Allergens) != 2 || got.Allergens[0] != "glúten" {
		t.Errorf("Allergens = %#v", got.Allergens)
	}
	if len(got.OptionGroups) != 1 {
		t.Fatalf("len(OptionGroups) = %d, want 1", len(got.OptionGroups))
	}
	group := got.OptionGroups[0]
	if group.ID != "doneness" || group.MinSelections != 1 || group.MaxSelections != 1 {
		t.Errorf("OptionGroup = %+v", group)
	}
	if len(group.Options) != 1 || group.Options[0].ID != "medium" {
		t.Errorf("Options = %+v", group.Options)
	}
	if len(got.AddonIDs) != 1 || got.AddonIDs[0] != addonID.Hex() {
		t.Errorf("AddonIDs = %#v, want [%q]", got.AddonIDs, addonID.Hex())
	}
}

func TestCatalogRepositoryListActiveAddons(t *testing.T) {
	ctx := testContext(t)
	database := openCatalogTestDatabase(t)
	repository := newCatalogRepository(t, database)

	second := addonDocument{
		ID:         bson.NewObjectID(),
		Name:       "Bacon Crocante",
		Slug:       "bacon-crocante",
		PriceCents: 500,
		Active:     true,
		SortOrder:  2,
		CreatedAt:  testTime(),
		UpdatedAt:  testTime(),
	}
	first := addonDocument{
		ID:         bson.NewObjectID(),
		Name:       "Cheddar Extra",
		Slug:       "cheddar-extra",
		PriceCents: 300,
		Active:     true,
		SortOrder:  1,
		CreatedAt:  testTime(),
		UpdatedAt:  testTime(),
	}
	inactive := addonDocument{
		ID:         bson.NewObjectID(),
		Name:       "Cebola",
		Slug:       "cebola",
		PriceCents: 200,
		Active:     false,
		SortOrder:  0,
		CreatedAt:  testTime(),
		UpdatedAt:  testTime(),
	}
	insertDocument(t, ctx, database.Collection(CollectionAddons), second, first, inactive)

	addons, err := repository.ListActiveAddons(ctx)
	if err != nil {
		t.Fatalf("ListActiveAddons() error = %v", err)
	}
	if len(addons) != 2 {
		t.Fatalf("len(addons) = %d, want 2", len(addons))
	}
	if addons[0].ID != first.ID.Hex() || addons[1].ID != second.ID.Hex() {
		t.Errorf("addons order = %q, %q", addons[0].Name, addons[1].Name)
	}
	if addons[0].PriceCents != 300 || addons[1].PriceCents != 500 {
		t.Errorf("addons prices = %d, %d", addons[0].PriceCents, addons[1].PriceCents)
	}
}

func TestCatalogRepositoryListActiveDeliveryZones(t *testing.T) {
	ctx := testContext(t)
	database := openCatalogTestDatabase(t)
	repository := newCatalogRepository(t, database)

	second := deliveryZoneDocument{
		ID:        bson.NewObjectID(),
		Name:      "Centro",
		Slug:      "centro",
		FeeCents:  500,
		Active:    true,
		SortOrder: 2,
		CreatedAt: testTime(),
		UpdatedAt: testTime(),
	}
	first := deliveryZoneDocument{
		ID:        bson.NewObjectID(),
		Name:      "Prainha",
		Slug:      "prainha",
		FeeCents:  800,
		Active:    true,
		SortOrder: 1,
		CreatedAt: testTime(),
		UpdatedAt: testTime(),
	}
	inactive := deliveryZoneDocument{
		ID:        bson.NewObjectID(),
		Name:      "Porto das Dunas",
		Slug:      "porto-das-dunas",
		FeeCents:  1200,
		Active:    false,
		SortOrder: 0,
		CreatedAt: testTime(),
		UpdatedAt: testTime(),
	}
	insertDocument(t, ctx, database.Collection(CollectionDeliveryZones), second, first, inactive)

	zones, err := repository.ListActiveDeliveryZones(ctx)
	if err != nil {
		t.Fatalf("ListActiveDeliveryZones() error = %v", err)
	}
	if len(zones) != 2 {
		t.Fatalf("len(zones) = %d, want 2", len(zones))
	}
	if zones[0].ID != first.ID.Hex() || zones[1].ID != second.ID.Hex() {
		t.Errorf("zones order = %q, %q", zones[0].Name, zones[1].Name)
	}
	if zones[0].FeeCents != 800 || zones[1].FeeCents != 500 {
		t.Errorf("zones fees = %d, %d", zones[0].FeeCents, zones[1].FeeCents)
	}
}

func TestCatalogRepositoryReturnsEmptySlices(t *testing.T) {
	ctx := testContext(t)
	database := openCatalogTestDatabase(t)
	repository := newCatalogRepository(t, database)

	categories, err := repository.ListActiveCategories(ctx)
	if err != nil {
		t.Fatalf("ListActiveCategories() error = %v", err)
	}
	if categories == nil {
		t.Error("categories = nil, want empty slice")
	}

	products, err := repository.ListActiveProducts(ctx)
	if err != nil {
		t.Fatalf("ListActiveProducts() error = %v", err)
	}
	if products == nil {
		t.Error("products = nil, want empty slice")
	}

	addons, err := repository.ListActiveAddons(ctx)
	if err != nil {
		t.Fatalf("ListActiveAddons() error = %v", err)
	}
	if addons == nil {
		t.Error("addons = nil, want empty slice")
	}

	zones, err := repository.ListActiveDeliveryZones(ctx)
	if err != nil {
		t.Fatalf("ListActiveDeliveryZones() error = %v", err)
	}
	if zones == nil {
		t.Error("zones = nil, want empty slice")
	}
}

func TestCatalogRepositoryGetProduct(t *testing.T) {
	ctx := testContext(t)
	database := openCatalogTestDatabase(t)
	repository := newCatalogRepository(t, database)

	product := productDocument{
		ID:          bson.NewObjectID(),
		CategoryID:  bson.NewObjectID(),
		Name:        "Brocado's Burguer Cheddar",
		Slug:        "brocados-burguer-cheddar",
		Description: "Blend artesanal 160g.",
		PriceCents:  2300,
		Active:      true,
		CreatedAt:   testTime(),
		UpdatedAt:   testTime(),
	}
	insertDocument(t, ctx, database.Collection(CollectionProducts), product)

	got, err := repository.GetProduct(ctx, product.ID.Hex())
	if err != nil {
		t.Fatalf("GetProduct() error = %v", err)
	}
	if got.ID != product.ID.Hex() || got.PriceCents != 2300 {
		t.Errorf("GetProduct() = %+v", got)
	}

	gotBySlug, err := repository.GetProductBySlug(ctx, product.Slug)
	if err != nil {
		t.Fatalf("GetProductBySlug() error = %v", err)
	}
	if gotBySlug.ID != product.ID.Hex() {
		t.Errorf("GetProductBySlug() ID = %q, want %q", gotBySlug.ID, product.ID.Hex())
	}

	tests := []struct {
		name string
		run  func() error
	}{
		{
			name: "id inexistente",
			run: func() error {
				_, err := repository.GetProduct(ctx, bson.NewObjectID().Hex())
				return err
			},
		},
		{
			name: "id inválido",
			run: func() error {
				_, err := repository.GetProduct(ctx, "not-an-object-id")
				return err
			},
		},
		{
			name: "slug inexistente",
			run: func() error {
				_, err := repository.GetProductBySlug(ctx, "nao-existe")
				return err
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := test.run(); !errors.Is(err, domain.ErrNotFound) {
				t.Fatalf("error = %v, want domain.ErrNotFound", err)
			}
		})
	}
}

func TestCatalogRepositoryGetCatalogEntities(t *testing.T) {
	ctx := testContext(t)
	database := openCatalogTestDatabase(t)
	repository := newCatalogRepository(t, database)

	category := categoryDocument{
		ID:        bson.NewObjectID(),
		Name:      "Hambúrgueres",
		Slug:      "hamburgueres",
		SortOrder: 1,
		Active:    true,
		CreatedAt: testTime(),
		UpdatedAt: testTime(),
	}
	addon := addonDocument{
		ID:         bson.NewObjectID(),
		Name:       "Bacon Crocante",
		Slug:       "bacon-crocante",
		PriceCents: 500,
		Active:     true,
		CreatedAt:  testTime(),
		UpdatedAt:  testTime(),
	}
	zone := deliveryZoneDocument{
		ID:        bson.NewObjectID(),
		Name:      "Centro",
		Slug:      "centro",
		FeeCents:  500,
		Active:    true,
		CreatedAt: testTime(),
		UpdatedAt: testTime(),
	}
	insertDocument(t, ctx, database.Collection(CollectionCategories), category)
	insertDocument(t, ctx, database.Collection(CollectionAddons), addon)
	insertDocument(t, ctx, database.Collection(CollectionDeliveryZones), zone)

	gotCategory, err := repository.GetCategory(ctx, category.ID.Hex())
	if err != nil {
		t.Fatalf("GetCategory() error = %v", err)
	}
	if gotCategory.ID != category.ID.Hex() || gotCategory.Name != category.Name {
		t.Errorf("GetCategory() = %+v", gotCategory)
	}

	gotAddon, err := repository.GetAddon(ctx, addon.ID.Hex())
	if err != nil {
		t.Fatalf("GetAddon() error = %v", err)
	}
	if gotAddon.ID != addon.ID.Hex() || gotAddon.PriceCents != 500 {
		t.Errorf("GetAddon() = %+v", gotAddon)
	}

	gotZone, err := repository.GetDeliveryZone(ctx, zone.ID.Hex())
	if err != nil {
		t.Fatalf("GetDeliveryZone() error = %v", err)
	}
	if gotZone.ID != zone.ID.Hex() || gotZone.FeeCents != 500 {
		t.Errorf("GetDeliveryZone() = %+v", gotZone)
	}

	if _, err := repository.GetCategory(ctx, bson.NewObjectID().Hex()); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("GetCategory() error = %v, want domain.ErrNotFound", err)
	}
	if _, err := repository.GetAddon(ctx, "invalid"); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("GetAddon() error = %v, want domain.ErrNotFound", err)
	}
	if _, err := repository.GetDeliveryZone(ctx, bson.NewObjectID().Hex()); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("GetDeliveryZone() error = %v, want domain.ErrNotFound", err)
	}
}

func openCatalogTestDatabase(t *testing.T) *mongo.Database {
	t.Helper()

	uri := os.Getenv("MONGO_TEST_URI")
	if uri == "" {
		t.Skip("MONGO_TEST_URI not set; catalog repository integration tests skipped")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := Connect(ctx, uri)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}

	database := client.Database(catalogTestDatabase)
	if err := database.Drop(ctx); err != nil {
		t.Fatalf("Drop() error = %v", err)
	}

	t.Cleanup(func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := database.Drop(cleanupCtx); err != nil {
			t.Errorf("cleanup Drop() error = %v", err)
		}
		if err := client.Close(cleanupCtx); err != nil {
			t.Errorf("cleanup Close() error = %v", err)
		}
	})

	if err := EnsureCatalogIndexes(ctx, database); err != nil {
		t.Fatalf("EnsureCatalogIndexes() error = %v", err)
	}

	return database
}

func newCatalogRepository(t *testing.T, database *mongo.Database) *CatalogRepository {
	t.Helper()

	repository, err := NewCatalogRepository(database)
	if err != nil {
		t.Fatalf("NewCatalogRepository() error = %v", err)
	}
	return repository
}

func insertDocument(t *testing.T, ctx context.Context, collection *mongo.Collection, documents ...any) {
	t.Helper()

	for _, document := range documents {
		if _, err := collection.InsertOne(ctx, document); err != nil {
			t.Fatalf("InsertOne() error = %v", err)
		}
	}
}

func testContext(t *testing.T) context.Context {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)
	return ctx
}

func testTime() time.Time {
	return time.Date(2026, 11, 20, 12, 0, 0, 0, time.UTC)
}
