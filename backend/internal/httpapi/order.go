package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/brocados-lanches/api/internal/domain"
)

// maxQuoteBodyBytes aplica o limite de 1 MB do contrato para corpos JSON.
const maxQuoteBodyBytes = 1 << 20

// OrderService é a dependência dos handlers de pedido.
type OrderService interface {
	Quote(ctx context.Context, request domain.QuoteRequest) (domain.Quote, error)
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
		return domain.QuoteRequest{}, &domain.ValidationError{
			Field:   "body",
			Code:    domain.CodeInvalid,
			Message: "invalid JSON body",
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
