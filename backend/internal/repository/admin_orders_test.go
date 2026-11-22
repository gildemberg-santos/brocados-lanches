package repository

import (
	"errors"
	"testing"
	"time"

	"github.com/brocados-lanches/api/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func adminOrderDocumentFixture(status string, number int64, fulfillment domain.Fulfillment, phone string, createdAt time.Time) orderDocument {
	return orderDocument{
		ID: bson.NewObjectID(), Number: number,
		PublicID:       "publico-" + bson.NewObjectID().Hex(),
		IdempotencyKey: "chave-" + bson.NewObjectID().Hex(),
		RequestHash:    "hash-" + bson.NewObjectID().Hex(),
		Status:         status, Fulfillment: string(fulfillment),
		Customer: customerDocument{Name: "Maria Silva", Phone: phone},
		Items: []orderItemDocument{{
			ProductID: bson.NewObjectID(), Name: "Burguer",
			UnitPriceCents: 2300, Quantity: 1, LineTotalCents: 2300,
			Options: []itemOptionDocument{{OptionID: "medium", Name: "Ao ponto", Kind: "option"}},
		}},
		SubtotalCents: 2300, DeliveryFeeCents: 500, TotalCents: 2800,
		Payment:       paymentDocument{Method: domain.PaymentMethodCash},
		StatusHistory: []statusChangeDocument{{Status: status, At: createdAt}},
		CreatedAt:     createdAt, UpdatedAt: createdAt,
	}
}

func TestAdminOrderRepositoryListFiltersAndPagination(t *testing.T) {
	ctx := testContext(t)
	database := openCatalogTestDatabase(t)
	repository, err := NewOrderRepository(database)
	if err != nil {
		t.Fatalf("NewOrderRepository() error = %v", err)
	}

	newest := time.Date(2026, 11, 22, 18, 0, 0, 0, time.UTC)
	insertDocument(t, ctx, database.Collection(CollectionOrders),
		adminOrderDocumentFixture(domain.OrderStatusPending, 41, domain.FulfillmentDelivery, "+5585992358708", newest),
		adminOrderDocumentFixture(domain.OrderStatusConfirmed, 42, domain.FulfillmentPickup, "+5585998854321", newest.Add(-time.Hour)),
		adminOrderDocumentFixture(domain.OrderStatusCancelled, 43, domain.FulfillmentDelivery, "+5585991112223", newest.Add(-2*time.Hour)),
	)

	all, total, err := repository.AdminListOrders(ctx, domain.AdminOrderListQuery{
		Page: 1, PageSize: 2, Sort: "createdAt", Order: "desc",
	})
	if err != nil {
		t.Fatalf("AdminListOrders() error = %v", err)
	}
	if total != 3 || len(all) != 2 || all[0].Number != 41 {
		t.Fatalf("list = %+v, total = %d", all, total)
	}

	from := newest.Add(-90 * time.Minute)
	to := newest.Add(time.Minute)
	filtered, total, err := repository.AdminListOrders(ctx, domain.AdminOrderListQuery{
		Page: 1, PageSize: 10, Status: domain.OrderStatusConfirmed,
		Fulfillment: domain.FulfillmentPickup, From: &from, To: &to,
	})
	if err != nil {
		t.Fatalf("filtered list error = %v", err)
	}
	if total != 1 || len(filtered) != 1 || filtered[0].Number != 42 {
		t.Fatalf("filtered = %+v, total = %d", filtered, total)
	}

	searched, total, err := repository.AdminListOrders(ctx, domain.AdminOrderListQuery{
		Page: 1, PageSize: 10, Search: "(85) 99235-8708",
	})
	if err != nil {
		t.Fatalf("search error = %v", err)
	}
	if total != 1 || len(searched) != 1 || searched[0].Number != 41 {
		t.Fatalf("search = %+v, total = %d", searched, total)
	}

	if _, _, err := repository.AdminListOrders(ctx, domain.AdminOrderListQuery{
		Page: 1, PageSize: 10, Sort: "customer.phone",
	}); err == nil {
		t.Error("unsupported sort error = nil, want validation error")
	}
}

