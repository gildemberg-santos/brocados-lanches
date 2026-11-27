package httpapi

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/brocados-lanches/api/internal/domain"
	"github.com/brocados-lanches/api/internal/realtime"
	"github.com/brocados-lanches/api/internal/service"
	"github.com/go-chi/chi/v5"
)

// AdminOrderService é a dependência dos handlers de operação de pedidos.
type AdminOrderService interface {
	List(ctx context.Context, query domain.AdminOrderListQuery) (service.AdminPage[domain.AdminOrderListItem], error)
	Get(ctx context.Context, id string) (domain.Order, error)
	GetCustomerProfile(ctx context.Context, phone string) (*domain.CustomerProfile, error)
	UpdateStatus(ctx context.Context, actorID, id string, request domain.OrderStatusChangeRequest) (domain.Order, error)
}

type adminOrderListItemResponse struct {
	ID               string `json:"id"`
	OrderNumber      int64  `json:"orderNumber"`
	PublicID         string `json:"publicId"`
	Status           string `json:"status"`
	Fulfillment      string `json:"fulfillment"`
	CustomerName     string `json:"customerName"`
	Phone            string `json:"phone"`
	ItemCount        int    `json:"itemCount"`
	SubtotalCents    int64  `json:"subtotalCents"`
	DeliveryFeeCents int64  `json:"deliveryFeeCents"`
	TotalCents       int64  `json:"totalCents"`
	CreatedAt        string `json:"createdAt"`
	UpdatedAt        string `json:"updatedAt"`
}

type adminOrderDetailResponse struct {
	ID               string                           `json:"id"`
	OrderNumber      int64                            `json:"orderNumber"`
	PublicID         string                           `json:"publicId"`
	Status           string                           `json:"status"`
	Fulfillment      string                           `json:"fulfillment"`
	Customer         adminOrderCustomerResponse       `json:"customer"`
	CustomerProfile  *adminCustomerProfileResponse    `json:"customerProfile,omitempty"`
	Address          *adminOrderAddressResponse       `json:"address,omitempty"`
	Zone             *adminOrderZoneResponse          `json:"zone,omitempty"`
	Items            []adminOrderItemResponse         `json:"items"`
	SubtotalCents    int64                            `json:"subtotalCents"`
	DeliveryFeeCents int64                            `json:"deliveryFeeCents"`
	TotalCents       int64                            `json:"totalCents"`
	Payment          adminOrderPaymentResponse        `json:"payment"`
	Note             string                           `json:"note,omitempty"`
	StatusHistory    []adminOrderStatusChangeResponse `json:"statusHistory"`
	ConfirmedAt      string                           `json:"confirmedAt,omitempty"`
	CancelledAt      string                           `json:"cancelledAt,omitempty"`
	CancelReason     string                           `json:"cancelReason,omitempty"`
	CreatedAt        string                           `json:"createdAt"`
	UpdatedAt        string                           `json:"updatedAt"`
}

type adminOrderCustomerResponse struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
}

type adminCustomerProfileResponse struct {
	Name          string                     `json:"name"`
	Phone         string                     `json:"phone"`
	OrderCount    int                        `json:"orderCount"`
	LastOrderAt   string                     `json:"lastOrderAt,omitempty"`
	Address       *adminOrderAddressResponse `json:"address,omitempty"`
	Fulfillment   string                     `json:"fulfillment,omitempty"`
	ZoneName      string                     `json:"zoneName,omitempty"`
	PaymentMethod string                     `json:"paymentMethod,omitempty"`
	OptedOut      bool                       `json:"optedOut"`
	UpdatedAt     string                     `json:"updatedAt,omitempty"`
	RetainedUntil string                     `json:"retainedUntil,omitempty"`
}

type adminOrderAddressResponse struct {
	Street     string `json:"street"`
	Number     string `json:"number"`
	Complement string `json:"complement,omitempty"`
	District   string `json:"district"`
	City       string `json:"city"`
	State      string `json:"state"`
	ZipCode    string `json:"zipCode"`
	Reference  string `json:"reference,omitempty"`
}

type adminOrderZoneResponse struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	FeeCents int64  `json:"feeCents"`
}

type adminOrderItemResponse struct {
	ProductID      string                         `json:"productId"`
	Name           string                         `json:"name"`
	UnitPriceCents int64                          `json:"unitPriceCents"`
	Quantity       int                            `json:"quantity"`
	Options        []adminOrderItemOptionResponse `json:"options"`
	Note           string                         `json:"note,omitempty"`
	LineTotalCents int64                          `json:"lineTotalCents"`
}

type adminOrderItemOptionResponse struct {
	GroupID    string `json:"groupId,omitempty"`
	GroupName  string `json:"groupName,omitempty"`
	OptionID   string `json:"optionId"`
	Name       string `json:"name"`
	Kind       string `json:"kind"`
	PriceCents int64  `json:"priceCents"`
}

