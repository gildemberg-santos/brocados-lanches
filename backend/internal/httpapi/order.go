package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/brocados-lanches/api/internal/domain"
	"github.com/brocados-lanches/api/internal/platform/ratelimit"
	"github.com/brocados-lanches/api/internal/platform/token"
	"github.com/brocados-lanches/api/internal/realtime"
	"github.com/brocados-lanches/api/internal/service"
	"github.com/go-chi/chi/v5"
)

// maxOrderBodyBytes aplica o limite de 1 MB do contrato para corpos JSON.
const maxOrderBodyBytes = 1 << 20

// maxQuoteBodyBytes é mantido como alias para os testes da cotação.
const maxQuoteBodyBytes = maxOrderBodyBytes

// OrderService é a dependência dos handlers de pedido.
type OrderService interface {
	Quote(ctx context.Context, request domain.QuoteRequest) (domain.Quote, error)
	Create(ctx context.Context, request domain.OrderRequest, idempotencyKey, accessToken string) (domain.Order, bool, error)
	GetByPublicID(ctx context.Context, publicID string) (domain.Order, error)
	ListByAccessToken(ctx context.Context, accessToken string, query domain.OrderHistoryQuery) (service.OrderHistoryPage, error)
	StoreSettings(ctx context.Context) (domain.StoreSettings, error)
}

type quoteEnvelope struct {
	Data quoteData `json:"data"`
}

type quoteData struct {
	SubtotalCents    int64           `json:"subtotalCents"`
	DeliveryFeeCents int64           `json:"deliveryFeeCents"`
	TotalCents       int64           `json:"totalCents"`
	Currency         string          `json:"currency"`
	Items            []quoteLineData `json:"items"`
	Warnings         []string        `json:"warnings"`
}

type quoteLineData struct {
	ProductID         string            `json:"productId"`
	Name              string            `json:"name"`
	Quantity          int               `json:"quantity"`
	UnitPriceCents    int64             `json:"unitPriceCents"`
	OptionsPriceCents int64             `json:"optionsPriceCents"`
	LineTotalCents    int64             `json:"lineTotalCents"`
	Options           []quoteOptionData `json:"options"`
}

type quoteOptionData struct {
	Name       string `json:"name"`
	PriceCents int64  `json:"priceCents"`
}

type orderEnvelope struct {
	Data orderData `json:"data"`
}

type orderData struct {
	OrderNumber      int64  `json:"orderNumber"`
	PublicID         string `json:"publicId"`
	Status           string `json:"status"`
	Fulfillment      string `json:"fulfillment"`
	SubtotalCents    int64  `json:"subtotalCents"`
	DeliveryFeeCents int64  `json:"deliveryFeeCents"`
	TotalCents       int64  `json:"totalCents"`
	CreatedAt        string `json:"createdAt"`
	WhatsAppURL      string `json:"whatsappUrl"`
	AccessToken      string `json:"accessToken"`
}

type trackingEnvelope struct {
	Data trackingData `json:"data"`
}

type trackingData struct {
	OrderNumber      int64              `json:"orderNumber"`
	PublicID         string             `json:"publicId"`
	Status           string             `json:"status"`
	Fulfillment      string             `json:"fulfillment"`
	PaymentMethod    string             `json:"paymentMethod"`
	Items            []trackingItemData `json:"items"`
	SubtotalCents    int64              `json:"subtotalCents"`
	DeliveryFeeCents int64              `json:"deliveryFeeCents"`
	TotalCents       int64              `json:"totalCents"`
	CreatedAt        string             `json:"createdAt"`
	UpdatedAt        string             `json:"updatedAt"`
	EstimatedTime    string             `json:"estimatedTime,omitempty"`
}

type trackingItemData struct {
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
}

type orderHistoryEnvelope struct {
	Data []trackingData `json:"data"`
	Meta adminMeta      `json:"meta"`
}

// handleQuote cota um pedido sem persistir. O cliente envia apenas IDs e
// escolhas; todo valor monetário é calculado no service.
func handleQuote(orders OrderService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if orders == nil {
			writeError(w, r, logger, errors.New("order service is not configured"))
			return
		}

		request, err := decodeQuoteRequest(w, r)
		if err != nil {
			writeError(w, r, logger, err)
			return
		}

		quote, err := orders.Quote(r.Context(), request)
		if err != nil {
			writeError(w, r, logger, err)
			return
		}

		writeJSON(w, http.StatusOK, quoteEnvelope{Data: newQuoteData(quote)})
	}
}

