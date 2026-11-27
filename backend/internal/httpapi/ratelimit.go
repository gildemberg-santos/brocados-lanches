package httpapi

import (
	"net/http"
	"strings"
	"time"

	"github.com/brocados-lanches/api/internal/domain"
	"github.com/brocados-lanches/api/internal/platform/ratelimit"
	"github.com/brocados-lanches/api/internal/platform/token"
)

// Limites do contrato (docs/06-contrato-api.md §15). A janela é fixa e curta
// para não reter chaves identificáveis por mais tempo que o necessário.
const (
	rateLimitWindow       = time.Minute
	rateLimitRetryAfter   = "60"
	orderPerIPLimit       = 10
	orderPerPhoneLimit    = 3
	customerPerIPLimit    = 10
	customerPerPhoneLimit = 3
	quotePerIPLimit       = 30
	trackingPerIPLimit    = 60
	historyPerTokenLimit  = 30
	menuPerIPLimit        = 120
	adminPerSessionLimit  = 120
)

// Limiters reúne os limitadores das rotas públicas e do painel.
type Limiters struct {
	Orders          *ratelimit.Limiter
	OrderPhones     *ratelimit.Limiter
	CustomerLookups *ratelimit.Limiter
	CustomerPhones  *ratelimit.Limiter
	Quotes          *ratelimit.Limiter
	Tracking        *ratelimit.Limiter
	HistoryTokens   *ratelimit.Limiter
	Menu            *ratelimit.Limiter
	Admin           *ratelimit.Limiter
}

// DefaultLimiters cria limitadores com os valores do contrato.
func DefaultLimiters() Limiters {
	return Limiters{
		Orders:          newLimiter(orderPerIPLimit),
		OrderPhones:     newLimiter(orderPerPhoneLimit),
		CustomerLookups: newLimiter(customerPerIPLimit),
		CustomerPhones:  newLimiter(customerPerPhoneLimit),
		Quotes:          newLimiter(quotePerIPLimit),
		Tracking:        newLimiter(trackingPerIPLimit),
		HistoryTokens:   newLimiter(historyPerTokenLimit),
		Menu:            newLimiter(menuPerIPLimit),
		Admin:           newLimiter(adminPerSessionLimit),
	}
}

func (l *Limiters) fillDefaults() {
	if l.Orders == nil {
		l.Orders = newLimiter(orderPerIPLimit)
	}
	if l.OrderPhones == nil {
		l.OrderPhones = newLimiter(orderPerPhoneLimit)
	}
	if l.CustomerLookups == nil {
		l.CustomerLookups = newLimiter(customerPerIPLimit)
	}
	if l.CustomerPhones == nil {
		l.CustomerPhones = newLimiter(customerPerPhoneLimit)
	}
	if l.Quotes == nil {
		l.Quotes = newLimiter(quotePerIPLimit)
	}
	if l.Tracking == nil {
		l.Tracking = newLimiter(trackingPerIPLimit)
	}
	if l.HistoryTokens == nil {
		l.HistoryTokens = newLimiter(historyPerTokenLimit)
	}
	if l.Menu == nil {
		l.Menu = newLimiter(menuPerIPLimit)
	}
	if l.Admin == nil {
		l.Admin = newLimiter(adminPerSessionLimit)
	}
}

func newLimiter(limit int) *ratelimit.Limiter {
	limiter, err := ratelimit.New(limit, rateLimitWindow)
	if err != nil {
		// Os limites são constantes válidas; erro indica programação, não configuração.
		panic(err)
	}
	return limiter
}

// limitByIP aplica o limite por IP do cliente. X-Forwarded-For não é aceito:
// o middleware ClientIPFromRemoteAddr já remove essa fonte não confiável.
func limitByIP(limiter *ratelimit.Limiter, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if limiter != nil && !limiter.Allow("ip|"+clientIP(r)) {
			writeRateLimited(w, r)
			return
		}
		next(w, r)
	}
}

// limitAdminSession conta por hash da sessão; o token bruto nunca fica em memória.
func limitAdminSession(limiter *ratelimit.Limiter, cfg AdminCookieConfig, next func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if limiter != nil {
				cookie, err := r.Cookie(cfg.Name)
				if err == nil && cookie.Value != "" && !limiter.Allow("session|"+token.Hash(cookie.Value)) {
					writeRateLimited(w, r)
					return
				}
			}
			next(handler).ServeHTTP(w, r)
		})
	}
}

// allowOrderPhone é chamado depois de decodificar o pedido, quando o telefone
// normalized está disponível. O telefone entra no limite apenas como hash.
func allowOrderPhone(limiter *ratelimit.Limiter, phone string) bool {
	if limiter == nil || phone == "" {
		return true
	}
	return limiter.Allow("phone|" + token.Hash(strings.ToLower(strings.TrimSpace(phone))))
}

// allowOrderToken limita o histórico por token sem guardar o token em claro.
func allowOrderToken(limiter *ratelimit.Limiter, accessToken string) bool {
	if limiter == nil || accessToken == "" {
		return true
	}
	return limiter.Allow("history|" + token.Hash(accessToken))
}

// allowCustomerPhone usa a chave normalizada apenas como hash; o telefone
// informado nunca fica na memória do limitador nem em log.
func allowCustomerPhone(limiter *ratelimit.Limiter, phone string) bool {
	if limiter == nil || phone == "" {
		return true
	}
	return limiter.Allow("customer|" + token.Hash(domain.NormalizePhoneKey(phone)))
}

func writeRateLimited(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Retry-After", rateLimitRetryAfter)
	writeError(w, r, nil, domain.ErrRateLimited)
}
