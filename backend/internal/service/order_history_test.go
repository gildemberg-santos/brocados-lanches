package service

import (
	"context"
	"errors"
	"testing"

	"github.com/brocados-lanches/api/internal/domain"
	"github.com/brocados-lanches/api/internal/platform/token"
)

func TestOrderServiceListByAccessToken(t *testing.T) {
	reader, service := orderServiceFixture()
	accessToken := token.DeriveOrderAccess("test-order-access-secret", "chave-estavel")
	reader.history = []domain.Order{{Number: 1042}, {Number: 1041}}
	reader.historyTotal = 5

	page, err := service.ListByAccessToken(context.Background(), accessToken, domain.OrderHistoryQuery{Page: 2, PageSize: 2})
	if err != nil {
		t.Fatalf("ListByAccessToken() error = %v", err)
	}
	if reader.historyHash != token.Hash(accessToken) {
		t.Error("repository received a hash different from the token")
	}
	if reader.historyQuery.Page != 2 || reader.historyQuery.PageSize != 2 {
		t.Errorf("query = %+v", reader.historyQuery)
	}
	if len(page.Items) != 2 || page.Total != 5 || page.TotalPages != 3 {
		t.Errorf("page = %+v", page)
	}
}

func TestOrderServiceListByAccessTokenDefaultsPagination(t *testing.T) {
	_, service := orderServiceFixture()
	accessToken := token.DeriveOrderAccess("test-order-access-secret", "chave-estavel")

	page, err := service.ListByAccessToken(context.Background(), accessToken, domain.OrderHistoryQuery{})
	if err != nil {
		t.Fatalf("ListByAccessToken() error = %v", err)
	}
	if page.Page != 1 || page.PageSize != 20 {
		t.Errorf("page = %+v, want defaults 1/20", page)
	}
	if len(page.Items) != 0 {
		t.Error("empty history must return an empty slice, not nil")
	}
}

func TestOrderServiceListByAccessTokenValidatesInput(t *testing.T) {
	_, service := orderServiceFixture()
	accessToken := token.DeriveOrderAccess("test-order-access-secret", "chave-estavel")
	tests := []struct {
		name  string
		token string
		query domain.OrderHistoryQuery
		field string
	}{
		{name: "invalid token", token: "curto", field: "accessToken"},
		{name: "invalid page", token: accessToken, query: domain.OrderHistoryQuery{Page: -1}, field: "page"},
		{name: "invalid page size", token: accessToken, query: domain.OrderHistoryQuery{PageSize: 101}, field: "pageSize"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := service.ListByAccessToken(context.Background(), test.token, test.query)
			var validationError *domain.ValidationError
			if !errors.As(err, &validationError) || validationError.Field != test.field {
				t.Fatalf("error = %v, want validation error on %s", err, test.field)
			}
		})
	}
}
