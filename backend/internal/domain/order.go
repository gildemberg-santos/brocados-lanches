package domain

import (
	"fmt"
	"math"
	"time"
)

// Limites do pedido e da cotação. São regras de proteção da API, não dados de negócio.
const (
	MaxQuoteItems    = 50
	MaxQuoteQuantity = 99
	MaxNoteLength    = 140
	CurrencyBRL      = "BRL"
)

// Tipos de entrada congelados no snapshot do pedido.
const (
	QuoteOptionKindOption = "option"
	QuoteOptionKindAddon  = "addon"
)

// Fulfillment é a forma de entrega do pedido.
type Fulfillment string

const (
	FulfillmentDelivery Fulfillment = "delivery"
	FulfillmentPickup   Fulfillment = "pickup"
)

// QuoteRequest contém apenas identificadores e escolhas. Preço, taxa, total,
// disponibilidade e status nunca vêm do cliente.
type QuoteRequest struct {
	Items       []QuoteItem `json:"items"`
	Fulfillment Fulfillment `json:"fulfillment"`
	ZoneID      string      `json:"zoneId,omitempty"`
}

// QuoteItem é um item enviado pelo carrinho. AddonIDs e Note são opcionais
// para manter compatibilidade com o contrato original da cotação.
type QuoteItem struct {
	ProductID string   `json:"productId"`
	Quantity  int      `json:"quantity"`
	OptionIDs []string `json:"optionIds"`
	AddonIDs  []string `json:"addonIds,omitempty"`
	Note      string   `json:"note,omitempty"`
}

// Quote é o resultado da cotação. Nenhum valor é persistido.
type Quote struct {
	SubtotalCents    int64
	DeliveryFeeCents int64
	TotalCents       int64
	Currency         string
	Zone             *OrderZone
	Items            []QuoteLine
	Warnings         []string
}

// QuoteLine é o item normalizado com os valores congelados no momento da cotação.
type QuoteLine struct {
	ProductID         string
	Name              string
	Quantity          int
	UnitPriceCents    int64
	OptionsPriceCents int64
	LineTotalCents    int64
	Options           []QuoteOption
}

// QuoteOption é uma opção ou adicional com nome e preço vindos do catálogo.
type QuoteOption struct {
	GroupID    string
	GroupName  string
	OptionID   string
	Name       string
	Kind       string
	PriceCents int64
}

// QuoteCatalog reúne os dados autoritativos necessários para calcular a cotação.
type QuoteCatalog struct {
	Products []Product
	Addons   []Addon
	Zones    []DeliveryZone
	Store    StoreSettings
	Location *time.Location
	Now      time.Time
}

// Validate confere a forma da requisição antes de qualquer leitura do catálogo.
func (r QuoteRequest) Validate() error {
	if len(r.Items) == 0 {
		return invalid("items", CodeRequired, "at least one item is required")
	}
	if len(r.Items) > MaxQuoteItems {
		return invalid("items", CodeOutOfRange, fmt.Sprintf("must have at most %d items", MaxQuoteItems))
	}

	for index, item := range r.Items {
		if err := item.Validate(); err != nil {
			return prefixValidationError(fmt.Sprintf("items[%d]", index), err)
		}
	}

	switch r.Fulfillment {
	case FulfillmentDelivery:
		return validateRequiredString("zoneId", r.ZoneID, MaxIDLength)
	case FulfillmentPickup:
		return nil
	default:
		return invalid("fulfillment", CodeInvalid, "must be delivery or pickup")
	}
}

// Validate confere os limites de um item da cotação.
func (i QuoteItem) Validate() error {
	if err := validateRequiredString("productId", i.ProductID, MaxIDLength); err != nil {
		return err
	}
	if i.Quantity < 1 || i.Quantity > MaxQuoteQuantity {
		return invalid("quantity", CodeOutOfRange, "must be between 1 and 99")
	}
	if err := validateIDList("optionIds", i.OptionIDs); err != nil {
		return err
	}
	if err := validateIDList("addonIds", i.AddonIDs); err != nil {
		return err
	}
	return validateMaxLength("note", i.Note, MaxNoteLength)
}

