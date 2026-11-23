package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/brocados-lanches/api/internal/domain"
	"github.com/brocados-lanches/api/internal/platform/ratelimit"
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
	Create(ctx context.Context, request domain.OrderRequest, idempotencyKey string) (domain.Order, bool, error)
	GetByPublicID(ctx context.Context, publicID string) (domain.Order, error)
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
}

type trackingEnvelope struct {
	Data trackingData `json:"data"`
}

type trackingData struct {
	OrderNumber      int64              `json:"orderNumber"`
	PublicID         string             `json:"publicId"`
	Status           string             `json:"status"`
	Fulfillment      string             `json:"fulfillment"`
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
func handleCreateOrder(orders OrderService, phoneLimiter *ratelimit.Limiter, logger *slog.Logger) http.HandlerFunc {
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

		if !allowOrderPhone(phoneLimiter, request.Customer.Phone) {
			writeRateLimited(w, r)
			return
		}

		order, replayed, err := orders.Create(r.Context(), request, idempotencyKey)
		if err != nil {
			writeError(w, r, logger, err)
			return
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
		Items:            items,
		SubtotalCents:    order.SubtotalCents,
		DeliveryFeeCents: order.DeliveryFeeCents,
		TotalCents:       order.TotalCents,
		CreatedAt:        order.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:        order.UpdatedAt.UTC().Format(time.RFC3339),
		EstimatedTime:    estimatedTime,
	}
}
