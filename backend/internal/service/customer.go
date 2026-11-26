package service

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/brocados-lanches/api/internal/domain"
)

// CustomerProfileStore é a superfície usada para reconhecer o cliente por
// telefone. A consulta pública é explícita, limitada e devolve só conveniência.
type CustomerProfileStore interface {
	Upsert(ctx context.Context, profile domain.CustomerProfile) error
	OptOut(ctx context.Context, phoneKey string, at time.Time) error
	FindByPhoneKey(ctx context.Context, phoneKey string) (domain.CustomerProfile, bool, error)
}

// CustomerService expõe apenas a consulta de conveniência autorizada pelo
// cliente no checkout. Nunca devolve observação, troco ou histórico de pedidos.
type CustomerService struct {
	profiles CustomerProfileStore
}

// NewCustomerService valida a dependência do serviço de consulta.
func NewCustomerService(store CustomerProfileStore) (*CustomerService, error) {
	if store == nil {
		return nil, errors.New("customer profile store is required")
	}
	return &CustomerService{profiles: store}, nil
}

// Lookup devolve o perfil ativo do telefone informado. Um telefone sem perfil
// ou com opt-out devolve nil sem erro, para não revelar o tombstone.
func (s *CustomerService) Lookup(ctx context.Context, phone string) (*domain.CustomerProfile, error) {
	if len(phone) > domain.MaxPhoneLength {
		return nil, &domain.ValidationError{
			Field:   "phone",
			Code:    domain.CodeTooLong,
			Message: "is too long",
		}
	}
	phoneKey := domain.NormalizePhoneKey(phone)
	if len(phoneKey) < 12 || len(phoneKey) > 13 || !strings.HasPrefix(phoneKey, "55") {
		return nil, &domain.ValidationError{
			Field:   "phone",
			Code:    domain.CodeInvalid,
			Message: "must be a valid Brazilian phone number",
		}
	}

	profile, found, err := s.profiles.FindByPhoneKey(ctx, phoneKey)
	if err != nil {
		return nil, err
	}
	if !found || profile.OptedOut {
		return nil, nil
	}
	return &profile, nil
}

// discardLogger mantém compatibilidade com serviços que recebem logs opcionais.
func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
