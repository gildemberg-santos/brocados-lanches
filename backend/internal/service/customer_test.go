package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/brocados-lanches/api/internal/domain"
)

type fakeCustomerProfileStore struct {
	upserts   []domain.CustomerProfile
	optOuts   []string
	findKey   string
	found     domain.CustomerProfile
	foundFlag bool
	upsertErr error
	optOutErr error
	findErr   error
}

func (f *fakeCustomerProfileStore) Upsert(_ context.Context, profile domain.CustomerProfile) error {
	f.upserts = append(f.upserts, profile)
	return f.upsertErr
}

func (f *fakeCustomerProfileStore) OptOut(_ context.Context, phoneKey string, _ time.Time) error {
	f.optOuts = append(f.optOuts, phoneKey)
	return f.optOutErr
}

func (f *fakeCustomerProfileStore) FindByPhoneKey(_ context.Context, phoneKey string) (domain.CustomerProfile, bool, error) {
	f.findKey = phoneKey
	return f.found, f.foundFlag, f.findErr
}

func customerProfileStoredOrder() domain.Order {
	return domain.Order{
		ID: "order-1", PublicID: "public-123", Number: 1042,
		Status:      domain.OrderStatusPending,
		Customer:    domain.OrderCustomer{Name: "Maria Silva", Phone: "+5585999999999"},
		Fulfillment: domain.FulfillmentDelivery,
		Zone:        &domain.OrderZone{ID: "zone-1", Name: "Centro", FeeCents: 800},
		Payment:     domain.OrderPayment{Method: domain.PaymentMethodPix},
		Address:     createOrderRequestFixture().Address,
		CreatedAt:   time.Now().UTC(),
	}
}

func TestOrderServiceCreateSavesCustomerProfile(t *testing.T) {
	reader, _ := orderServiceFixture()
	reader.created = customerProfileStoredOrder()
	store := &fakeCustomerProfileStore{}
	service, err := NewOrderService(
		reader,
		time.UTC,
		WithOrderAccessSecret("test-order-access-secret"),
		WithCustomerProfiles(store, nil),
	)
	if err != nil {
		t.Fatalf("NewOrderService() error = %v", err)
	}
	service.publicID = func() (string, error) { return "public-123", nil }

	if _, _, err := service.Create(context.Background(), createOrderRequestFixture(), testIdempotencyKey, ""); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if len(store.upserts) != 1 {
		t.Fatalf("upserts = %d, want 1", len(store.upserts))
	}
	profile := store.upserts[0]
	if profile.PhoneKey != "5585999999999" || profile.Name != "Maria Silva" || profile.Zone == nil || profile.Zone.Name != "Centro" {
		t.Errorf("profile = %+v", profile)
	}
	if len(store.optOuts) != 0 {
		t.Error("OptOut() called without explicit opt-out")
	}
}

func TestOrderServiceCreateHonorsCustomerProfileOptOut(t *testing.T) {
	reader, service := orderServiceFixture()
	reader.created = customerProfileStoredOrder()
	service.publicID = func() (string, error) { return "public-123", nil }
	store := &fakeCustomerProfileStore{}
	service.profiles = store
	service.logger = discardLogger()
	request := createOrderRequestFixture()
	optedOut := false
	request.SaveCustomerProfile = &optedOut

	if _, _, err := service.Create(context.Background(), request, testIdempotencyKey, ""); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if len(store.upserts) != 0 || len(store.optOuts) != 1 || store.optOuts[0] != "5585999999999" {
		t.Fatalf("upserts = %d, optOuts = %v", len(store.upserts), store.optOuts)
	}
}

func TestOrderServiceCreateDoesNotFailWhenProfileWriteFails(t *testing.T) {
	reader, service := orderServiceFixture()
	reader.created = customerProfileStoredOrder()
	service.publicID = func() (string, error) { return "public-123", nil }
	store := &fakeCustomerProfileStore{upsertErr: errors.New("mongo indisponível")}
	service.profiles = store
	service.logger = discardLogger()

	if _, _, err := service.Create(context.Background(), createOrderRequestFixture(), testIdempotencyKey, ""); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if len(store.upserts) != 1 {
		t.Fatalf("upserts = %d, want 1", len(store.upserts))
	}
}

func TestOrderServiceCreateReplayDoesNotRewriteCustomerProfile(t *testing.T) {
	reader, service := orderServiceFixture()
	reader.existingFound = true
	hash, err := domain.CanonicalRequestHash(createOrderRequestFixture())
	if err != nil {
		t.Fatal(err)
	}
	reader.existing = domain.Order{ID: "order-1", PublicID: "public-123", RequestHash: hash}
	store := &fakeCustomerProfileStore{}
	service.profiles = store
	service.logger = discardLogger()

	if _, _, err := service.Create(context.Background(), createOrderRequestFixture(), testIdempotencyKey, ""); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if len(store.upserts) != 0 || len(store.optOuts) != 0 {
		t.Fatalf("replay rewrote profile: %+v", store)
	}
}

func TestAdminOrderServiceGetCustomerProfile(t *testing.T) {
	repo := newFakeAdminOrderRepository()
	store := &fakeCustomerProfileStore{
		foundFlag: true,
		found:     domain.CustomerProfile{PhoneKey: "5585992358708", Name: "Maria Silva", OrderCount: 3},
	}
	service, err := NewAdminOrderService(repo, WithAdminCustomerProfiles(store))
	if err != nil {
		t.Fatalf("NewAdminOrderService() error = %v", err)
	}

	profile, err := service.GetCustomerProfile(context.Background(), "(85) 99235-8708")
	if err != nil {
		t.Fatalf("GetCustomerProfile() error = %v", err)
	}
	if profile == nil || profile.OrderCount != 3 || store.findKey != "5585992358708" {
		t.Fatalf("profile = %+v, key = %q", profile, store.findKey)
	}

	withoutStore, err := NewAdminOrderService(repo)
	if err != nil {
		t.Fatal(err)
	}
	if profile, err := withoutStore.GetCustomerProfile(context.Background(), "+5585992358708"); err != nil || profile != nil {
		t.Fatalf("profile = %+v, error = %v", profile, err)
	}
}
