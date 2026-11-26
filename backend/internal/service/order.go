package service

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/brocados-lanches/api/internal/domain"
)

// OrderReader é a superfície do repositório necessária para cotar, persistir e
// acompanhar pedidos. O cálculo permanece centralizado em domain.BuildQuote.
type OrderReader interface {
	ListActiveProducts(ctx context.Context) ([]domain.Product, error)
	ListActiveAddons(ctx context.Context) ([]domain.Addon, error)
	ListActiveDeliveryZones(ctx context.Context) ([]domain.DeliveryZone, error)
	GetStoreSettings(ctx context.Context) (domain.StoreSettings, error)
	FindByIdempotencyKey(ctx context.Context, key string) (domain.Order, bool, error)
	FindByPublicID(ctx context.Context, publicID string) (domain.Order, error)
	Create(ctx context.Context, order domain.Order) (domain.Order, error)
}

// OrderService aplica as regras de pedido usando dados autoritativos do banco.
type OrderService struct {
	reader   OrderReader
	location *time.Location
	publicID publicIDGenerator

	profiles CustomerProfileStore
	logger   *slog.Logger
}

// OrderServiceOption adiciona dependências opcionais sem alterar os chamadores
// que só precisam cotar e gravar pedidos.
type OrderServiceOption func(*OrderService)

// WithCustomerProfiles habilita o perfil de conveniência reconhecido por
// telefone. O logger recebe apenas erros genéricos, sem dados pessoais.
func WithCustomerProfiles(store CustomerProfileStore, logger *slog.Logger) OrderServiceOption {
	return func(service *OrderService) {
		if store == nil {
			return
		}
		if logger == nil {
			logger = discardLogger()
		}
		service.profiles = store
		service.logger = logger
	}
}

// NewOrderService vincula o service ao repositório e ao fuso da operação.
func NewOrderService(reader OrderReader, location *time.Location, options ...OrderServiceOption) (*OrderService, error) {
	if reader == nil {
		return nil, errors.New("order reader is required")
	}
	if location == nil {
		return nil, errors.New("operation timezone is required")
	}
	service := &OrderService{
		reader:   reader,
		location: location,
		publicID: newPublicID,
		logger:   discardLogger(),
	}
	for _, option := range options {
		option(service)
	}
	return service, nil
}

// Quote valida a requisição e devolve subtotal, taxa e total sem gravar nada.
// A validação de forma acontece antes das leituras para evitar trabalho inútil.
func (s *OrderService) Quote(ctx context.Context, request domain.QuoteRequest) (domain.Quote, error) {
	if err := request.Validate(); err != nil {
		return domain.Quote{}, err
	}

	catalog, err := s.loadCatalog(ctx)
	if err != nil {
		return domain.Quote{}, err
	}
	return domain.BuildQuote(request, catalog)
}