// BuildQuote valida a disponibilidade no catálogo e calcula subtotal, taxa e
// total. É a única função de cálculo usada por quote e, no futuro, por order.
func BuildQuote(request QuoteRequest, catalog QuoteCatalog) (Quote, error) {
	if err := request.Validate(); err != nil {
		return Quote{}, err
	}
	if !catalog.Store.IsOpenAt(catalog.Now, catalog.Location) {
		return Quote{}, ErrStoreClosed
	}

	products := make(map[string]Product, len(catalog.Products))
	for _, product := range catalog.Products {
		if product.Active {
			products[product.ID] = product
		}
	}
	addons := make(map[string]Addon, len(catalog.Addons))
	for _, addon := range catalog.Addons {
		if addon.Active {
			addons[addon.ID] = addon
		}
	}

	var deliveryFeeCents int64
	var quoteZone *OrderZone
	if request.Fulfillment == FulfillmentDelivery {
		zone, err := findActiveZone(request.ZoneID, catalog.Zones)
		if err != nil {
			return Quote{}, err
		}
		deliveryFeeCents = zone.FeeCents
		quoteZone = &OrderZone{ID: zone.ID, Name: zone.Name, FeeCents: zone.FeeCents}
	}

	quote := Quote{
		Currency: CurrencyBRL,
		Zone:     quoteZone,
		Items:    make([]QuoteLine, 0, len(request.Items)),
		Warnings: []string{},
	}

	for itemIndex, item := range request.Items {
		product, ok := products[item.ProductID]
		if !ok || !product.Available {
			return Quote{}, ErrProductUnavailable
		}

		selectedOptions, err := validateSelectedOptions(item, product, itemIndex)
		if err != nil {
			return Quote{}, err
		}

		options := make([]QuoteOption, 0, len(item.OptionIDs)+len(item.AddonIDs))
		var optionsPriceCents int64
		for _, option := range selectedOptions {
			options = append(options, newQuoteOption(option, QuoteOptionKindOption))
			optionsPriceCents, err = addCents(optionsPriceCents, option.PriceCents)
			if err != nil {
				return Quote{}, err
			}
		}

		for addonIndex, addonID := range item.AddonIDs {
			addon, ok := addons[addonID]
			if !ok || !productHasAddon(product, addonID) {
				return Quote{}, &ValidationError{
					Field:   fmt.Sprintf("items[%d].addonIds[%d]", itemIndex, addonIndex),
					Code:    CodeInvalid,
					Message: "is not available for this product",
				}
			}
			options = append(options, newQuoteOption(addonToOption(addon), QuoteOptionKindAddon))
			optionsPriceCents, err = addCents(optionsPriceCents, addon.PriceCents)
			if err != nil {
				return Quote{}, err
			}
		}

		unitPriceCents, err := addCents(product.PriceCents, optionsPriceCents)
		if err != nil {
			return Quote{}, err
		}
		lineTotalCents, err := multiplyCents(unitPriceCents, item.Quantity)
		if err != nil {
			return Quote{}, err
		}
		quote.SubtotalCents, err = addCents(quote.SubtotalCents, lineTotalCents)
		if err != nil {
			return Quote{}, err
		}

		quote.Items = append(quote.Items, QuoteLine{
			ProductID:         product.ID,
			Name:              product.Name,
			Quantity:          item.Quantity,
			UnitPriceCents:    product.PriceCents,
			OptionsPriceCents: optionsPriceCents,
			LineTotalCents:    lineTotalCents,
			Options:           options,
		})
	}

	quote.DeliveryFeeCents = deliveryFeeCents
	totalCents, err := addCents(quote.SubtotalCents, deliveryFeeCents)
	if err != nil {
		return Quote{}, err
	}
	quote.TotalCents = totalCents
	return quote, nil
}

func validateIDList(field string, ids []string) error {
	seen := make(map[string]struct{}, len(ids))
	for index, id := range ids {
		itemField := fmt.Sprintf("%s[%d]", field, index)
		if err := validateRequiredString(itemField, id, MaxIDLength); err != nil {
			return err
		}
		if _, exists := seen[id]; exists {
			return invalid(itemField, CodeDuplicate, "duplicates another selection")
		}
		seen[id] = struct{}{}
	}
	return nil
}

