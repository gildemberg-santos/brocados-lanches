package service

import (
	"context"
	"io"
	"log/slog"
	"time"

	"github.com/brocados-lanches/api/internal/domain"
)

// CustomerProfileStore é a superfície usada para reconhecer o cliente por
// telefone sem expor uma API pública de consulta.
type CustomerProfileStore interface {
	Upsert(ctx context.Context, profile domain.CustomerProfile) error
	OptOut(ctx context.Context, phoneKey string, at time.Time) error
	FindByPhoneKey(ctx context.Context, phoneKey string) (domain.CustomerProfile, bool, error)
}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
