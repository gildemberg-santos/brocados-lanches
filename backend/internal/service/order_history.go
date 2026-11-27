package service

import (
	"context"
	"math"

	"github.com/brocados-lanches/api/internal/domain"
	"github.com/brocados-lanches/api/internal/platform/token"
)

const (
	defaultOrderHistoryPageSize = 20
	maxOrderHistoryPageSize     = 100
)

// ListByAccessToken devolve o histórico do navegador que apresentou o token.
// Nenhum dado pessoal é necessário para localizar os pedidos.
func (s *OrderService) ListByAccessToken(ctx context.Context, accessToken string, query domain.OrderHistoryQuery) (OrderHistoryPage, error) {
	if err := token.ValidateOrderAccess(accessToken); err != nil {
		return OrderHistoryPage{}, &domain.ValidationError{
			Field:   "accessToken",
			Code:    domain.CodeInvalid,
			Message: "invalid history token",
		}
	}
	normalized, err := normalizeOrderHistoryQuery(query)
	if err != nil {
		return OrderHistoryPage{}, err
	}

	orders, total, err := s.reader.ListByAccessToken(ctx, token.Hash(accessToken), normalized)
	if err != nil {
		return OrderHistoryPage{}, err
	}
	if orders == nil {
		orders = []domain.Order{}
	}
	totalPages := int64(0)
	if total > 0 {
		totalPages = int64(math.Ceil(float64(total) / float64(normalized.PageSize)))
	}
	return OrderHistoryPage{
		Items:      orders,
		Page:       normalized.Page,
		PageSize:   normalized.PageSize,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

func normalizeOrderHistoryQuery(query domain.OrderHistoryQuery) (domain.OrderHistoryQuery, error) {
	if query.Page == 0 {
		query.Page = 1
	}
	if query.Page < 1 {
		return query, &domain.ValidationError{Field: "page", Code: domain.CodeOutOfRange, Message: "must be at least 1"}
	}
	if query.PageSize == 0 {
		query.PageSize = defaultOrderHistoryPageSize
	}
	if query.PageSize < 1 || query.PageSize > maxOrderHistoryPageSize {
		return query, &domain.ValidationError{Field: "pageSize", Code: domain.CodeOutOfRange, Message: "must be between 1 and 100"}
	}
	return query, nil
}
