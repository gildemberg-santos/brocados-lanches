package httpapi

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/brocados-lanches/api/internal/domain"
)

// AdminWebhookService é a dependência dos handlers de integrações.
type AdminWebhookService interface {
	Get(ctx context.Context) (domain.WebhookSettings, error)
	Update(ctx context.Context, actorID string, patch domain.WebhookSettingsPatch) (domain.WebhookSettings, error)
}

// OrderNotifier publica eventos de pedido para integrações externas. A
// implementação de produção devolve na hora e envia em segundo plano.
type OrderNotifier interface {
	Notify(ctx context.Context, event domain.WebhookEvent)
}

type adminWebhookRequest struct {
	URL    *string `json:"url"`
	Active *bool   `json:"active"`
}

type adminWebhookResponse struct {
	URL       string `json:"url"`
	Active    bool   `json:"active"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

func handleAdminGetWebhook(service AdminWebhookService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if service == nil {
			writeUnconfigured(w, r, logger, "admin webhook service")
			return
		}
		settings, err := service.Get(r.Context())
		if err != nil {
			writeError(w, r, logger, err)
			return
		}
		writeJSON(w, http.StatusOK, adminItemEnvelope[adminWebhookResponse]{Data: newAdminWebhookResponse(settings)})
	}
}

func handleAdminUpdateWebhook(service AdminWebhookService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if service == nil {
			writeUnconfigured(w, r, logger, "admin webhook service")
			return
		}
		var request adminWebhookRequest
		if err := decodeAdminRequest(w, r, &request); err != nil {
			writeError(w, r, logger, err)
			return
		}
		actor, ok := adminFromContext(r.Context())
		if !ok {
			writeError(w, r, logger, domain.ErrUnauthorized)
			return
		}
		patch := domain.WebhookSettingsPatch{URL: request.URL, Active: request.Active}
		settings, err := service.Update(r.Context(), actor.ID, patch)
		if err != nil {
			writeError(w, r, logger, err)
			return
		}
		writeJSON(w, http.StatusOK, adminItemEnvelope[adminWebhookResponse]{Data: newAdminWebhookResponse(settings)})
	}
}

func newAdminWebhookResponse(settings domain.WebhookSettings) adminWebhookResponse {
	return adminWebhookResponse{
		URL:       settings.URL,
		Active:    settings.Active,
		CreatedAt: formatAdminTime(settings.CreatedAt),
		UpdatedAt: formatAdminTime(settings.UpdatedAt),
	}
}