type adminOrderPaymentResponse struct {
	Method         string `json:"method"`
	ChangeForCents *int64 `json:"changeForCents,omitempty"`
}

type adminOrderStatusChangeResponse struct {
	Status  string `json:"status"`
	At      string `json:"at"`
	ActorID string `json:"actorId,omitempty"`
}

type adminOrderStatusRequest struct {
	Status string `json:"status"`
	Reason string `json:"reason,omitempty"`
}

func handleAdminListOrders(orders AdminOrderService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if orders == nil {
			writeUnconfigured(w, r, logger, "admin order service")
			return
		}
		query, err := parseAdminOrderListQuery(r)
		if err != nil {
			writeError(w, r, logger, err)
			return
		}
		page, err := orders.List(r.Context(), query)
		if err != nil {
			writeError(w, r, logger, err)
			return
		}
		items := make([]adminOrderListItemResponse, 0, len(page.Items))
		for _, item := range page.Items {
			items = append(items, newAdminOrderListItemResponse(item))
		}
		writeJSON(w, http.StatusOK, adminListEnvelope[adminOrderListItemResponse]{
			Data: items,
			Meta: newAdminMeta(page.Page, page.PageSize, page.Total, page.TotalPages),
		})
	}
}

func handleAdminGetOrder(orders AdminOrderService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if orders == nil {
			writeUnconfigured(w, r, logger, "admin order service")
			return
		}
		order, err := orders.Get(r.Context(), chi.URLParam(r, "id"))
		if err != nil {
			writeError(w, r, logger, err)
			return
		}

		detail := newAdminOrderDetailResponse(order)
		profile, profileErr := orders.GetCustomerProfile(r.Context(), order.Customer.Phone)
		if profileErr != nil {
			// Perfil é conveniência: o detalhe do pedido não pode falhar por ele.
			// Nenhum dado do cliente entra no log.
			logger.Warn("admin customer profile lookup failed", "reason", "lookup_failed")
		} else if profile != nil {
			detail.CustomerProfile = newAdminCustomerProfileResponse(*profile)
		}

		writeJSON(w, http.StatusOK, adminItemEnvelope[adminOrderDetailResponse]{
			Data: detail,
		})
	}
}

func handleAdminUpdateOrderStatus(
	orders AdminOrderService,
	hub *realtime.Hub,
	notifier OrderNotifier,
	logger *slog.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if orders == nil {
			writeUnconfigured(w, r, logger, "admin order service")
			return
		}
		var request adminOrderStatusRequest
		if err := decodeAdminRequest(w, r, &request); err != nil {
			writeError(w, r, logger, err)
			return
		}
		actor, ok := adminFromContext(r.Context())
		if !ok {
			writeError(w, r, logger, domain.ErrUnauthorized)
			return
		}
		statusRequest := domain.OrderStatusChangeRequest{Status: request.Status, Reason: request.Reason}
		order, err := orders.UpdateStatus(r.Context(), actor.ID, chi.URLParam(r, "id"), statusRequest)
		if err != nil {
			writeError(w, r, logger, err)
			return
		}
		publishOrderEvent(hub, realtime.EventOrderUpdated, order)
		if notifier != nil {
			notifier.Notify(r.Context(), domain.WebhookEvent{
				Type:  domain.WebhookEventOrderUpdated,
				Order: order,
			})
		}
		writeJSON(w, http.StatusOK, adminItemEnvelope[adminOrderDetailResponse]{
			Data: newAdminOrderDetailResponse(order),
		})
	}
}

func parseAdminOrderListQuery(r *http.Request) (domain.AdminOrderListQuery, error) {
	values := r.URL.Query()
	page, err := parseOptionalInt(values.Get("page"), "page")
	if err != nil {
		return domain.AdminOrderListQuery{}, err
	}
	pageSize, err := parseOptionalInt(values.Get("pageSize"), "pageSize")
	if err != nil {
		return domain.AdminOrderListQuery{}, err
	}
	from, err := parseOptionalAdminTime(values.Get("from"), "from")
	if err != nil {
		return domain.AdminOrderListQuery{}, err
	}
	to, err := parseOptionalAdminTime(values.Get("to"), "to")
	if err != nil {
		return domain.AdminOrderListQuery{}, err
	}
	return domain.AdminOrderListQuery{
		Page: page, PageSize: pageSize,
		Sort: values.Get("sort"), Order: values.Get("order"),
		Status: values.Get("status"), Fulfillment: domain.Fulfillment(values.Get("fulfillment")),
		From: from, To: to, Search: strings.TrimSpace(values.Get("q")),
	}, nil
}

func parseOptionalAdminTime(raw, field string) (*time.Time, error) {
	if raw == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return nil, &domain.ValidationError{Field: field, Code: domain.CodeInvalid, Message: "must be an RFC 3339 timestamp"}
	}
	utc := parsed.UTC()
	return &utc, nil
}

