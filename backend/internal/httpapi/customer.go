package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/brocados-lanches/api/internal/domain"
	"github.com/brocados-lanches/api/internal/platform/ratelimit"
)

// CustomerLookupService é a dependência pública de conveniência. O cliente
// dispara a consulta explicitamente ao informar o telefone no checkout.
type CustomerLookupService interface {
	Lookup(ctx context.Context, phone string) (*domain.CustomerProfile, error)
}

type customerLookupRequest struct {
	Phone string `json:"phone"`
}

type customerLookupEnvelope struct {
	Data customerLookupData `json:"data"`
}

type customerLookupData struct {
	Found   bool                 `json:"found"`
	Profile *customerProfileData `json:"profile,omitempty"`
}

type customerProfileData struct {
	Name          string               `json:"name,omitempty"`
	Phone         string               `json:"phone,omitempty"`
	Address       *customerAddressData `json:"address,omitempty"`
	Fulfillment   string               `json:"fulfillment,omitempty"`
	Zone          *customerZoneData    `json:"zone,omitempty"`
	PaymentMethod string               `json:"paymentMethod,omitempty"`
}

type customerAddressData struct {
	Street     string `json:"street"`
	Number     string `json:"number"`
	Complement string `json:"complement,omitempty"`
	District   string `json:"district"`
	City       string `json:"city"`
	State      string `json:"state"`
	ZipCode    string `json:"zipCode"`
	Reference  string `json:"reference,omitempty"`
}

// A taxa não é devolvida: zona e preço são sempre recalculados no servidor.
type customerZoneData struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func handleCustomerProfileLookup(
	customers CustomerLookupService,
	phoneLimiter *ratelimit.Limiter,
	logger *slog.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if customers == nil {
			writeError(w, r, logger, errors.New("customer service is not configured"))
			return
		}

		request, err := decodeCustomerLookupRequest(w, r)
		if err != nil {
			writeError(w, r, logger, err)
			return
		}
		if !allowCustomerPhone(phoneLimiter, request.Phone) {
			writeRateLimited(w, r)
			return
		}
		profile, err := customers.Lookup(r.Context(), request.Phone)
		if err != nil {
			writeError(w, r, logger, err)
			return
		}

		response := customerLookupEnvelope{Data: customerLookupData{Found: profile != nil}}
		if profile != nil {
			response.Data.Profile = newCustomerProfileData(*profile)
		}
		writeJSON(w, http.StatusOK, response)
	}
}

// decodeCustomerLookupRequest lê o corpo uma única vez e rejeita campos extras.
func decodeCustomerLookupRequest(w http.ResponseWriter, r *http.Request) (customerLookupRequest, error) {
	var request customerLookupRequest

	body := http.MaxBytesReader(w, r.Body, maxOrderBodyBytes)
	decoder := json.NewDecoder(body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		return customerLookupRequest{}, &malformedRequestError{
			field:   "body",
			code:    domain.CodeInvalid,
			message: "invalid JSON body",
		}
	}
	if len(request.Phone) > domain.MaxPhoneLength {
		return customerLookupRequest{}, &domain.ValidationError{
			Field:   "phone",
			Code:    domain.CodeTooLong,
			Message: "is too long",
		}
	}
	if !domain.IsValidBrazilianPhone(request.Phone) {
		return customerLookupRequest{}, &domain.ValidationError{
			Field:   "phone",
			Code:    domain.CodeInvalid,
			Message: "must be a valid Brazilian phone number",
		}
	}

	return request, nil
}

func newCustomerProfileData(profile domain.CustomerProfile) *customerProfileData {
	data := &customerProfileData{
		Name:          profile.Name,
		Phone:         profile.Phone,
		Fulfillment:   string(profile.Fulfillment),
		PaymentMethod: profile.PaymentMethod,
	}
	if profile.Address != nil {
		data.Address = &customerAddressData{
			Street:     profile.Address.Street,
			Number:     profile.Address.Number,
			Complement: profile.Address.Complement,
			District:   profile.Address.District,
			City:       profile.Address.City,
			State:      profile.Address.State,
			ZipCode:    profile.Address.ZipCode,
			Reference:  profile.Address.Reference,
		}
	}
	if profile.Zone != nil {
		data.Zone = &customerZoneData{
			ID:   profile.Zone.ID,
			Name: profile.Zone.Name,
		}
	}
	return data
}
