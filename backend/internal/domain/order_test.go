package domain

import (
	"errors"
	"math"
	"strings"
	"testing"
	"time"
)

func quoteFixture() (QuoteRequest, QuoteCatalog) {
	location := time.FixedZone("UTC-3", -3*60*60)
	request := QuoteRequest{
		Items: []QuoteItem{{
			ProductID: "prod-1",
			Quantity:  2,
			OptionIDs: []string{"medium"},
			AddonIDs:  []string{"addon-1"},
		}},
		Fulfillment: FulfillmentDelivery,
		ZoneID:      "zone-1",
	}
	catalog := QuoteCatalog{
		Location: location,
		Now:      time.Date(2026, 11, 22, 22, 0, 0, 0, time.UTC), // domingo, 19:00 local
		Store:    StoreSettings{IsOpen: true},
		Products: []Product{{
			ID:         "prod-1",
			Name:       "Bacon Cheddar",
			PriceCents: 2200,
			Available:  true,
			Active:     true,
			AddonIDs:   []string{"addon-1"},
			OptionGroups: []OptionGroup{{
				ID:            "doneness",
				Name:          "Ponto da carne",
				Required:      true,
				MinSelections: 1,
				MaxSelections: 1,
				Options: []Option{
					{ID: "medium", Name: "Ao ponto", Available: true},
					{ID: "well-done", Name: "Bem passado", PriceCents: 200, Available: true},
					{ID: "rare", Name: "Mal passado", Available: false},
				},
			}},
		}},
		Addons: []Addon{{ID: "addon-1", Name: "Bacon", PriceCents: 500, Active: true}},
		Zones:  []DeliveryZone{{ID: "zone-1", Name: "Centro", FeeCents: 800, Active: true}},
	}
	return request, catalog
}

func TestBuildQuoteDelivery(t *testing.T) {
	request, catalog := quoteFixture()

	quote, err := BuildQuote(request, catalog)
	if err != nil {
		t.Fatalf("BuildQuote() error = %v", err)
	}
	if quote.SubtotalCents != 5400 || quote.DeliveryFeeCents != 800 || quote.TotalCents != 6200 {
		t.Errorf("totals = %d/%d/%d, want 5400/800/6200", quote.SubtotalCents, quote.DeliveryFeeCents, quote.TotalCents)
	}
	if quote.Currency != CurrencyBRL {
		t.Errorf("currency = %q, want %q", quote.Currency, CurrencyBRL)
	}
	if len(quote.Items) != 1 {
		t.Fatalf("len(items) = %d, want 1", len(quote.Items))
	}
	line := quote.Items[0]
	if line.ProductID != "prod-1" || line.Name != "Bacon Cheddar" || line.Quantity != 2 {
		t.Errorf("line = %+v", line)
	}
	if line.UnitPriceCents != 2200 || line.OptionsPriceCents != 500 || line.LineTotalCents != 5400 {
		t.Errorf("line prices = %+v", line)
	}
	if len(line.Options) != 2 || line.Options[0].Name != "Ao ponto" || line.Options[1].PriceCents != 500 {
		t.Errorf("options = %+v", line.Options)
	}
	if quote.Warnings == nil {
		t.Error("warnings = nil, want empty slice")
	}
}

func TestBuildQuotePickupIgnoresZone(t *testing.T) {
	request, catalog := quoteFixture()
	request.Fulfillment = FulfillmentPickup
	request.ZoneID = " nonexistent"

	quote, err := BuildQuote(request, catalog)
	if err != nil {
		t.Fatalf("BuildQuote() error = %v", err)
	}
	if quote.DeliveryFeeCents != 0 || quote.TotalCents != 5400 {
		t.Errorf("pickup totals = %d/%d, want 0/5400", quote.DeliveryFeeCents, quote.TotalCents)
	}
}

func TestBuildQuoteAvailabilityErrors(t *testing.T) {
	tests := []struct {
		name      string
		mutate    func(*QuoteRequest, *QuoteCatalog)
		wantError error
	}{
		{
			name: "loja fechada pelo kill switch",
			mutate: func(_ *QuoteRequest, catalog *QuoteCatalog) {
				catalog.Store.IsOpen = false
			},
			wantError: ErrStoreClosed,
		},
		{
			name: "loja fechada por horário",
			mutate: func(_ *QuoteRequest, catalog *QuoteCatalog) {
				catalog.Store.OpeningHours = []OpeningHours{{
					Day:    "sunday",
					Closed: true,
				}}
			},
			wantError: ErrStoreClosed,
		},
		{
			name: "produto esgotado",
			mutate: func(_ *QuoteRequest, catalog *QuoteCatalog) {
				catalog.Products[0].Available = false
			},
			wantError: ErrProductUnavailable,
		},
		{
			name: "produto fora do catálogo ativo",
			mutate: func(request *QuoteRequest, _ *QuoteCatalog) {
				request.Items[0].ProductID = "missing"
			},
			wantError: ErrProductUnavailable,
		},
		{
			name: "zona inexistente",
			mutate: func(request *QuoteRequest, _ *QuoteCatalog) {
				request.ZoneID = "missing"
			},
			wantError: ErrZoneUnavailable,
		},
		{
			name: "zona inativa",
			mutate: func(_ *QuoteRequest, catalog *QuoteCatalog) {
				catalog.Zones[0].Active = false
			},
			wantError: ErrZoneUnavailable,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request, catalog := quoteFixture()
			test.mutate(&request, &catalog)

			if _, err := BuildQuote(request, catalog); !errors.Is(err, test.wantError) {
				t.Errorf("BuildQuote() error = %v, want %v", err, test.wantError)
			}
		})
	}
}

