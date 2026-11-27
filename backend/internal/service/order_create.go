package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/brocados-lanches/api/internal/domain"
	"github.com/brocados-lanches/api/internal/platform/token"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type publicIDGenerator func() (string, error)

// Create valida, calcula e persiste um pedido. Nenhum valor financeiro vem do
// cliente; o mesmo domain.BuildQuote usado pela cotação é aplicado aqui.
func (s *OrderService) Create(ctx context.Context, request domain.OrderRequest, idempotencyKey, accessToken string) (domain.Order, bool, error) {
	if err := validateIdempotencyKey(idempotencyKey); err != nil {
		return domain.Order{}, false, err
	}
	if err := request.Validate(); err != nil {
		return domain.Order{}, false, err
	}
	accessToken, accessTokenHash, err := s.resolveOrderAccessToken(idempotencyKey, accessToken)
	if err != nil {
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
		if err := s.attachAccessToken(ctx, &existing, accessToken, accessTokenHash); err != nil {
			return domain.Order{}, false, err
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
	order.AccessTokenHash = accessTokenHash

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
			if err := s.attachAccessToken(ctx, &existing, accessToken, accessTokenHash); err != nil {
				return domain.Order{}, false, err
			}
			return existing, true, nil
		}
		return domain.Order{}, false, err
	}
	// O repositório devolve o documento persistido, sem o token em claro.
	stored.AccessToken = accessToken
	s.syncCustomerProfile(ctx, stored, request)
	return stored, false, nil
}

// resolveOrderAccessToken usa o token já guardado no navegador quando presente.
// Na primeira compra, o servidor emite o token a partir da chave de idempotência
// e de um segredo que nunca sai do servidor.
func (s *OrderService) resolveOrderAccessToken(idempotencyKey, accessToken string) (string, string, error) {
	if accessToken == "" {
		accessToken = token.DeriveOrderAccess(s.orderAccessSecret, idempotencyKey)
	}
	if err := token.ValidateOrderAccess(accessToken); err != nil {
		return "", "", &domain.ValidationError{
			Field:   "accessToken",
			Code:    domain.CodeInvalid,
			Message: "invalid history token",
		}
	}
	return accessToken, token.Hash(accessToken), nil
}

// attachAccessToken mantém o retry idempotente associado ao token apresentado
// pelo navegador. Também preenche pedidos gravados antes do histórico existir.
func (s *OrderService) attachAccessToken(ctx context.Context, order *domain.Order, accessToken, accessTokenHash string) error {
	if order.AccessTokenHash == accessTokenHash {
		order.AccessToken = accessToken
		return nil
	}
	if err := s.reader.SetAccessTokenHash(ctx, order.ID, accessTokenHash); err != nil {
		return fmt.Errorf("attach order access token: %w", err)
	}
	order.AccessToken = accessToken
	order.AccessTokenHash = accessTokenHash
	return nil
}

// syncCustomerProfile registra os dados de conveniência depois que o pedido já
// foi gravado. Uma falha aqui não desfaz nem bloqueia o pedido.
func (s *OrderService) syncCustomerProfile(ctx context.Context, order domain.Order, request domain.OrderRequest) {
	if s.profiles == nil {
		return
	}

	now := time.Now().UTC()
	profile, err := domain.NewCustomerProfileFromOrder(order, now)
	if err != nil {
		s.logger.Warn("customer profile skipped", "reason", "invalid_phone")
		return
	}

	if request.SaveCustomerProfile != nil && !*request.SaveCustomerProfile {
		err = s.profiles.OptOut(ctx, profile.PhoneKey, now)
	} else {
		err = s.profiles.Upsert(ctx, profile)
	}
	if err != nil {
		// O erro pode carregar detalhe do driver; só a causa genérica entra no log.
		s.logger.Warn("customer profile update failed", "reason", "write_failed")
	}
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
