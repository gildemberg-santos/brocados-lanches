package httpapi

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/brocados-lanches/api/internal/service"
)

// StoreService é a dependência do handler público de status da loja.
type StoreService interface {
	Get(ctx context.Context) (service.StoreView, error)
}

type storeEnvelope struct {
	Data storeResponse `json:"data"`
}

type storeResponse struct {
	Name          string                      `json:"name"`
	Status        string                      `json:"status"`
	IsOpenManual  bool                        `json:"isOpenManual"`
	Timezone      string                      `json:"timezone"`
	DeliveryTime  string                      `json:"deliveryTime"`
	WhatsApp      string                      `json:"whatsapp"`
	Instagram     string                      `json:"instagram"`
	Address       string                      `json:"address"`
	OpeningHours  []adminOpeningHoursResponse `json:"openingHours"`
	NextOpeningAt *string                     `json:"nextOpeningAt"`
	UpdatedAt     string                      `json:"updatedAt"`
}

const storeTimezone = "America/Fortaleza"

func handleStore(store StoreService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if store == nil {
			writeError(w, r, logger, errors.New("store service is not configured"))
			return
		}

		view, err := store.Get(r.Context())
		if err != nil {
			writeError(w, r, logger, err)
			return
		}

		writeJSON(w, http.StatusOK, storeEnvelope{Data: newStoreResponse(view)})
	}
}

func newStoreResponse(view service.StoreView) storeResponse {
	settings := view.Settings
	days := make([]adminOpeningHoursResponse, 0, len(settings.OpeningHours))
	for _, day := range settings.OpeningHours {
		intervals := make([]adminOpeningIntervalResponse, 0, len(day.Intervals))
		for _, interval := range day.Intervals {
			intervals = append(intervals, adminOpeningIntervalResponse{Open: interval.Open, Close: interval.Close})
		}
		days = append(days, adminOpeningHoursResponse{Day: day.Day, Closed: day.Closed, Intervals: intervals})
	}

	var nextOpeningAt *string
	if view.NextOpeningAt != nil {
		formatted := view.NextOpeningAt.UTC().Format(time.RFC3339)
		nextOpeningAt = &formatted
	}

	return storeResponse{
		Name:          settings.Name,
		Status:        view.Status,
		IsOpenManual:  settings.IsOpen,
		Timezone:      storeTimezone,
		DeliveryTime:  settings.DeliveryTime,
		WhatsApp:      settings.WhatsApp,
		Instagram:     settings.Instagram,
		Address:       settings.Address,
		OpeningHours:  days,
		NextOpeningAt: nextOpeningAt,
		UpdatedAt:     formatAdminTime(settings.UpdatedAt),
	}
}
