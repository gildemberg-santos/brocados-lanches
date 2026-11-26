package seed

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func TestApplyIntegration(t *testing.T) {
	uri := os.Getenv("MONGO_TEST_URI")
	if uri == "" {
		t.Skip("MONGO_TEST_URI not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	client, err := mongo.Connect(options.Client().ApplyURI(uri).SetServerSelectionTimeout(5 * time.Second))
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			t.Errorf("Disconnect() error = %v", err)
		}
	}()

	database := client.Database("brocados_seed_test")
	if err := database.Drop(ctx); err != nil {
		t.Fatalf("Drop() error = %v", err)
	}
	defer func() {
		if err := database.Drop(context.Background()); err != nil {
			t.Errorf("cleanup Drop() error = %v", err)
		}
	}()

	data, err := Load(filepath.Join("..", "..", "..", "design-system", "seed.json"))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	plan, err := Build(data)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	result, err := Apply(ctx, database, plan, true)
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	if result.StoreSettings != 1 || result.Categories != 5 || result.Products != 37 || result.DeliveryZones != 30 {
		t.Fatalf("Apply() result = %+v", result)
	}

	assertDocumentCounts(t, ctx, database)

	var product struct {
		PriceCents int64  `bson:"priceCents"`
		ImageURL   string `bson:"imageUrl"`
		ImageAlt   string `bson:"imageAlt"`
		Seed       struct {
			Fictitious bool `bson:"fictitious"`
		} `bson:"seed"`
	}
	if err := database.Collection(CollectionProducts).
		FindOne(ctx, bson.D{{Key: "slug", Value: "combo-3-hamburguer"}}).
		Decode(&product); err != nil {
		t.Fatalf("FindOne() error = %v", err)
	}
	if product.PriceCents != 4999 {
		t.Fatalf("product priceCents = %d, want 4999", product.PriceCents)
	}
	if product.ImageURL != "/products/combo-3-hamburguer.png" || product.ImageAlt != "Combo 3 hambúrguer" {
		t.Fatalf("product image = %q / %q", product.ImageURL, product.ImageAlt)
	}
	if !product.Seed.Fictitious {
		t.Fatal("product is not marked as fictitious")
	}

	// A segunda execução, sem reset, precisa manter a mesma quantidade.
	second, err := Apply(ctx, database, plan, false)
	if err != nil {
		t.Fatalf("second Apply() error = %v", err)
	}
	if second.StoreSettings != 1 || second.Categories != 5 || second.Products != 37 || second.DeliveryZones != 30 {
		t.Fatalf("second Apply() result = %+v", second)
	}
	assertDocumentCounts(t, ctx, database)

	if _, err := Apply(ctx, client.Database("brocados"), plan, false); err == nil {
		t.Fatal("Apply() against production database returned nil error")
	}
}

func assertDocumentCounts(t *testing.T, ctx context.Context, database *mongo.Database) {
	t.Helper()

	expected := map[string]int64{
		CollectionStoreSettings: 1,
		CollectionCategories:    5,
		CollectionProducts:      37,
		CollectionDeliveryZones: 30,
	}
	for collection, want := range expected {
		got, err := database.Collection(collection).CountDocuments(ctx, bson.D{})
		if err != nil {
			t.Fatalf("CountDocuments(%s) error = %v", collection, err)
		}
		if got != want {
			t.Fatalf("CountDocuments(%s) = %d, want %d", collection, got, want)
		}
	}
}
