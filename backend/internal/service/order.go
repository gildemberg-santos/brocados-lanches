package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/brocados-lanches/api/internal/domain"
)

// OrderReader é a superfície do repositório necessária para cotar e, no futuro,
// persistir pedidos. O cálculo permanece centralizado em domain.BuildQuote.
type OrderReader interface {
	ListActiveProducts(ctx context.Context) ([]domain.Product, error)
	ListActiveAddons(ctx context.Context) ([]domain.Addon, error)
	ListActiveDeliveryZones(ctx context.Context) ([]domain.DeliveryZone, error)
	GetStoreSettings(ctx context.Context) (domain.StoreSettings, error)
}

// OrderService aplica as regras de pedido usando dados autoritativos do banco.
type OrderService struct {
	reader   OrderReader
	location *time.Location
}

// NewOrderService vincula o service ao repositório e ao fuso da operação.
func NewOrderService(reader OrderReader, location *time.Location) (*OrderService, error) {
	if reader == nil {
		return nil, errors.New("order reader is required")
	}
	if location == nil {
		return nil, errors.New("operation timezone is required")
	}
	return &OrderService{reader: reader, location: location}, nil
}

// Quote valida a requisição e devolve subtotal, taxa e total sem gravar nada.
// A validação de forma acontece antes das leituras para evitar trabalho inútil.
func (s *OrderService) Quote(ctx context.Context, request domain.QuoteRequest) (domain.Quote, error) {
	if err := request.Validate(); err != nil {
		return domain.Quote{}, err
	}

	products, err := s.reader.ListActiveProducts(ctx)
	if err != nil {
		return domain.Quote{}, fmt.Errorf("list active products: %w", err)
	}
	addons, err := s.reader.ListActiveAddons(ctx)
	if err != nil {
		return domain.Quote{}, fmt.Errorf("list active addons: %w", err)
	}
	zones, err := s.reader.ListActiveDeliveryZones(ctx)
	if err != nil {
		return domain.Quote{}, fmt.Errorf("list active delivery zones: %w", err)
	}
	store, err := s.reader.GetStoreSettings(ctx)
	if err != nil {
		return domain.Quote{}, fmt.Errorf("get store settings: %w", err)
	}

	return domain.BuildQuote(request, domain.QuoteCatalog{
		Products: products,
		Addons:   addons,
		Zones:    zones,
		Store:    store,
		Location: s.location,
		Now:      time.Now(),
	})
}
