package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/brocados-lanches/api/internal/domain"
)

// StoreRepository é a leitura necessária para expor o status efetivo da loja.
type StoreRepository interface {
	GetStoreSettings(ctx context.Context) (domain.StoreSettings, error)
}

// StoreView é o dado público da loja: sem segredos e sem recálculo no cliente.
type StoreView struct {
	Settings      domain.StoreSettings
	Status        string
	NextOpeningAt *time.Time
}

// StoreService calcula o status efetivo no servidor.
type StoreService struct {
	repo     StoreRepository
	location *time.Location
	now      func() time.Time
}

// NewStoreService valida as dependências do serviço público da loja.
func NewStoreService(repo StoreRepository, location *time.Location) (*StoreService, error) {
	if repo == nil {
		return nil, errors.New("store repository is required")
	}
	if location == nil {
		return nil, errors.New("operation timezone is required")
	}
	return &StoreService{repo: repo, location: location, now: time.Now}, nil
}

// Get devolve configuração e status efetivo da loja no instante da chamada.
func (s *StoreService) Get(ctx context.Context) (StoreView, error) {
	settings, err := s.repo.GetStoreSettings(ctx)
	if err != nil {
		return StoreView{}, fmt.Errorf("get store settings: %w", err)
	}
	now := s.now().UTC()
	status := "closed"
	if settings.IsOpenAt(now, s.location) {
		status = "open"
	}
	return StoreView{
		Settings:      settings,
		Status:        status,
		NextOpeningAt: settings.NextOpeningAt(now, s.location),
	}, nil
}
