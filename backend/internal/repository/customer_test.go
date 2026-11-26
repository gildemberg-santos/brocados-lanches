package repository

import (
	"context"
	"testing"
	"time"

	"github.com/brocados-lanches/api/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func customerProfileOrderFixture() domain.Order {
	return domain.Order{
		PublicID:    "publico-123",
		Fulfillment: domain.FulfillmentDelivery,
		Customer:    domain.OrderCustomer{Name: "Maria Silva", Phone: "+5585999999999"},
		Address: &domain.OrderAddress{
			Street: "Rua das Flores", Number: "450", District: "Tapera",
			City: "Aquiraz", State: "CE", ZipCode: "61700000",
		},
		Zone:      &domain.OrderZone{ID: bson.NewObjectID().Hex(), Name: "Centro", FeeCents: 800},
		Payment:   domain.OrderPayment{Method: domain.PaymentMethodPix},
		CreatedAt: time.Date(2026, 11, 26, 12, 0, 0, 0, time.UTC),
	}
}

func TestNewCustomerProfileRepositoryRequiresDatabase(t *testing.T) {
	if _, err := NewCustomerProfileRepository(nil); err == nil {
		t.Fatal("NewCustomerProfileRepository(nil) error = nil, want error")
	}
}

func TestEnsureCustomerProfileIndexesRequiresDatabase(t *testing.T) {
	if err := EnsureCustomerProfileIndexes(context.Background(), nil); err == nil {
		t.Fatal("EnsureCustomerProfileIndexes(nil) error = nil, want error")
	}
}

func TestCustomerProfileRepositoryUpsertAndFind(t *testing.T) {
	ctx := testContext(t)
	database := openCatalogTestDatabase(t)
	if err := EnsureCustomerProfileIndexes(ctx, database); err != nil {
		t.Fatalf("EnsureCustomerProfileIndexes() error = %v", err)
	}
	repository, err := NewCustomerProfileRepository(database)
	if err != nil {
		t.Fatalf("NewCustomerProfileRepository() error = %v", err)
	}

	profile, err := domain.NewCustomerProfileFromOrder(customerProfileOrderFixture(), time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	if err := repository.Upsert(ctx, profile); err != nil {
		t.Fatalf("first Upsert() error = %v", err)
	}
	stored, found, err := repository.FindByPhoneKey(ctx, "5585999999999")
	if err != nil || !found {
		t.Fatalf("FindByPhoneKey() error = %v, found = %v", err, found)
	}
	if stored.OrderCount != 1 || stored.Name != "Maria Silva" || stored.Address == nil || stored.Address.Street != "Rua das Flores" {
		t.Fatalf("stored = %+v", stored)
	}
	if stored.Zone == nil || stored.Zone.Name != "Centro" || stored.PaymentMethod != domain.PaymentMethodPix {
		t.Fatalf("preferences = %+v", stored)
	}

	secondOrder := customerProfileOrderFixture()
	secondOrder.PublicID = "publico-456"
	secondOrder.Zone.Name = "Jóquei"
	secondOrder.CreatedAt = stored.LastOrderAt.Add(time.Hour)
	second, err := domain.NewCustomerProfileFromOrder(secondOrder, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	if err := repository.Upsert(ctx, second); err != nil {
		t.Fatalf("second Upsert() error = %v", err)
	}
	updated, found, err := repository.FindByPhoneKey(ctx, "5585999999999")
	if err != nil || !found {
		t.Fatalf("FindByPhoneKey() error = %v, found = %v", err, found)
	}
	if updated.OrderCount != 2 || updated.Zone == nil || updated.Zone.Name != "Jóquei" {
		t.Fatalf("updated = %+v", updated)
	}
}

func TestCustomerProfileRepositoryOptOutPreventsRecreation(t *testing.T) {
	ctx := testContext(t)
	database := openCatalogTestDatabase(t)
	repository, err := NewCustomerProfileRepository(database)
	if err != nil {
		t.Fatal(err)
	}
	profile, err := domain.NewCustomerProfileFromOrder(customerProfileOrderFixture(), time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	if err := repository.Upsert(ctx, profile); err != nil {
		t.Fatalf("Upsert() error = %v", err)
	}

	if err := repository.OptOut(ctx, profile.PhoneKey, time.Now().UTC()); err != nil {
		t.Fatalf("OptOut() error = %v", err)
	}
	optedOut, found, err := repository.FindByPhoneKey(ctx, profile.PhoneKey)
	if err != nil || !found {
		t.Fatalf("FindByPhoneKey() error = %v, found = %v", err, found)
	}
	if !optedOut.OptedOut || optedOut.Name != "" || optedOut.Address != nil || optedOut.OrderCount != 0 {
		t.Fatalf("optedOut profile = %+v", optedOut)
	}

	if err := repository.Upsert(ctx, profile); err != nil {
		t.Fatalf("Upsert() after opt-out error = %v", err)
	}
	stillOptedOut, found, err := repository.FindByPhoneKey(ctx, profile.PhoneKey)
	if err != nil || !found {
		t.Fatalf("FindByPhoneKey() error = %v, found = %v", err, found)
	}
	if !stillOptedOut.OptedOut || stillOptedOut.Name != "" || stillOptedOut.OrderCount != 0 {
		t.Fatalf("profile recreated after opt-out: %+v", stillOptedOut)
	}
}

func TestCustomerProfileRepositoryFindByPhoneKeyNotFound(t *testing.T) {
	ctx := testContext(t)
	database := openCatalogTestDatabase(t)
	repository, err := NewCustomerProfileRepository(database)
	if err != nil {
		t.Fatal(err)
	}
	if _, found, err := repository.FindByPhoneKey(ctx, "5585000000000"); err != nil || found {
		t.Fatalf("FindByPhoneKey() error = %v, found = %v", err, found)
	}
}

func TestEnsureCustomerProfileIndexesIsIdempotent(t *testing.T) {
	ctx := testContext(t)
	database := openCatalogTestDatabase(t)

	if err := EnsureCustomerProfileIndexes(ctx, database); err != nil {
		t.Fatalf("EnsureCustomerProfileIndexes() error = %v", err)
	}
	indexes := listIndexes(t, ctx, database.Collection(CollectionCustomerProfiles))
	for _, name := range []string{
		"customer_profiles_phone_key_unique",
		"customer_profiles_expires_at_ttl",
		"customer_profiles_last_order",
	} {
		if _, ok := indexes[name]; !ok {
			t.Errorf("missing index %q", name)
		}
	}
}
