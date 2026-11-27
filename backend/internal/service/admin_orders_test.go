package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/brocados-lanches/api/internal/domain"
)

const (
	adminOrderTestID    = "665f1c2a9b1e4a0012a3b4f1"
	adminOrderTestActor = "665f1c2a9b1e4a0012a3b4f9"
)

type fakeAdminOrderRepository struct {
	orders             map[string]domain.Order
	listed             []domain.Order
	total              int64
	lastQuery          domain.AdminOrderListQuery
	lastExpectedStatus string
	lastUpdated        domain.Order
	lastAudit          domain.AuditLog
	listErr            error
	getErr             error
	writeErr           error
	deleteErr          error
	deletedID          string
	deleted            bool
	writes             int
}

func newFakeAdminOrderRepository() *fakeAdminOrderRepository {
	return &fakeAdminOrderRepository{orders: make(map[string]domain.Order)}
}

func (f *fakeAdminOrderRepository) AdminListOrders(_ context.Context, query domain.AdminOrderListQuery) ([]domain.Order, int64, error) {
	f.lastQuery = query
	return f.listed, f.total, f.listErr
}

func (f *fakeAdminOrderRepository) FindByID(_ context.Context, id string) (domain.Order, error) {
	if f.getErr != nil {
		return domain.Order{}, f.getErr
	}
	order, ok := f.orders[id]
	if !ok {
		return domain.Order{}, domain.ErrNotFound
	}
	return order, nil
}

func (f *fakeAdminOrderRepository) AdminUpdateOrderStatus(_ context.Context, id, expectedStatus string, order domain.Order, audit domain.AuditLog) error {
	f.writes++
	f.lastExpectedStatus = expectedStatus
	f.lastUpdated = order
	f.lastAudit = audit
	if f.writeErr != nil {
		return f.writeErr
	}
	f.orders[id] = order
	return nil
}

func (f *fakeAdminOrderRepository) AdminDeleteOrder(_ context.Context, id string, audit domain.AuditLog) error {
	f.writes++
	f.deletedID = id
	f.lastAudit = audit
	if f.deleteErr != nil {
		return f.deleteErr
	}
	delete(f.orders, id)
	f.deleted = true
	return nil
}

func newAdminOrderTestService(repo *fakeAdminOrderRepository) *AdminOrderService {
	service, err := NewAdminOrderService(repo)
	if err != nil {
		panic(err)
	}
	service.now = func() time.Time { return time.Date(2026, 11, 22, 18, 30, 0, 0, time.UTC) }
	return service
}

func adminOrderFixture() domain.Order {
	return domain.Order{
		ID: adminOrderTestID, Number: 42, PublicID: "opaco",
		Status: domain.OrderStatusPending, Fulfillment: domain.FulfillmentDelivery,
		Customer:      domain.OrderCustomer{Name: "Maria Silva", Phone: "+5585992358708"},
		Items:         []domain.OrderItem{{Name: "Burguer"}, {Name: "Fritas"}},
		SubtotalCents: 2300, DeliveryFeeCents: 500, TotalCents: 2800,
		CreatedAt: time.Date(2026, 11, 22, 12, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 11, 22, 12, 0, 0, 0, time.UTC),
		StatusHistory: []domain.OrderStatusChange{{
			Status: domain.OrderStatusPending, At: time.Date(2026, 11, 22, 12, 0, 0, 0, time.UTC),
		}},
	}
}

func TestAdminOrderServiceListNormalizesAndMasks(t *testing.T) {
	repo := newFakeAdminOrderRepository()
	repo.listed = []domain.Order{adminOrderFixture(), adminOrderFixture()}
	repo.total = 21
	service := newAdminOrderTestService(repo)

	page, err := service.List(context.Background(), domain.AdminOrderListQuery{})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if page.Page != 1 || page.PageSize != 20 || page.Total != 21 || page.TotalPages != 2 {
		t.Fatalf("page = %+v", page)
	}
	if repo.lastQuery.Sort != "createdAt" || repo.lastQuery.Order != "desc" {
		t.Errorf("query = %+v", repo.lastQuery)
	}
	if len(page.Items) != 2 || page.Items[0].CustomerName != "M*** S***" || page.Items[0].Phone != "****8708" {
		t.Errorf("items = %+v", page.Items)
	}
	if page.Items[0].ItemCount != 2 || page.Items[0].TotalCents != 2800 {
		t.Errorf("item = %+v", page.Items[0])
	}

	from := time.Date(2026, 11, 20, 0, 0, 0, 0, time.UTC)
	to := from.Add(time.Hour)
	_, err = service.List(context.Background(), domain.AdminOrderListQuery{
		Page: 2, PageSize: 10, Sort: "number", Order: "asc",
		Status: domain.OrderStatusPending, Fulfillment: domain.FulfillmentPickup,
		From: &from, To: &to, Search: "8708",
	})
	if err != nil {
		t.Fatalf("List() filtered error = %v", err)
	}
	if repo.lastQuery.Page != 2 || repo.lastQuery.Status != domain.OrderStatusPending ||
		repo.lastQuery.Fulfillment != domain.FulfillmentPickup || repo.lastQuery.Search != "8708" {
		t.Errorf("filtered query = %+v", repo.lastQuery)
	}
}

