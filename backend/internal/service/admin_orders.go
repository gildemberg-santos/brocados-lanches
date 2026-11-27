package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/brocados-lanches/api/internal/domain"
)

// AdminOrderRepository é a superfície da fila, do detalhe e da transição de
// status. O compare-and-set na escrita impede duas transições concorrentes.
type AdminOrderRepository interface {
	AdminListOrders(ctx context.Context, query domain.AdminOrderListQuery) ([]domain.Order, int64, error)
	FindByID(ctx context.Context, id string) (domain.Order, error)
	AdminUpdateOrderStatus(ctx context.Context, id, expectedStatus string, order domain.Order, audit domain.AuditLog) error
	AdminDeleteOrder(ctx context.Context, id string, audit domain.AuditLog) error
}

// AdminOrderService aplica validação, máscara de PII e auditoria à operação.
type AdminOrderService struct {
	repo     AdminOrderRepository
	profiles CustomerProfileStore
	now      func() time.Time
}

// AdminOrderServiceOption adiciona dependências opcionais ao serviço do painel.
type AdminOrderServiceOption func(*AdminOrderService)

// WithAdminCustomerProfiles permite que o detalhe do pedido mostre o perfil
// reconhecido pelo telefone, sem criar uma rota pública de consulta.
func WithAdminCustomerProfiles(store CustomerProfileStore) AdminOrderServiceOption {
	return func(service *AdminOrderService) {
		if store != nil {
			service.profiles = store
		}
	}
}

// NewAdminOrderService valida as dependências do serviço de operação.
func NewAdminOrderService(repo AdminOrderRepository, options ...AdminOrderServiceOption) (*AdminOrderService, error) {
	if repo == nil {
		return nil, errors.New("admin order repository is required")
	}
	service := &AdminOrderService{repo: repo, now: time.Now}
	for _, option := range options {
		option(service)
	}
	return service, nil
}

// List devolve a fila com PII mascarada e metadados de paginação.
func (s *AdminOrderService) List(ctx context.Context, query domain.AdminOrderListQuery) (AdminPage[domain.AdminOrderListItem], error) {
	normalized, err := normalizeAdminOrderQuery(query)
	if err != nil {
		return AdminPage[domain.AdminOrderListItem]{}, err
	}
	orders, total, err := s.repo.AdminListOrders(ctx, normalized)
	if err != nil {
		return AdminPage[domain.AdminOrderListItem]{}, fmt.Errorf("list admin orders: %w", err)
	}
	items := make([]domain.AdminOrderListItem, 0, len(orders))
	for _, order := range orders {
		items = append(items, domain.NewAdminOrderListItem(order))
	}
	return newAdminPage(items, domain.AdminListQuery{
		Page: normalized.Page, PageSize: normalized.PageSize,
	}, total), nil
}

// Get devolve o detalhe completo: o painel autenticado precisa dos dados de
// entrega para operar. O acesso é registrado pelo middleware de autenticação;
// nenhum dado do cliente vai para log.
func (s *AdminOrderService) Get(ctx context.Context, id string) (domain.Order, error) {
	if !isObjectID(id) {
		return domain.Order{}, domain.ErrNotFound
	}
	order, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return domain.Order{}, err
	}
	return order, nil
}

// GetCustomerProfile devolve o perfil reconhecido para o telefone do pedido.
// O painel já está autenticado; ainda assim nenhuma chave normalizada é
// devolvida ao cliente.
func (s *AdminOrderService) GetCustomerProfile(ctx context.Context, phone string) (*domain.CustomerProfile, error) {
	if s.profiles == nil {
		return nil, nil
	}
	phoneKey := domain.NormalizePhoneKey(phone)
	if phoneKey == "" {
		return nil, nil
	}
	profile, found, err := s.profiles.FindByPhoneKey(ctx, phoneKey)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, nil
	}
	return &profile, nil
}