func validateSelectedOptions(item QuoteItem, product Product, itemIndex int) ([]Option, error) {
	selected := make(map[string]struct{}, len(item.OptionIDs))
	byID := make(map[string]Option)
	for _, group := range product.OptionGroups {
		for _, option := range group.Options {
			option.GroupID = group.ID
			option.GroupName = group.Name
			byID[option.ID] = option
		}
	}

	for index, optionID := range item.OptionIDs {
		option, ok := byID[optionID]
		field := fmt.Sprintf("items[%d].optionIds[%d]", itemIndex, index)
		if !ok {
			return nil, &ValidationError{Field: field, Code: CodeInvalid, Message: "is not an option of this product"}
		}
		if !option.Available {
			return nil, &ValidationError{Field: field, Code: CodeInvalid, Message: "is not available"}
		}
		selected[optionID] = struct{}{}
	}

	for _, group := range product.OptionGroups {
		count := 0
		for _, option := range group.Options {
			if _, ok := selected[option.ID]; ok && option.Available {
				count++
			}
		}
		minimum := group.MinSelections
		if group.Required && minimum < 1 {
			minimum = 1
		}
		if count < minimum {
			return nil, &ValidationError{
				Field:   fmt.Sprintf("items[%d].optionIds", itemIndex),
				Code:    CodeRequired,
				Message: fmt.Sprintf("group %q requires at least %d option(s)", group.Name, minimum),
			}
		}
		if count > group.MaxSelections {
			return nil, &ValidationError{
				Field:   fmt.Sprintf("items[%d].optionIds", itemIndex),
				Code:    CodeOutOfRange,
				Message: fmt.Sprintf("group %q allows at most %d option(s)", group.Name, group.MaxSelections),
			}
		}
	}

	ordered := make([]Option, 0, len(selected))
	for _, group := range product.OptionGroups {
		for _, option := range group.Options {
			if _, ok := selected[option.ID]; ok {
				option.GroupID = group.ID
				option.GroupName = group.Name
				ordered = append(ordered, option)
			}
		}
	}
	return ordered, nil
}

func newQuoteOption(option Option, kind string) QuoteOption {
	groupID, groupName := option.GroupID, option.GroupName
	if kind == QuoteOptionKindAddon {
		groupID, groupName = "addons", "Adicionais"
	}
	return QuoteOption{
		GroupID:    groupID,
		GroupName:  groupName,
		OptionID:   option.ID,
		Name:       option.Name,
		Kind:       kind,
		PriceCents: option.PriceCents,
	}
}

func addonToOption(addon Addon) Option {
	return Option{ID: addon.ID, Name: addon.Name, PriceCents: addon.PriceCents}
}

func productHasAddon(product Product, addonID string) bool {
	for _, id := range product.AddonIDs {
		if id == addonID {
			return true
		}
	}
	return false
}

func findActiveZone(zoneID string, zones []DeliveryZone) (DeliveryZone, error) {
	for _, zone := range zones {
		if zone.ID == zoneID && zone.Active {
			return zone, nil
		}
	}
	return DeliveryZone{}, ErrZoneUnavailable
}

func addCents(left, right int64) (int64, error) {
	if right > 0 && left > math.MaxInt64-right {
		return 0, fmt.Errorf("%w: monetary overflow", ErrValidation)
	}
	if right < 0 && left < math.MinInt64-right {
		return 0, fmt.Errorf("%w: monetary overflow", ErrValidation)
	}
	return left + right, nil
}

func multiplyCents(value int64, quantity int) (int64, error) {
	if quantity <= 0 {
		return 0, fmt.Errorf("%w: invalid quantity", ErrValidation)
	}
	factor := int64(quantity)
	if value > math.MaxInt64/factor || value < math.MinInt64/factor {
		return 0, fmt.Errorf("%w: monetary overflow", ErrValidation)
	}
	return value * factor, nil
}