// decodeQuoteRequest lê o corpo uma única vez e rejeita campos desconhecidos.
func decodeQuoteRequest(w http.ResponseWriter, r *http.Request) (domain.QuoteRequest, error) {
	var request domain.QuoteRequest

	body := http.MaxBytesReader(w, r.Body, maxQuoteBodyBytes)
	decoder := json.NewDecoder(body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		return domain.QuoteRequest{}, &malformedRequestError{
			field:   "body",
			code:    domain.CodeInvalid,
			message: "invalid JSON body",
		}
	}

	return request, nil
}

func newQuoteData(quote domain.Quote) quoteData {
	data := quoteData{
		SubtotalCents:    quote.SubtotalCents,
		DeliveryFeeCents: quote.DeliveryFeeCents,
		TotalCents:       quote.TotalCents,
		Currency:         quote.Currency,
		Items:            make([]quoteLineData, 0, len(quote.Items)),
		Warnings:         append([]string{}, quote.Warnings...),
	}

	for _, item := range quote.Items {
		options := make([]quoteOptionData, 0, len(item.Options))
		for _, option := range item.Options {
			options = append(options, quoteOptionData{
				Name:       option.Name,
				PriceCents: option.PriceCents,
			})
		}
		data.Items = append(data.Items, quoteLineData{
			ProductID:         item.ProductID,
			Name:              item.Name,
			Quantity:          item.Quantity,
			UnitPriceCents:    item.UnitPriceCents,
			OptionsPriceCents: item.OptionsPriceCents,
			LineTotalCents:    item.LineTotalCents,
			Options:           options,
		})
	}

	return data
}

// handleCreateOrder persiste o pedido antes de devolver o link do WhatsApp.
func handleCreateOrder(
	orders OrderService,
	phoneLimiter *ratelimit.Limiter,
	hub *realtime.Hub,
	notifier OrderNotifier,
	logger *slog.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if orders == nil {
			writeError(w, r, logger, errors.New("order service is not configured"))
			return
		}

		idempotencyKey := r.Header.Get("Idempotency-Key")
		if err := validateIdempotencyHeader(idempotencyKey); err != nil {
			writeError(w, r, logger, err)
			return
		}

		request, err := decodeOrderRequest(w, r)
		if err != nil {
			writeError(w, r, logger, err)
			return
		}
		accessToken, err := parseOrderAccessToken(r)
		if err != nil {
			writeError(w, r, logger, err)
			return
		}

		if !allowOrderPhone(phoneLimiter, request.Customer.Phone) {
			writeRateLimited(w, r)
			return
		}

		order, replayed, err := orders.Create(r.Context(), request, idempotencyKey, accessToken)
		if err != nil {
			writeError(w, r, logger, err)
			return
		}
		if !replayed {
			publishOrderEvent(hub, realtime.EventOrderCreated, order)
			if notifier != nil {
				notifier.Notify(r.Context(), domain.WebhookEvent{
					Type:  domain.WebhookEventOrderCreated,
					Order: order,
				})
			}
		}

		whatsappURL := ""
		store, storeErr := orders.StoreSettings(r.Context())
		if storeErr != nil {
			logger.Warn("whatsapp handoff unavailable: store settings failed")
		} else {
			link, linkErr := service.BuildWhatsAppURL(store, order)
			if linkErr != nil {
				logger.Warn("whatsapp handoff unavailable: invalid store number")
			} else {
				whatsappURL = link
			}
		}

		status := http.StatusCreated
		if replayed {
			status = http.StatusOK
			w.Header().Set("Idempotent-Replayed", "true")
		}
		w.Header().Set("Location", "/api/v1/orders/"+order.PublicID)
		w.Header().Set("Cache-Control", "no-store")
		writeJSON(w, status, orderEnvelope{Data: newOrderData(order, whatsappURL)})
	}
}

// handleOrderHistory devolve os pedidos ligados ao token guardado no navegador.
// O token viaja no cabeçalho Authorization e nunca aparece na URL nem em logs.
func handleOrderHistory(orders OrderService, tokenLimiter *ratelimit.Limiter, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if orders == nil {
			writeError(w, r, logger, errors.New("order service is not configured"))
			return
		}
		accessToken, err := parseOrderAccessToken(r)
		if err != nil {
			writeError(w, r, logger, err)
			return
		}
		if accessToken == "" {
			writeError(w, r, logger, &malformedRequestError{
				field:   "Authorization",
				code:    domain.CodeRequired,
				message: "Bearer token is required",
			})
			return
		}
		if !allowOrderToken(tokenLimiter, accessToken) {
			writeRateLimited(w, r)
			return
		}
		query, err := parseOrderHistoryQuery(r)
		if err != nil {
			writeError(w, r, logger, err)
			return
		}

		page, err := orders.ListByAccessToken(r.Context(), accessToken, query)
		if err != nil {
			writeError(w, r, logger, err)
			return
		}
		items := make([]trackingData, 0, len(page.Items))
		for _, order := range page.Items {
			items = append(items, newTrackingData(order, ""))
		}

		w.Header().Set("Cache-Control", "no-store")
		writeJSON(w, http.StatusOK, orderHistoryEnvelope{
			Data: items,
			Meta: newAdminMeta(page.Page, page.PageSize, page.Total, page.TotalPages),
		})
	}
}