func TestAdminOrderServiceListRejectsInvalidFilters(t *testing.T) {
	repo := newFakeAdminOrderRepository()
	service := newAdminOrderTestService(repo)
	from := time.Date(2026, 11, 22, 12, 0, 0, 0, time.UTC)
	to := from.Add(-time.Hour)

	tests := []struct {
		name  string
		query domain.AdminOrderListQuery
		field string
	}{
		{"página", domain.AdminOrderListQuery{Page: -1}, "page"},
		{"tamanho", domain.AdminOrderListQuery{PageSize: 101}, "pageSize"},
		{"ordenação", domain.AdminOrderListQuery{Sort: "customer.phone"}, "sort"},
		{"direção", domain.AdminOrderListQuery{Order: "sideways"}, "order"},
		{"status", domain.AdminOrderListQuery{Status: "cozendo"}, "status"},
		{"entrega", domain.AdminOrderListQuery{Fulfillment: "teleporte"}, "fulfillment"},
		{"período", domain.AdminOrderListQuery{From: &from, To: &to}, "from"},
		{"busca longa", domain.AdminOrderListQuery{Search: string(make([]byte, 121))}, "q"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := service.List(context.Background(), test.query)
			var validationError *domain.ValidationError
			if !errors.As(err, &validationError) || validationError.Field != test.field {
				t.Fatalf("error = %v, want field %q", err, test.field)
			}
		})
	}
	if repo.writes != 0 {
		t.Error("invalid filters must not write")
	}
}

func TestAdminOrderServiceGet(t *testing.T) {
	repo := newFakeAdminOrderRepository()
	repo.orders[adminOrderTestID] = adminOrderFixture()
	service := newAdminOrderTestService(repo)

	if _, err := service.Get(context.Background(), "curto"); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("invalid id error = %v, want ErrNotFound", err)
	}
	if _, err := service.Get(context.Background(), "665f1c2a9b1e4a0012a3b4f2"); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("unknown id error = %v, want ErrNotFound", err)
	}
	order, err := service.Get(context.Background(), adminOrderTestID)
	if err != nil || order.Customer.Name != "Maria Silva" {
		t.Fatalf("Get() = %+v, %v", order, err)
	}
}

func TestAdminOrderServiceUpdateStatus(t *testing.T) {
	repo := newFakeAdminOrderRepository()
	repo.orders[adminOrderTestID] = adminOrderFixture()
	service := newAdminOrderTestService(repo)

	updated, err := service.UpdateStatus(context.Background(), adminOrderTestActor, adminOrderTestID,
		domain.OrderStatusChangeRequest{Status: domain.OrderStatusConfirmed})
	if err != nil {
		t.Fatalf("UpdateStatus() error = %v", err)
	}
	if updated.Status != domain.OrderStatusConfirmed || updated.ConfirmedAt == nil ||
		len(updated.StatusHistory) != 2 || updated.StatusHistory[1].ActorID != adminOrderTestActor {
		t.Errorf("updated = %+v", updated)
	}
	if repo.lastExpectedStatus != domain.OrderStatusPending || repo.writes != 1 {
		t.Errorf("expected status = %q, writes = %d", repo.lastExpectedStatus, repo.writes)
	}
	if repo.lastAudit.Action != domain.AuditActionUpdate || repo.lastAudit.Entity != domain.AuditEntityOrder ||
		repo.lastAudit.EntityID != adminOrderTestID || repo.lastAudit.ActorID != adminOrderTestActor {
		t.Errorf("audit = %+v", repo.lastAudit)
	}
	if repo.lastAudit.Before["status"] != domain.OrderStatusPending || repo.lastAudit.After["status"] != domain.OrderStatusConfirmed {
		t.Errorf("audit snapshots = %+v / %+v", repo.lastAudit.Before, repo.lastAudit.After)
	}
	if _, ok := repo.lastAudit.After["customer"]; ok {
		t.Error("audit must not contain customer data")
	}
}