// UpdateStatus valida a matriz de transições e grava status, histórico e
// auditoria de forma atômica quando o MongoDB suporta transação.
func (s *AdminOrderService) UpdateStatus(ctx context.Context, actorID, id string, request domain.OrderStatusChangeRequest) (domain.Order, error) {
	if !isObjectID(id) {
		return domain.Order{}, domain.ErrNotFound
	}
	if err := request.Validate(); err != nil {
		return domain.Order{}, err
	}

	current, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return domain.Order{}, err
	}
	now := s.now().UTC()
	updated, err := domain.ApplyOrderStatusChange(current, request, actorID, now)
	if err != nil {
		return domain.Order{}, err
	}

	audit := domain.AuditLog{
		ActorID:   actorID,
		Action:    domain.AuditActionUpdate,
		Entity:    domain.AuditEntityOrder,
		EntityID:  id,
		Before:    orderStatusAuditSnapshot(current),
		After:     orderStatusAuditSnapshot(updated),
		CreatedAt: now,
	}
	if err := s.repo.AdminUpdateOrderStatus(ctx, id, current.Status, updated, audit); err != nil {
		return domain.Order{}, err
	}
	return updated, nil
}

// Delete remove o pedido do banco e registra auditoria sem PII. A operação é
// permitida em qualquer status porque também serve para limpar pedidos de teste
// e lançamentos incorretos; a confirmação fica no painel, não no domínio.
func (s *AdminOrderService) Delete(ctx context.Context, actorID, id string) (domain.Order, error) {
	if !isObjectID(id) {
		return domain.Order{}, domain.ErrNotFound
	}
	order, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return domain.Order{}, err
	}
	audit := domain.AuditLog{
		ActorID:   actorID,
		Action:    domain.AuditActionDelete,
		Entity:    domain.AuditEntityOrder,
		EntityID:  id,
		Before:    orderDeleteAuditSnapshot(order),
		CreatedAt: s.now().UTC(),
	}
	if err := s.repo.AdminDeleteOrder(ctx, id, audit); err != nil {
		return domain.Order{}, err
	}
	return order, nil
}

func normalizeAdminOrderQuery(query domain.AdminOrderListQuery) (domain.AdminOrderListQuery, error) {
	if query.Page == 0 {
		query.Page = 1
	}
	if query.Page < 1 {
		return query, invalidField("page", domain.CodeOutOfRange, "must be at least 1")
	}
	if query.PageSize == 0 {
		query.PageSize = 20
	}
	if query.PageSize < 1 || query.PageSize > 100 {
		return query, invalidField("pageSize", domain.CodeOutOfRange, "must be between 1 and 100")
	}

	if query.Sort == "" {
		query.Sort = "createdAt"
	}
	permitted := false
	for _, candidate := range []string{"createdAt", "number", "status", "totalCents"} {
		if candidate == query.Sort {
			permitted = true
			break
		}
	}
	if !permitted {
		return query, invalidField("sort", domain.CodeInvalid, "unsupported sort field")
	}
	if query.Order == "" {
		query.Order = "desc"
	}
	if query.Order != "asc" && query.Order != "desc" {
		return query, invalidField("order", domain.CodeInvalid, "must be asc or desc")
	}

	if query.Status != "" && !domain.IsKnownOrderStatus(query.Status) {
		return query, invalidField("status", domain.CodeInvalid, "unsupported status")
	}
	if query.Fulfillment != "" && query.Fulfillment != domain.FulfillmentDelivery && query.Fulfillment != domain.FulfillmentPickup {
		return query, invalidField("fulfillment", domain.CodeInvalid, "must be delivery or pickup")
	}
	if query.From != nil && query.To != nil && query.From.After(*query.To) {
		return query, invalidField("from", domain.CodeInvalid, "must be before to")
	}
	if len(query.Search) > 120 {
		return query, invalidField("q", domain.CodeTooLong, "must have at most 120 characters")
	}
	return query, nil
}

func isObjectID(value string) bool {
	return validateObjectID("id", value) == nil
}

func orderStatusAuditSnapshot(order domain.Order) map[string]any {
	return map[string]any{
		"status":      order.Status,
		"fulfillment": string(order.Fulfillment),
		"totalCents":  order.TotalCents,
	}
}

func orderDeleteAuditSnapshot(order domain.Order) map[string]any {
	return map[string]any{
		"number":      order.Number,
		"status":      order.Status,
		"fulfillment": string(order.Fulfillment),
		"totalCents":  order.TotalCents,
	}
}
