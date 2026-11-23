package service

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/brocados-lanches/api/internal/domain"
)

// concurrentOrderReader simula o índice único do MongoDB: apenas a primeira
// inserção com uma chave vence; as demais recebem ErrConflict e recarregam.
// Todos os campos mutáveis são protegidos para suportar -race.
type concurrentOrderReader struct {
	products []domain.Product
	addons   []domain.Addon
	zones    []domain.DeliveryZone
	store    domain.StoreSettings

	mu          sync.Mutex
	inserted    domain.Order
	insertCalls int
	reloadCalls int
	requestHash string
}

func (r *concurrentOrderReader) ListActiveProducts(context.Context) ([]domain.Product, error) {
	return r.products, nil
}

func (r *concurrentOrderReader) ListActiveAddons(context.Context) ([]domain.Addon, error) {
	return r.addons, nil
}

func (r *concurrentOrderReader) ListActiveDeliveryZones(context.Context) ([]domain.DeliveryZone, error) {
	return r.zones, nil
}

func (r *concurrentOrderReader) GetStoreSettings(context.Context) (domain.StoreSettings, error) {
	return r.store, nil
}

func (r *concurrentOrderReader) FindByIdempotencyKey(context.Context, string) (domain.Order, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.reloadCalls++
	if r.insertCalls == 0 {
		return domain.Order{}, false, nil
	}
	stored := r.inserted
	stored.RequestHash = r.requestHash
	return stored, true, nil
}

func (r *concurrentOrderReader) FindByPublicID(context.Context, string) (domain.Order, error) {
	return domain.Order{}, domain.ErrNotFound
}

func (r *concurrentOrderReader) Create(_ context.Context, order domain.Order) (domain.Order, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.insertCalls++
	if r.insertCalls > 1 {
		return domain.Order{}, domain.ErrConflict
	}
	r.inserted = order
	stored := order
	stored.Number = 1042
	return stored, nil
}

func TestOrderServiceCreateConcurrentSameKeyCreatesOneOrder(t *testing.T) {
	base, service := orderServiceFixture()
	service.publicID = func() (string, error) { return "public-123", nil }
	reader := &concurrentOrderReader{
		products: base.products,
		addons:   base.addons,
		zones:    base.zones,
		store:    base.store,
	}
	service.reader = reader

	request := createOrderRequestFixture()
	hash, err := domain.CanonicalRequestHash(request)
	if err != nil {
		t.Fatal(err)
	}
	reader.requestHash = hash

	const attempts = 12
	type result struct {
		id       string
		replayed bool
		err      error
	}
	results := make(chan result, attempts)

	var wait sync.WaitGroup
	for i := 0; i < attempts; i++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			order, replayed, err := service.Create(context.Background(), request, testIdempotencyKey)
			results <- result{id: order.PublicID, replayed: replayed, err: err}
		}()
	}
	wait.Wait()
	close(results)

	created := 0
	for value := range results {
		if value.err != nil {
			t.Errorf("Create() error = %v", value.err)
			continue
		}
		if value.id != "public-123" {
			t.Errorf("publicId = %q, want public-123", value.id)
		}
		if !value.replayed {
			created++
		}
	}
	if created != 1 {
		t.Errorf("created orders = %d, want 1", created)
	}
	if reader.insertCalls < 1 || reader.insertCalls > attempts {
		t.Errorf("insert attempts = %d, want between 1 and %d", reader.insertCalls, attempts)
	}
}

func TestOrderServiceCreateDoesNotReturnHandoffWhenPersistenceFails(t *testing.T) {
	reader, service := orderServiceFixture()
	reader.createErr = errors.New("mongo indisponível")

	order, replayed, err := service.Create(context.Background(), createOrderRequestFixture(), testIdempotencyKey)
	if !errors.Is(err, reader.createErr) {
		t.Fatalf("error = %v, want persistence failure", err)
	}
	if replayed || order.PublicID != "" || order.Number != 0 {
		t.Errorf("order = %+v, replayed = %v; handoff must not be released", order, replayed)
	}
}