// handleTrackOrder expõe apenas o acompanhamento público, sem dados pessoais.
func handleTrackOrder(orders OrderService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if orders == nil {
			writeError(w, r, logger, errors.New("order service is not configured"))
			return
		}

		order, err := orders.GetByPublicID(r.Context(), chi.URLParam(r, "publicId"))
		if err != nil {
			writeError(w, r, logger, err)
			return
		}
		store, err := orders.StoreSettings(r.Context())
		if err != nil {
			writeError(w, r, logger, err)
			return
		}

		w.Header().Set("Cache-Control", "no-store")
		writeJSON(w, http.StatusOK, trackingEnvelope{Data: newTrackingData(order, store.DeliveryTime)})
	}
}

func decodeOrderRequest(w http.ResponseWriter, r *http.Request) (domain.OrderRequest, error) {
	var request domain.OrderRequest
	body := http.MaxBytesReader(w, r.Body, maxOrderBodyBytes)
	decoder := json.NewDecoder(body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		return domain.OrderRequest{}, &malformedRequestError{
			field:   "body",
			code:    domain.CodeInvalid,
			message: "invalid JSON body",
		}
	}
	return request, nil
}

// parseOrderAccessToken aceita o token de histórico no cabeçalho Bearer.
// Ausência é permitida apenas na criação do pedido, quando o servidor emite um.
func parseOrderAccessToken(r *http.Request) (string, error) {
	raw := strings.TrimSpace(r.Header.Get("Authorization"))
	if raw == "" {
		return "", nil
	}
	if len(raw) < 7 || !strings.EqualFold(raw[:7], "Bearer ") {
		return "", &malformedRequestError{field: "Authorization", code: domain.CodeInvalid, message: "use Bearer token"}
	}
	value := strings.TrimSpace(raw[7:])
	if err := token.ValidateOrderAccess(value); err != nil {
		return "", &malformedRequestError{field: "Authorization", code: domain.CodeInvalid, message: "invalid history token"}
	}
	return value, nil
}

func parseOrderHistoryQuery(r *http.Request) (domain.OrderHistoryQuery, error) {
	values := r.URL.Query()
	page, err := parseOptionalInt(values.Get("page"), "page")
	if err != nil {
		return domain.OrderHistoryQuery{}, err
	}
	pageSize, err := parseOptionalInt(values.Get("pageSize"), "pageSize")
	if err != nil {
		return domain.OrderHistoryQuery{}, err
	}
	return domain.OrderHistoryQuery{Page: page, PageSize: pageSize}, nil
}

func validateIdempotencyHeader(key string) error {
	if len(key) < domain.MinIdempotencyKeyLength || len(key) > domain.MaxIdempotencyKeyLength {
		return &malformedRequestError{
			field:   "Idempotency-Key",
			code:    domain.CodeOutOfRange,
			message: "header is required and must have between 16 and 128 characters",
		}
	}
	return nil
}

func newOrderData(order domain.Order, whatsappURL string) orderData {
	return orderData{
		OrderNumber:      order.Number,
		PublicID:         order.PublicID,
		Status:           order.Status,
		Fulfillment:      string(order.Fulfillment),
		SubtotalCents:    order.SubtotalCents,
		DeliveryFeeCents: order.DeliveryFeeCents,
		TotalCents:       order.TotalCents,
		CreatedAt:        order.CreatedAt.UTC().Format(time.RFC3339),
		WhatsAppURL:      whatsappURL,
		AccessToken:      order.AccessToken,
	}
}

func newTrackingData(order domain.Order, estimatedTime string) trackingData {
	items := make([]trackingItemData, 0, len(order.Items))
	for _, item := range order.Items {
		items = append(items, trackingItemData{Name: item.Name, Quantity: item.Quantity})
	}
	return trackingData{
		OrderNumber:      order.Number,
		PublicID:         order.PublicID,
		Status:           order.Status,
		Fulfillment:      string(order.Fulfillment),
		PaymentMethod:    order.Payment.Method,
		Items:            items,
		SubtotalCents:    order.SubtotalCents,
		DeliveryFeeCents: order.DeliveryFeeCents,
		TotalCents:       order.TotalCents,
		CreatedAt:        order.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:        order.UpdatedAt.UTC().Format(time.RFC3339),
		EstimatedTime:    estimatedTime,
	}
}
