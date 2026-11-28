package httpapi

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/brocados-lanches/api/internal/domain"
)

// AdminPixService é a dependência dos handlers de configuração do Pix.
type AdminPixService interface {
	Get(ctx context.Context) (domain.PixSettings, error)
	Update(ctx context.Context, actorID string, patch domain.PixSettingsPatch) (domain.PixSettings, error)
}

type adminPixRequest struct {
	Active       *bool   `json:"active"`
	Key          *string `json:"key"`
	MerchantName *string `json:"merchantName"`
	MerchantCity *string `json:"merchantCity"`
}

type adminPixResponse struct {
	Active       bool   `json:"active"`
	Key          string `json:"key"`
	MerchantName string `json:"merchantName"`
	MerchantCity string `json:"merchantCity"`
	CreatedAt    string `json:"createdAt"`
	UpdatedAt    string `json:"updatedAt"`
}

func handleAdminGetPix(service AdminPixService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if service == nil {
			writeUnconfigured(w, r, logger, "admin pix service")
			return
		}
		settings, err := service.Get(r.Context())
		if err != nil {
			writeError(w, r, logger, err)
			return
		}
		writeJSON(w, http.StatusOK, adminItemEnvelope[adminPixResponse]{Data: newAdminPixResponse(settings)})
	}
}

func handleAdminUpdatePix(service AdminPixService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if service == nil {
			writeUnconfigured(w, r, logger, "admin pix service")
			return
		}
		var request adminPixRequest
		if err := decodeAdminRequest(w, r, &request); err != nil {
			writeError(w, r, logger, err)
			return
		}
		actor, ok := adminFromContext(r.Context())
		if !ok {
			writeError(w, r, logger, domain.ErrUnauthorized)
			return
		}
		patch := domain.PixSettingsPatch{
			Active: request.Active, Key: request.Key,
			MerchantName: request.MerchantName, MerchantCity: request.MerchantCity,
		}
		settings, err := service.Update(r.Context(), actor.ID, patch)
		if err != nil {
			writeError(w, r, logger, err)
			return
		}
		writeJSON(w, http.StatusOK, adminItemEnvelope[adminPixResponse]{Data: newAdminPixResponse(settings)})
	}
}

func newAdminPixResponse(settings domain.PixSettings) adminPixResponse {
	return adminPixResponse{
		Active:       settings.Active,
		Key:          settings.Key,
		MerchantName: settings.MerchantName,
		MerchantCity: settings.MerchantCity,
		CreatedAt:    formatAdminTime(settings.CreatedAt),
		UpdatedAt:    formatAdminTime(settings.UpdatedAt),
	}
}
