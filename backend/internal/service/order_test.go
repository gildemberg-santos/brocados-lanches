package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/brocados-lanches/api/internal/domain"
)

type fakeOrderReader struct {
	products []domain.Product
	addons   []domain.Addon
	zones    []domain.DeliveryZone
	store    domain.StoreSettings

	productsErr error
	addonsErr   error
	zonesErr    error
	storeErr    error

	existing      domain.Order
	existingFound bool
	existingErr   error
	created       domain.Order
	createErr     error
	byPublicID    domain.Order
	byPublicIDErr error

	history      []domain.Order
	historyTotal int64
	historyErr   error
	historyQuery domain.OrderHistoryQuery
	historyHash  string
	attachedID   string
	attachedHash string
	attachErr    error

	receivedOrder domain.Order

	calls []string
}

func (f *fakeOrderReader) ListActiveProducts(context.Context) ([]domain.Product, error) {
	f.calls = append(f.calls, "products")
	return f.products, f.productsErr
}

func (f *fakeOrderReader) ListActiveAddons(context.Context) ([]domain.Addon, error) {
	f.calls = append(f.calls, "addons")
	return f.addons, f.addonsErr
}

func (f *fakeOrderReader) ListActiveDeliveryZones(context.Context) ([]domain.DeliveryZone, error) {
	f.calls = append(f.calls, "zones")
	return f.zones, f.zonesErr
}

func (f *fakeOrderReader) GetStoreSettings(context.Context) (domain.StoreSettings, error) {
	f.calls = append(f.calls, "store")
	return f.store, f.storeErr
}

func (f *fakeOrderReader) FindByIdempotencyKey(_ context.Context, key string) (domain.Order, bool, error) {
	f.calls = append(f.calls, "findByIdempotency")
	return f.existing, f.existingFound, f.existingErr
}

func (f *fakeOrderReader) FindByPublicID(context.Context, string) (domain.Order, error) {
	f.calls = append(f.calls, "findByPublicID")
	return f.byPublicID, f.byPublicIDErr
}

func (f *fakeOrderReader) ListByAccessToken(_ context.Context, hash string, query domain.OrderHistoryQuery) ([]domain.Order, int64, error) {
	f.calls = append(f.calls, "listByAccessToken")
	f.historyHash = hash
	f.historyQuery = query
	return f.history, f.historyTotal, f.historyErr
}

func (f *fakeOrderReader) SetAccessTokenHash(_ context.Context, id, hash string) error {
	f.calls = append(f.calls, "setAccessTokenHash")
	f.attachedID = id
	f.attachedHash = hash
	return f.attachErr
}

func (f *fakeOrderReader) Create(_ context.Context, order domain.Order) (domain.Order, error) {
	f.calls = append(f.calls, "create")
	f.receivedOrder = order
	if f.createErr != nil {
		return domain.Order{}, f.createErr
	}
	return f.created, nil
}

func orderServiceFixture() (*fakeOrderReader, *OrderService) {
	reader := &fakeOrderReader{
		products: []domain.Product{{
			ID:         "prod-1",
			Name:       "Bacon Cheddar",
			PriceCents: 2200,
			Available:  true,
			Active:     true,
			AddonIDs:   []string{"addon-1"},
			OptionGroups: []domain.OptionGroup{{
				ID:            "size",
				Name:          "Tamanho",
				Required:      true,
				MinSelections: 1,
				MaxSelections: 1,
				Options: []domain.Option{
					{ID: "medium", Name: "Ao ponto", Available: true},
				},
			}},
		}},
		addons: []domain.Addon{{ID: "addon-1", Name: "Bacon", PriceCents: 500, Active: true}},
		zones:  []domain.DeliveryZone{{ID: "zone-1", Name: "Centro", FeeCents: 800, Active: true}},
		store:  domain.StoreSettings{IsOpen: true},
	}
	service, err := NewOrderService(
		reader,
		time.FixedZone("UTC-3", -3*60*60),
		WithOrderAccessSecret("test-order-access-secret"),
	)
	if err != nil {
		panic(err)
	}
	return reader, service
}

func orderServiceRequest() domain.QuoteRequest {
	return domain.QuoteRequest{
		Items: []domain.QuoteItem{{
			ProductID: "prod-1",
			Quantity:  2,
			OptionIDs: []string{"medium"},
			AddonIDs:  []string{"addon-1"},
		}},
		Fulfillment: domain.FulfillmentDelivery,
		ZoneID:      "zone-1",
	}
}

