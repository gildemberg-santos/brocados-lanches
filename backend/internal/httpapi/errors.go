package httpapi

import (
	"errors"
	"log/slog"
	"net/http"
	"regexp"

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
	codeUnauthorized       = "UNAUTHORIZED"
	codeForbidden          = "FORBIDDEN"
	codeRateLimited        = "RATE_LIMITED"
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

// malformedRequestError representa problema de sintaxe/transporte (400),
// distinto de regra semântica de negócio (422).
type malformedRequestError struct {
	field   string
	code    string
	message string
}

func (e *malformedRequestError) Error() string {
	return e.field + ": " + e.message
}

// writeError traduz erros de domínio para o envelope público. Erros internos
// são registrados com contexto e nunca expõem detalhe do banco ao cliente.
func writeError(w http.ResponseWriter, r *http.Request, logger *slog.Logger, err error) {
	status := http.StatusInternalServerError
	body := errorBody{
		Code:    codeInternalError,
		Message: "Não foi possível processar a solicitação.",
	}

	var malformed *malformedRequestError
	var validationError *domain.ValidationError
	switch {
	case errors.As(err, &malformed):
		status = http.StatusBadRequest
		body = errorBody{
			Code:    codeValidationError,
			Message: "Dados inválidos.",
			Details: []errorDetail{{
				Field:   malformed.field,
				Code:    malformed.code,
				Message: malformed.message,
			}},
		}
	case errors.As(err, &validationError):
		status = http.StatusUnprocessableEntity
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
		status = http.StatusUnprocessableEntity
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
	case errors.Is(err, domain.ErrUnauthorized):
		status = http.StatusUnauthorized
		body = errorBody{
			Code:    codeUnauthorized,
			Message: "Credenciais inválidas ou sessão expirada.",
		}
	case errors.Is(err, domain.ErrForbidden):
		status = http.StatusForbidden
		body = errorBody{
			Code:    codeForbidden,
			Message: "Acesso não autorizado.",
		}
	case errors.Is(err, domain.ErrRateLimited):
		status = http.StatusTooManyRequests
		body = errorBody{
			Code:    codeRateLimited,
			Message: "Muitas tentativas. Tente novamente em instantes.",
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
			"error", scrubForLog(err.Error()),
		)
		reporterFrom(r.Context()).Capture(
			r.Context(),
			errors.New(scrubForLog(err.Error())),
			requestTags(r),
		)
	}

	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, status, errorEnvelope{Error: body})
}

var (
	emailPattern  = regexp.MustCompile(`[\w.+-]+@[\w-]+\.[\w.-]+`)
	phonePattern  = regexp.MustCompile(`(?:\+?55)?\(?\d{2}\)?[\s.-]?\d{4,5}-?\d{4}`)
	cpfPattern    = regexp.MustCompile(`\d{3}\.?\d{3}\.?\d{3}-?\d{2}`)
	zipPattern    = regexp.MustCompile(`\d{5}-?\d{3}`)
	secretPattern = regexp.MustCompile(`(?i)(cookie|authorization|token)([:=]\s*)[^\s,;]+`)
)

// scrubForLog remove identificadores comuns antes de registrar erros internos.
// Nomes, endereços e observações não entram nas mensagens de erro do domínio;
// este filtro é uma defesa adicional para mensagens vindas de infraestrutura.
func scrubForLog(message string) string {
	result := emailPattern.ReplaceAllString(message, "[email]")
	result = phonePattern.ReplaceAllString(result, "[telefone]")
	result = cpfPattern.ReplaceAllString(result, "[cpf]")
	result = zipPattern.ReplaceAllString(result, "[cep]")
	return secretPattern.ReplaceAllString(result, "$1$2[redacted]")
}
