package httpapi

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/brocados-lanches/api/internal/domain"
)

// AdminStoreService é a dependência dos handlers de configuração da loja.
type AdminStoreService interface {
	Get(ctx context.Context) (domain.StoreSettings, error)
	Update(ctx context.Context, actorID string, patch domain.StoreSettingsPatch) (domain.StoreSettings, error)
}

type adminStoreRequest struct {
	Name         *string                   `json:"name"`
	IsOpen       *bool                     `json:"isOpen"`
	OpeningHours *[]adminOpeningHoursInput `json:"openingHours"`
	DeliveryTime *string                   `json:"deliveryTime"`
	WhatsApp     *string                   `json:"whatsapp"`
	Instagram    *string                   `json:"instagram"`
	Address      *string                   `json:"address"`
}

type adminOpeningHoursInput struct {
	Day       string                      `json:"day"`
	Closed    bool                        `json:"closed"`
	Intervals []adminOpeningIntervalInput `json:"intervals"`
}

type adminOpeningIntervalInput struct {
	Open  string `json:"open"`
	Close string `json:"close"`
}

type adminStoreResponse struct {
	Name         string                      `json:"name"`
	IsOpen       bool                        `json:"isOpen"`
	OpeningHours []adminOpeningHoursResponse `json:"openingHours"`
	DeliveryTime string                      `json:"deliveryTime"`
	WhatsApp     string                      `json:"whatsapp"`
	Instagram    string                      `json:"instagram"`
	Address      string                      `json:"address"`
	CreatedAt    string                      `json:"createdAt"`
	UpdatedAt    string                      `json:"updatedAt"`
}

type adminOpeningHoursResponse struct {
	Day       string                         `json:"day"`
	Closed    bool                           `json:"closed"`
	Intervals []adminOpeningIntervalResponse `json:"intervals"`
}

type adminOpeningIntervalResponse struct {
	Open  string `json:"open"`
	Close string `json:"close"`
}

func handleAdminGetStore(service AdminStoreService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if service == nil {
			writeUnconfigured(w, r, logger, "admin store service")
			return
		}
		settings, err := service.Get(r.Context())
		if err != nil {
			writeError(w, r, logger, err)
			return
		}
		writeJSON(w, http.StatusOK, adminItemEnvelope[adminStoreResponse]{Data: newAdminStoreResponse(settings)})
	}
}

func handleAdminUpdateStore(service AdminStoreService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if service == nil {
			writeUnconfigured(w, r, logger, "admin store service")
			return
		}
		var request adminStoreRequest
		if err := decodeAdminRequest(w, r, &request); err != nil {
			writeError(w, r, logger, err)
			return
		}
		actor, ok := adminFromContext(r.Context())
		if !ok {
			writeError(w, r, logger, domain.ErrUnauthorized)
			return
		}
		settings, err := service.Update(r.Context(), actor.ID, request.patch())
		if err != nil {
			writeError(w, r, logger, err)
			return
		}
		writeJSON(w, http.StatusOK, adminItemEnvelope[adminStoreResponse]{Data: newAdminStoreResponse(settings)})
	}
}

func (r adminStoreRequest) patch() domain.StoreSettingsPatch {
	var hours *[]domain.OpeningHours
	if r.OpeningHours != nil {
		converted := make([]domain.OpeningHours, 0, len(*r.OpeningHours))
		for _, day := range *r.OpeningHours {
			intervals := make([]domain.OpeningInterval, 0, len(day.Intervals))
			for _, interval := range day.Intervals {
				intervals = append(intervals, domain.OpeningInterval{Open: interval.Open, Close: interval.Close})
			}
			converted = append(converted, domain.OpeningHours{Day: day.Day, Closed: day.Closed, Intervals: intervals})
		}
		hours = &converted
	}
	return domain.StoreSettingsPatch{
		Name: r.Name, IsOpen: r.IsOpen, OpeningHours: hours,
		DeliveryTime: r.DeliveryTime, WhatsApp: r.WhatsApp,
		Instagram: r.Instagram, Address: r.Address,
	}
}

func newAdminStoreResponse(settings domain.StoreSettings) adminStoreResponse {
	days := make([]adminOpeningHoursResponse, 0, len(settings.OpeningHours))
	for _, day := range settings.OpeningHours {
		intervals := make([]adminOpeningIntervalResponse, 0, len(day.Intervals))
		for _, interval := range day.Intervals {
			intervals = append(intervals, adminOpeningIntervalResponse{Open: interval.Open, Close: interval.Close})
		}
		days = append(days, adminOpeningHoursResponse{Day: day.Day, Closed: day.Closed, Intervals: intervals})
	}
	return adminStoreResponse{
		Name: settings.Name, IsOpen: settings.IsOpen, OpeningHours: days,
		DeliveryTime: settings.DeliveryTime, WhatsApp: settings.WhatsApp,
		Instagram: settings.Instagram, Address: settings.Address,
		CreatedAt: formatAdminTime(settings.CreatedAt), UpdatedAt: formatAdminTime(settings.UpdatedAt),
	}
}