func TestNewOrderServiceRequiresDependencies(t *testing.T) {
	location := time.UTC

	if _, err := NewOrderService(nil, location); err == nil {
		t.Error("NewOrderService(nil, location) error = nil, want error")
	}
	if _, err := NewOrderService(&fakeOrderReader{}, nil); err == nil {
		t.Error("NewOrderService(reader, nil) error = nil, want error")
	}
	if _, err := NewOrderService(&fakeOrderReader{}, time.UTC); err == nil {
		t.Error("NewOrderService(reader, location) without secret error = nil, want error")
	}
}

func TestOrderServiceQuoteUsesCatalogValues(t *testing.T) {
	_, service := orderServiceFixture()

	quote, err := service.Quote(context.Background(), orderServiceRequest())
	if err != nil {
		t.Fatalf("Quote() error = %v", err)
	}
	if quote.SubtotalCents != 5400 || quote.DeliveryFeeCents != 800 || quote.TotalCents != 6200 {
		t.Errorf("totals = %d/%d/%d, want 5400/800/6200", quote.SubtotalCents, quote.DeliveryFeeCents, quote.TotalCents)
	}
}

func TestOrderServiceQuotePickupIgnoresZone(t *testing.T) {
	_, service := orderServiceFixture()
	request := orderServiceRequest()
	request.Fulfillment = domain.FulfillmentPickup
	request.ZoneID = "does-not-exist"

	quote, err := service.Quote(context.Background(), request)
	if err != nil {
		t.Fatalf("Quote() error = %v", err)
	}
	if quote.DeliveryFeeCents != 0 || quote.TotalCents != 5400 {
		t.Errorf("pickup totals = %d/%d, want 0/5400", quote.DeliveryFeeCents, quote.TotalCents)
	}
}

func TestOrderServiceQuoteValidatesBeforeReading(t *testing.T) {
	reader, service := orderServiceFixture()
	request := orderServiceRequest()
	request.Items[0].Quantity = 0

	if _, err := service.Quote(context.Background(), request); err == nil {
		t.Fatal("Quote() error = nil, want validation error")
	}
	if len(reader.calls) != 0 {
		t.Errorf("reader calls = %v, want none", reader.calls)
	}
}

func TestOrderServiceQuoteReaderErrors(t *testing.T) {
	tests := []struct {
		name     string
		mutate   func(*fakeOrderReader)
		wantText string
	}{
		{
			name:     "produtos",
			mutate:   func(reader *fakeOrderReader) { reader.productsErr = errors.New("mongo down") },
			wantText: "list active products",
		},
		{
			name:     "adicionais",
			mutate:   func(reader *fakeOrderReader) { reader.addonsErr = errors.New("mongo down") },
			wantText: "list active addons",
		},
		{
			name:     "zonas",
			mutate:   func(reader *fakeOrderReader) { reader.zonesErr = errors.New("mongo down") },
			wantText: "list active delivery zones",
		},
		{
			name:     "loja",
			mutate:   func(reader *fakeOrderReader) { reader.storeErr = errors.New("mongo down") },
			wantText: "get store settings",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			reader, service := orderServiceFixture()
			test.mutate(reader)

			_, err := service.Quote(context.Background(), orderServiceRequest())
			if err == nil {
				t.Fatal("Quote() error = nil, want error")
			}
			if !strings.Contains(err.Error(), test.wantText) {
				t.Errorf("error = %q, want context %q", err.Error(), test.wantText)
			}
		})
	}
}

func TestOrderServiceQuoteStoreClosed(t *testing.T) {
	reader, service := orderServiceFixture()
	reader.store.IsOpen = false

	if _, err := service.Quote(context.Background(), orderServiceRequest()); !errors.Is(err, domain.ErrStoreClosed) {
		t.Errorf("Quote() error = %v, want %v", err, domain.ErrStoreClosed)
	}
}

func TestOrderServiceQuoteZoneUnavailable(t *testing.T) {
	reader, service := orderServiceFixture()
	reader.zones[0].Active = false

	if _, err := service.Quote(context.Background(), orderServiceRequest()); !errors.Is(err, domain.ErrZoneUnavailable) {
		t.Errorf("Quote() error = %v, want %v", err, domain.ErrZoneUnavailable)
	}
}