func TestAdminOrderServiceUpdateStatusCancelAndErrors(t *testing.T) {
	repo := newFakeAdminOrderRepository()
	repo.orders[adminOrderTestID] = adminOrderFixture()
	service := newAdminOrderTestService(repo)

	cancelled, err := service.UpdateStatus(context.Background(), adminOrderTestActor, adminOrderTestID,
		domain.OrderStatusChangeRequest{Status: domain.OrderStatusCancelled, Reason: "  desistiu  "})
	if err != nil {
		t.Fatalf("cancel error = %v", err)
	}
	if cancelled.CancelledAt == nil || cancelled.CancelReason != "desistiu" {
		t.Errorf("cancelled = %+v", cancelled)
	}

	repo.orders[adminOrderTestID] = adminOrderFixture()
	if _, err := service.UpdateStatus(context.Background(), adminOrderTestActor, adminOrderTestID,
		domain.OrderStatusChangeRequest{Status: domain.OrderStatusPreparing}); !errors.Is(err, domain.ErrConflict) {
		t.Errorf("invalid transition error = %v, want ErrConflict", err)
	}
	if _, err := service.UpdateStatus(context.Background(), adminOrderTestActor, adminOrderTestID,
		domain.OrderStatusChangeRequest{Status: domain.OrderStatusCancelled}); err == nil {
		t.Error("cancel without reason error = nil, want validation error")
	}
	if _, err := service.UpdateStatus(context.Background(), adminOrderTestActor, "665f1c2a9b1e4a0012a3b4f2",
		domain.OrderStatusChangeRequest{Status: domain.OrderStatusConfirmed}); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("unknown id error = %v, want ErrNotFound", err)
	}

	writesBefore := repo.writes
	repo.writeErr = domain.ErrConflict
	if _, err := service.UpdateStatus(context.Background(), adminOrderTestActor, adminOrderTestID,
		domain.OrderStatusChangeRequest{Status: domain.OrderStatusConfirmed}); !errors.Is(err, domain.ErrConflict) {
		t.Errorf("stale transition error = %v, want ErrConflict", err)
	}
	if repo.writes != writesBefore+1 {
		t.Error("stale transition must reach repository once")
	}
}

func TestAdminOrderServiceDeleteAnyStatusWithAuditableSnapshot(t *testing.T) {
	for _, status := range []string{
		domain.OrderStatusPending, domain.OrderStatusConfirmed, domain.OrderStatusPreparing,
		domain.OrderStatusReady, domain.OrderStatusOutForDelivery, domain.OrderStatusPickedUp,
		domain.OrderStatusDelivered, domain.OrderStatusCancelled,
	} {
		t.Run(status, func(t *testing.T) {
			repo := newFakeAdminOrderRepository()
			order := adminOrderFixture()
			order.Status = status
			repo.orders[adminOrderTestID] = order
			service := newAdminOrderTestService(repo)

			deleted, err := service.Delete(context.Background(), adminOrderTestActor, adminOrderTestID)
			if err != nil {
				t.Fatalf("Delete() error = %v", err)
			}
			if deleted.Number != 42 || repo.deletedID != adminOrderTestID || !repo.deleted {
				t.Fatalf("deleted = %+v, repository = %+v", deleted, repo)
			}
			if repo.lastAudit.Action != domain.AuditActionDelete || repo.lastAudit.Entity != domain.AuditEntityOrder ||
				repo.lastAudit.EntityID != adminOrderTestID || repo.lastAudit.ActorID != adminOrderTestActor {
				t.Fatalf("audit = %+v", repo.lastAudit)
			}
			if repo.lastAudit.Before["number"] != int64(42) || repo.lastAudit.Before["status"] != status {
				t.Fatalf("audit before = %+v", repo.lastAudit.Before)
			}
			if repo.lastAudit.After != nil {
				t.Fatalf("delete audit must not have after snapshot: %+v", repo.lastAudit.After)
			}
			for _, key := range []string{"customer", "address", "items", "note", "phone"} {
				if _, ok := repo.lastAudit.Before[key]; ok {
					t.Fatalf("audit leaked %q: %+v", key, repo.lastAudit.Before)
				}
			}
		})
	}
}

func TestAdminOrderServiceDeleteErrors(t *testing.T) {
	repo := newFakeAdminOrderRepository()
	service := newAdminOrderTestService(repo)

	if _, err := service.Delete(context.Background(), adminOrderTestActor, "invalido"); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("invalid id error = %v, want ErrNotFound", err)
	}
	if _, err := service.Delete(context.Background(), adminOrderTestActor, adminOrderTestID); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("unknown id error = %v, want ErrNotFound", err)
	}

	repo.orders[adminOrderTestID] = adminOrderFixture()
	repo.deleteErr = domain.ErrConflict
	if _, err := service.Delete(context.Background(), adminOrderTestActor, adminOrderTestID); !errors.Is(err, domain.ErrConflict) {
		t.Errorf("repository error = %v, want ErrConflict", err)
	}
	if repo.deleted {
		t.Error("failed delete must not be reported as done")
	}
}