func TestBuildQuoteValidation(t *testing.T) {
	longNote := strings.Repeat("a", MaxNoteLength+1)
	tooManyItems := make([]QuoteItem, MaxQuoteItems+1)

	tests := []struct {
		name      string
		mutate    func(*QuoteRequest)
		wantField string
		wantCode  string
	}{
		{
			name: "sem itens",
			mutate: func(request *QuoteRequest) {
				request.Items = nil
			},
			wantField: "items",
			wantCode:  CodeRequired,
		},
		{
			name: "itens demais",
			mutate: func(request *QuoteRequest) {
				request.Items = tooManyItems
			},
			wantField: "items",
			wantCode:  CodeOutOfRange,
		},
		{
			name: "quantidade zero",
			mutate: func(request *QuoteRequest) {
				request.Items[0].Quantity = 0
			},
			wantField: "items[0].quantity",
			wantCode:  CodeOutOfRange,
		},
		{
			name: "quantidade acima do limite",
			mutate: func(request *QuoteRequest) {
				request.Items[0].Quantity = MaxQuoteQuantity + 1
			},
			wantField: "items[0].quantity",
			wantCode:  CodeOutOfRange,
		},
		{
			name: "opção duplicada",
			mutate: func(request *QuoteRequest) {
				request.Items[0].OptionIDs = []string{"medium", "medium"}
			},
			wantField: "items[0].optionIds[1]",
			wantCode:  CodeDuplicate,
		},
		{
			name: "opção desconhecida",
			mutate: func(request *QuoteRequest) {
				request.Items[0].OptionIDs = []string{"unknown"}
			},
			wantField: "items[0].optionIds[0]",
			wantCode:  CodeInvalid,
		},
		{
			name: "opção indisponível",
			mutate: func(request *QuoteRequest) {
				request.Items[0].OptionIDs = []string{"rare"}
			},
			wantField: "items[0].optionIds[0]",
			wantCode:  CodeInvalid,
		},
		{
			name: "grupo obrigatório sem escolha",
			mutate: func(request *QuoteRequest) {
				request.Items[0].OptionIDs = nil
			},
			wantField: "items[0].optionIds",
			wantCode:  CodeRequired,
		},
		{
			name: "grupo com escolhas demais",
			mutate: func(request *QuoteRequest) {
				request.Items[0].OptionIDs = []string{"medium", "well-done"}
			},
			wantField: "items[0].optionIds",
			wantCode:  CodeOutOfRange,
		},
		{
			name: "adicional não vinculado",
			mutate: func(request *QuoteRequest) {
				request.Items[0].OptionIDs = []string{"medium"}
				request.Items[0].AddonIDs = []string{"unlinked"}
			},
			wantField: "items[0].addonIds[0]",
			wantCode:  CodeInvalid,
		},
		{
			name: "observação longa",
			mutate: func(request *QuoteRequest) {
				request.Items[0].Note = longNote
			},
			wantField: "items[0].note",
			wantCode:  CodeTooLong,
		},
		{
			name: "fulfillment inválido",
			mutate: func(request *QuoteRequest) {
				request.Fulfillment = "courier"
			},
			wantField: "fulfillment",
			wantCode:  CodeInvalid,
		},
		{
			name: "entrega sem zona",
			mutate: func(request *QuoteRequest) {
				request.ZoneID = ""
			},
			wantField: "zoneId",
			wantCode:  CodeRequired,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request, catalog := quoteFixture()
			test.mutate(&request)

			_, err := BuildQuote(request, catalog)
			var validationError *ValidationError
			if !errors.As(err, &validationError) {
				t.Fatalf("BuildQuote() error = %v, want ValidationError", err)
			}
			if validationError.Field != test.wantField || validationError.Code != test.wantCode {
				t.Errorf("validation = %s/%s, want %s/%s", validationError.Field, validationError.Code, test.wantField, test.wantCode)
			}
		})
	}
}

func TestBuildQuoteRejectsMonetaryOverflow(t *testing.T) {
	request, catalog := quoteFixture()
	catalog.Products[0].PriceCents = math.MaxInt64

	if _, err := BuildQuote(request, catalog); !errors.Is(err, ErrValidation) {
		t.Errorf("BuildQuote() error = %v, want %v", err, ErrValidation)
	}
}
