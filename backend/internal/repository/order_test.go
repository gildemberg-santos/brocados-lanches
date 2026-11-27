package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/brocados-lanches/api/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func orderFixture() domain.Order {
	now := time.Date(2026, 11, 22, 22, 0, 0, 0, time.UTC)
	changeForCents := int64(10000)
	return domain.Order{
		ID:          bson.NewObjectID().Hex(),
		PublicID:    "public-123",
		Status:      domain.OrderStatusPending,
		Fulfillment: domain.FulfillmentDelivery,
		Customer:    domain.OrderCustomer{Name: "Maria Silva", Phone: "+5585999999999"},
		Address:     &domain.OrderAddress{Street: "Rua das Flores", Number: "450", District: "Tapera", City: "Aquiraz", State: "CE", ZipCode: "61700-000"},
		Zone:        &domain.OrderZone{ID: bson.NewObjectID().Hex(), Name: "Centro", FeeCents: 800},
		Items: []domain.OrderItem{{
			ProductID:      bson.NewObjectID().Hex(),
			Name:           "Bacon Cheddar",
			UnitPriceCents: 2200,
			Quantity:       2,
			Options:        []domain.OrderItemOption{{OptionID: "medium", Name: "Ao ponto", Kind: domain.QuoteOptionKindOption}},
			Note:           "sem cebola",
			LineTotalCents: 4400,
		}},
		SubtotalCents:    4400,
		DeliveryFeeCents: 800,
		TotalCents:       5200,
		Payment:          domain.OrderPayment{Method: domain.PaymentMethodCash, ChangeForCents: &changeForCents},
		IdempotencyKey:   "4d1f6b1e-7c2a-4e2f-9f2a-1c3d5e7f9a0b",
		RequestHash:      "request-hash",
		AccessTokenHash:  "hash-do-token-de-historico",
		StatusHistory:    []domain.OrderStatusChange{{Status: domain.OrderStatusPending, At: now}},
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

func TestNewOrderRepositoryRequiresDatabase(t *testing.T) {
	if _, err := NewOrderRepository(nil); err == nil {
		t.Fatal("NewOrderRepository(nil) error = nil, want error")
	}
}

func TestEnsureOrderIndexesRequiresDatabase(t *testing.T) {
	if err := EnsureOrderIndexes(context.Background(), nil); err == nil {
		t.Fatal("EnsureOrderIndexes(nil) error = nil, want error")
	}
}

func TestOrderRepositoryCreateAndFind(t *testing.T) {
	ctx := testContext(t)
	database := openCatalogTestDatabase(t)
	repository, err := NewOrderRepository(database)
	if err != nil {
		t.Fatalf("NewOrderRepository() error = %v", err)
	}

	stored, err := repository.Create(ctx, orderFixture())
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if stored.Number != 1 || stored.PublicID != "public-123" || stored.Status != domain.OrderStatusPending {
		t.Errorf("stored = %+v", stored)
	}
	if len(stored.Items) != 1 || stored.Items[0].Note != "sem cebola" {
		t.Errorf("items = %+v", stored.Items)
	}
	if stored.Zone == nil || stored.Zone.FeeCents != 800 {
		t.Errorf("zone = %+v", stored.Zone)
	}

	byKey, found, err := repository.FindByIdempotencyKey(ctx, "4d1f6b1e-7c2a-4e2f-9f2a-1c3d5e7f9a0b")
	if err != nil || !found {
		t.Fatalf("FindByIdempotencyKey() error = %v, found = %v", err, found)
	}
	if byKey.ID != stored.ID {
		t.Errorf("id by key = %q, want %q", byKey.ID, stored.ID)
	}

	byPublicID, err := repository.FindByPublicID(ctx, "public-123")
	if err != nil {
		t.Fatalf("FindByPublicID() error = %v", err)
	}
	if byPublicID.Number != stored.Number {
		t.Errorf("number by public id = %d, want %d", byPublicID.Number, stored.Number)
	}
	if stored.AccessTokenHash != "hash-do-token-de-historico" {
		t.Errorf("access token hash = %q", stored.AccessTokenHash)
	}
}

func TestOrderRepositoryListAndSetAccessTokenHash(t *testing.T) {
	ctx := testContext(t)
	database := openCatalogTestDatabase(t)
	repository, err := NewOrderRepository(database)
	if err != nil {
		t.Fatal(err)
	}

	legacy := orderFixture()
	legacy.PublicID = "public-legacy"
	legacy.IdempotencyKey = "key-legacy-requisicao-1"
	legacy.AccessTokenHash = ""
	stored, err := repository.Create(ctx, legacy)
	if err != nil {
		t.Fatal(err)
	}
	if err := repository.SetAccessTokenHash(ctx, stored.ID, "hash-novo"); err != nil {
		t.Fatalf("SetAccessTokenHash() error = %v", err)
	}

	orders, total, err := repository.ListByAccessToken(ctx, "hash-novo", domain.OrderHistoryQuery{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("ListByAccessToken() error = %v", err)
	}
	if total != 1 || len(orders) != 1 || orders[0].ID != stored.ID {
		t.Errorf("history = %+v, total = %d", orders, total)
	}
	if orders[0].AccessTokenHash != "hash-novo" || !orders[0].UpdatedAt.After(stored.UpdatedAt) {
		t.Errorf("attached order = %+v", orders[0])
	}
}

func TestOrderRepositoryCreateIncrementsNumber(t *testing.T) {
	ctx := testContext(t)
	database := openCatalogTestDatabase(t)
	repository, err := NewOrderRepository(database)
	if err != nil {
		t.Fatal(err)
	}

	first := orderFixture()
	first.IdempotencyKey = "key-primeira-requisicao-1"
	first.PublicID = "public-1"
	second := orderFixture()
	second.IdempotencyKey = "key-segunda-requisicao-1"
	second.PublicID = "public-2"

	storedFirst, err := repository.Create(ctx, first)
	if err != nil {
		t.Fatalf("first Create() error = %v", err)
	}
	storedSecond, err := repository.Create(ctx, second)
	if err != nil {
		t.Fatalf("second Create() error = %v", err)
	}
	if storedSecond.Number != storedFirst.Number+1 {
		t.Errorf("numbers = %d/%d", storedFirst.Number, storedSecond.Number)
	}
}

func TestOrderRepositoryCreateRejectsDuplicateIdempotencyKey(t *testing.T) {
	ctx := testContext(t)
	database := openCatalogTestDatabase(t)
	repository, err := NewOrderRepository(database)
	if err != nil {
		t.Fatal(err)
	}

	first := orderFixture()
	if _, err := repository.Create(ctx, first); err != nil {
		t.Fatalf("first Create() error = %v", err)
	}

	duplicate := orderFixture()
	duplicate.ID = bson.NewObjectID().Hex()
	duplicate.PublicID = "public-different"
	if _, err := repository.Create(ctx, duplicate); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("duplicate Create() error = %v, want ErrConflict", err)
	}
}

func TestOrderRepositoryFindByPublicIDNotFound(t *testing.T) {
	ctx := testContext(t)
	database := openCatalogTestDatabase(t)
	repository, err := NewOrderRepository(database)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := repository.FindByPublicID(ctx, "inexistente"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("FindByPublicID() error = %v, want ErrNotFound", err)
	}
}

func TestEnsureOrderIndexesIsIdempotent(t *testing.T) {
	ctx := testContext(t)
	database := openCatalogTestDatabase(t)

	if err := EnsureOrderIndexes(ctx, database); err != nil {
		t.Fatalf("second EnsureOrderIndexes() error = %v", err)
	}
	indexes := listIndexes(t, ctx, database.Collection(CollectionOrders))
	for _, name := range []string{
		"orders_number_unique",
		"orders_public_id_unique",
		"orders_idempotency_unique",
		"orders_status_created",
		"orders_created",
		"orders_phone_created",
	} {
		if _, ok := indexes[name]; !ok {
			t.Errorf("missing index %q", name)
		}
	}
}