func TestAdminOrderRepositoryFindByID(t *testing.T) {
	ctx := testContext(t)
	database := openCatalogTestDatabase(t)
	repository, err := NewOrderRepository(database)
	if err != nil {
		t.Fatalf("NewOrderRepository() error = %v", err)
	}
	fixture := adminOrderDocumentFixture(domain.OrderStatusPending, 41, domain.FulfillmentDelivery, "+5585992358708", testTime())
	insertDocument(t, ctx, database.Collection(CollectionOrders), fixture)

	found, err := repository.FindByID(ctx, fixture.ID.Hex())
	if err != nil || found.Number != 41 || len(found.StatusHistory) != 1 {
		t.Fatalf("FindByID() = %+v, %v", found, err)
	}
	if _, err := repository.FindByID(ctx, "invalido"); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("invalid id error = %v, want ErrNotFound", err)
	}
	if _, err := repository.FindByID(ctx, bson.NewObjectID().Hex()); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("unknown id error = %v, want ErrNotFound", err)
	}
}

func TestAdminOrderRepositoryUpdateStatusWithAuditAndCompareAndSet(t *testing.T) {
	ctx := testContext(t)
	database := openCatalogTestDatabase(t)
	if err := EnsureAuditIndexes(ctx, database); err != nil {
		t.Fatalf("EnsureAuditIndexes() error = %v", err)
	}
	repository, err := NewOrderRepository(database)
	if err != nil {
		t.Fatalf("NewOrderRepository() error = %v", err)
	}
	fixture := adminOrderDocumentFixture(domain.OrderStatusPending, 41, domain.FulfillmentDelivery, "+5585992358708", testTime())
	insertDocument(t, ctx, database.Collection(CollectionOrders), fixture)

	current, err := repository.FindByID(ctx, fixture.ID.Hex())
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	now := time.Date(2026, 11, 22, 18, 30, 0, 0, time.UTC)
	updated, err := domain.ApplyOrderStatusChange(current, domain.OrderStatusChangeRequest{
		Status: domain.OrderStatusConfirmed,
	}, adminTestActorID, now)
	if err != nil {
		t.Fatalf("ApplyOrderStatusChange() error = %v", err)
	}
	audit := domain.AuditLog{
		ActorID: adminTestActorID, Action: domain.AuditActionUpdate,
		Entity: domain.AuditEntityOrder, EntityID: fixture.ID.Hex(),
		Before:    map[string]any{"status": domain.OrderStatusPending},
		After:     map[string]any{"status": domain.OrderStatusConfirmed},
		CreatedAt: now,
	}
	if err := repository.AdminUpdateOrderStatus(ctx, fixture.ID.Hex(), current.Status, updated, audit); err != nil {
		t.Fatalf("AdminUpdateOrderStatus() error = %v", err)
	}

	stored, err := repository.FindByID(ctx, fixture.ID.Hex())
	if err != nil {
		t.Fatalf("FindByID() after update error = %v", err)
	}
	if stored.Status != domain.OrderStatusConfirmed || len(stored.StatusHistory) != 2 ||
		stored.ConfirmedAt == nil || !stored.ConfirmedAt.Equal(now) {
		t.Errorf("stored = %+v", stored)
	}
	count, err := database.Collection(CollectionAuditLogs).CountDocuments(ctx, bson.D{{Key: "entityId", Value: fixture.ID.Hex()}})
	if err != nil || count != 1 {
		t.Fatalf("audit count = %d, err = %v; want 1", count, err)
	}

	if err := repository.AdminUpdateOrderStatus(ctx, fixture.ID.Hex(), domain.OrderStatusPending, updated, audit); !errors.Is(err, domain.ErrConflict) {
		t.Errorf("stale transition error = %v, want ErrConflict", err)
	}
	count, err = database.Collection(CollectionAuditLogs).CountDocuments(ctx, bson.D{{Key: "entityId", Value: fixture.ID.Hex()}})
	if err != nil || count != 1 {
		t.Errorf("audit count after stale transition = %d, err = %v; want 1", count, err)
	}
	if err := repository.AdminUpdateOrderStatus(ctx, bson.NewObjectID().Hex(), domain.OrderStatusPending, updated, audit); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("unknown id error = %v, want ErrNotFound", err)
	}
}