func newAdminOrderListItemResponse(item domain.AdminOrderListItem) adminOrderListItemResponse {
	return adminOrderListItemResponse{
		ID: item.ID, OrderNumber: item.Number, PublicID: item.PublicID,
		Status: item.Status, Fulfillment: string(item.Fulfillment),
		CustomerName: item.CustomerName, Phone: item.Phone,
		ItemCount:     item.ItemCount,
		SubtotalCents: item.SubtotalCents, DeliveryFeeCents: item.DeliveryFeeCents,
		TotalCents: item.TotalCents,
		CreatedAt:  formatAdminTime(item.CreatedAt), UpdatedAt: formatAdminTime(item.UpdatedAt),
	}
}

func newAdminOrderDetailResponse(order domain.Order) adminOrderDetailResponse {
	items := make([]adminOrderItemResponse, 0, len(order.Items))
	for _, item := range order.Items {
		options := make([]adminOrderItemOptionResponse, 0, len(item.Options))
		for _, option := range item.Options {
			options = append(options, adminOrderItemOptionResponse{
				GroupID: option.GroupID, GroupName: option.GroupName,
				OptionID: option.OptionID, Name: option.Name,
				Kind: option.Kind, PriceCents: option.PriceCents,
			})
		}
		items = append(items, adminOrderItemResponse{
			ProductID: item.ProductID, Name: item.Name,
			UnitPriceCents: item.UnitPriceCents, Quantity: item.Quantity,
			Options: options, Note: item.Note, LineTotalCents: item.LineTotalCents,
		})
	}
	history := make([]adminOrderStatusChangeResponse, 0, len(order.StatusHistory))
	for _, change := range order.StatusHistory {
		history = append(history, adminOrderStatusChangeResponse{
			Status: change.Status, At: formatAdminTime(change.At), ActorID: change.ActorID,
		})
	}

	response := adminOrderDetailResponse{
		ID: order.ID, OrderNumber: order.Number, PublicID: order.PublicID,
		Status: order.Status, Fulfillment: string(order.Fulfillment),
		Customer:      adminOrderCustomerResponse{Name: order.Customer.Name, Phone: order.Customer.Phone},
		Items:         items,
		SubtotalCents: order.SubtotalCents, DeliveryFeeCents: order.DeliveryFeeCents,
		TotalCents: order.TotalCents,
		Payment:    adminOrderPaymentResponse{Method: order.Payment.Method, ChangeForCents: order.Payment.ChangeForCents},
		Note:       order.Note, StatusHistory: history,
		ConfirmedAt:  formatOptionalAdminTime(order.ConfirmedAt),
		CancelledAt:  formatOptionalAdminTime(order.CancelledAt),
		CancelReason: order.CancelReason,
		CreatedAt:    formatAdminTime(order.CreatedAt), UpdatedAt: formatAdminTime(order.UpdatedAt),
	}
	if order.Address != nil {
		response.Address = &adminOrderAddressResponse{
			Street: order.Address.Street, Number: order.Address.Number,
			Complement: order.Address.Complement, District: order.Address.District,
			City: order.Address.City, State: order.Address.State,
			ZipCode: order.Address.ZipCode, Reference: order.Address.Reference,
		}
	}
	if order.Zone != nil {
		response.Zone = &adminOrderZoneResponse{
			ID: order.Zone.ID, Name: order.Zone.Name, FeeCents: order.Zone.FeeCents,
		}
	}
	return response
}

func newAdminCustomerProfileResponse(profile domain.CustomerProfile) *adminCustomerProfileResponse {
	response := &adminCustomerProfileResponse{
		Name:          profile.Name,
		Phone:         profile.Phone,
		OrderCount:    profile.OrderCount,
		LastOrderAt:   formatAdminTime(profile.LastOrderAt),
		Fulfillment:   string(profile.Fulfillment),
		PaymentMethod: profile.PaymentMethod,
		OptedOut:      profile.OptedOut,
		UpdatedAt:     formatAdminTime(profile.UpdatedAt),
		RetainedUntil: formatAdminTime(profile.ExpiresAt),
	}
	if profile.Zone != nil {
		response.ZoneName = profile.Zone.Name
	}
	if profile.Address != nil {
		response.Address = &adminOrderAddressResponse{
			Street:     profile.Address.Street,
			Number:     profile.Address.Number,
			Complement: profile.Address.Complement,
			District:   profile.Address.District,
			City:       profile.Address.City,
			State:      profile.Address.State,
			ZipCode:    profile.Address.ZipCode,
			Reference:  profile.Address.Reference,
		}
	}
	return response
}

func formatOptionalAdminTime(value *time.Time) string {
	if value == nil {
		return ""
	}
	return formatAdminTime(*value)
}
