package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/brocados-lanches/api/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type publicIDGenerator func() (string, error)

// Create valida, calcula e persiste um pedido. Nenhum valor financeiro vem do
// cliente; o mesmo domain.BuildQuote usado pela cotação é aplicado aqui.
func (s *OrderService) Create(ctx context.Context, request domain.OrderRequest, idempotencyKey string) (domain.Order, bool, error) {
	if err := validateIdempotencyKey(idempotencyKey); err != nil {
		return domain.Order{}, false, err
	}
	if err := request.Validate(); err != nil {
		return domain.Order{}, false, err
	}

	requestHash, err := domain.CanonicalRequestHash(request)
	if err != nil {
		return domain.Order{}, false, err
	}

	existing, found, err := s.reader.FindByIdempotencyKey(ctx, idempotencyKey)
	if err != nil {
		return domain.Order{}, false, fmt.Errorf("find existing order: %w", err)
	}
	if found {
		if existing.RequestHash != requestHash {
			return domain.Order{}, false, domain.ErrConflict
		}
		return existing, true, nil
	}

	catalog, err := s.loadCatalog(ctx)
	if err != nil {
		return domain.Order{}, false, err
	}
	quote, err := domain.BuildQuote(quoteFromOrderRequest(request), catalog)
	if err != nil {
		return domain.Order{}, false, err
	}
	if err := request.ValidateWithTotal(quote.TotalCents); err != nil {
		return domain.Order{}, false, err
	}

	publicID, err := s.publicID()
	if err != nil {
		return domain.Order{}, false, fmt.Errorf("generate public id: %w", err)
	}
	order := domain.NewOrder(
		request,
		quote,
		quote.Zone,
		bson.NewObjectID().Hex(),
		0,
		publicID,
		idempotencyKey,
		requestHash,
		time.Now().UTC(),
	)

	stored, err := s.reader.Create(ctx, order)
	if err != nil {
		// Concorrência na mesma chave: a unique index decide e o pedido já
		// gravado é devolvido, sem criar duplicata.
		if errors.Is(err, domain.ErrConflict) {
			existing, found, findErr := s.reader.FindByIdempotencyKey(ctx, idempotencyKey)
			if findErr != nil {
				return domain.Order{}, false, fmt.Errorf("reload idempotent order: %w", findErr)
			}
			if !found || existing.RequestHash != requestHash {
				return domain.Order{}, false, domain.ErrConflict
			}
			return existing, true, nil
		}
		return domain.Order{}, false, err
	}
	return stored, false, nil
}

// GetByPublicID devolve o pedido para acompanhamento. O handler escolhe quais
// campos são públicos; nenhum dado pessoal sai daqui diretamente.
func (s *OrderService) GetByPublicID(ctx context.Context, publicID string) (domain.Order, error) {
	if err := validatePublicID(publicID); err != nil {
		return domain.Order{}, err
	}
	return s.reader.FindByPublicID(ctx, publicID)
}

// StoreSettings expõe a configuração atual para construir o link de WhatsApp.
func (s *OrderService) StoreSettings(ctx context.Context) (domain.StoreSettings, error) {
	return s.reader.GetStoreSettings(ctx)
}

func (s *OrderService) loadCatalog(ctx context.Context) (domain.QuoteCatalog, error) {
	products, err := s.reader.ListActiveProducts(ctx)
	if err != nil {
		return domain.QuoteCatalog{}, fmt.Errorf("list active products: %w", err)
	}
	addons, err := s.reader.ListActiveAddons(ctx)
	if err != nil {
		return domain.QuoteCatalog{}, fmt.Errorf("list active addons: %w", err)
	}
	zones, err := s.reader.ListActiveDeliveryZones(ctx)
	if err != nil {
		return domain.QuoteCatalog{}, fmt.Errorf("list active delivery zones: %w", err)
	}
	store, err := s.reader.GetStoreSettings(ctx)
	if err != nil {
		return domain.QuoteCatalog{}, fmt.Errorf("get store settings: %w", err)
	}
	return domain.QuoteCatalog{
		Products: products,
		Addons:   addons,
		Zones:    zones,
		Store:    store,
		Location: s.location,
		Now:      time.Now(),
	}, nil
}

func quoteFromOrderRequest(request domain.OrderRequest) domain.QuoteRequest {
	return domain.QuoteRequest{
		Items:       request.Items,
		Fulfillment: request.Fulfillment,
		ZoneID:      request.ZoneID,
	}
}

func validateIdempotencyKey(key string) error {
	if len(key) < domain.MinIdempotencyKeyLength || len(key) > domain.MaxIdempotencyKeyLength {
		return &domain.ValidationError{
			Field:   "Idempotency-Key",
			Code:    domain.CodeOutOfRange,
			Message: fmt.Sprintf("must have between %d and %d characters", domain.MinIdempotencyKeyLength, domain.MaxIdempotencyKeyLength),
		}
	}
	return nil
}

func validatePublicID(publicID string) error {
	if publicID == "" || len(publicID) > 64 {
		return &domain.ValidationError{Field: "publicId", Code: domain.CodeInvalid, Message: "invalid public identifier"}
	}
	return nil
}

func newPublicID() (string, error) {
	buffer := make([]byte, 16)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buffer), nil
}
