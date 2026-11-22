package httpapi

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/brocados-lanches/api/internal/domain"
	"github.com/go-chi/chi/v5/middleware"
)

// Códigos estáveis do envelope de erro da API.
const (
	codeValidationError    = "VALIDATION_ERROR"
	codeNotFound           = "NOT_FOUND"
	codeConflict           = "CONFLICT"
	codeStoreClosed        = "STORE_CLOSED"
	codeZoneUnavailable    = "ZONE_UNAVAILABLE"
	codeProductUnavailable = "PRODUCT_UNAVAILABLE"
	codeInternalError      = "INTERNAL_ERROR"
)

type errorEnvelope struct {
	Error errorBody `json:"error"`
}

type errorBody struct {
	Code    string        `json:"code"`
	Message string        `json:"message"`
	Details []errorDetail `json:"details,omitempty"`
}

type errorDetail struct {
	Field   string `json:"field"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// writeError traduz erros de domínio para o envelope público. Erros internos
// são registrados com contexto e nunca expõem detalhe do banco ao cliente.
func writeError(w http.ResponseWriter, r *http.Request, logger *slog.Logger, err error) {
	status := http.StatusInternalServerError
	body := errorBody{
		Code:    codeInternalError,
		Message: "Não foi possível processar a solicitação.",
	}

	var validationError *domain.ValidationError
	switch {
	case errors.As(err, &validationError):
		status = http.StatusBadRequest
		body = errorBody{
			Code:    codeValidationError,
			Message: "Dados inválidos.",
			Details: []errorDetail{{
				Field:   validationError.Field,
				Code:    validationError.Code,
				Message: validationError.Message,
			}},
		}
	case errors.Is(err, domain.ErrValidation):
		status = http.StatusBadRequest
		body = errorBody{
			Code:    codeValidationError,
			Message: "Dados inválidos.",
		}
	case errors.Is(err, domain.ErrStoreClosed):
		status = http.StatusConflict
		body = errorBody{
			Code:    codeStoreClosed,
			Message: "A loja está fechada no momento.",
		}
	case errors.Is(err, domain.ErrZoneUnavailable):
		status = http.StatusConflict
		body = errorBody{
			Code:    codeZoneUnavailable,
			Message: "A zona de entrega informada não está disponível.",
		}
	case errors.Is(err, domain.ErrProductUnavailable):
		status = http.StatusConflict
		body = errorBody{
			Code:    codeProductUnavailable,
			Message: "Um ou mais itens não estão disponíveis.",
		}
	case errors.Is(err, domain.ErrNotFound):
		status = http.StatusNotFound
		body = errorBody{
			Code:    codeNotFound,
			Message: "Recurso não encontrado.",
		}
	case errors.Is(err, domain.ErrConflict):
		status = http.StatusConflict
		body = errorBody{
			Code:    codeConflict,
			Message: "Conflito ao processar a solicitação.",
		}
	}

	if status >= http.StatusInternalServerError {
		if logger == nil {
			logger = slog.Default()
		}
		logger.Error("request failed",
			"request_id", middleware.GetReqID(r.Context()),
			"method", r.Method,
			"path", r.URL.Path,
			"error", err,
		)
	}

	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, status, errorEnvelope{Error: body})
}
