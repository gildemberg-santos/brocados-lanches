package httpapi

import (
	"log/slog"
	"net/http"
	"regexp"

	"github.com/getsentry/sentry-go"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

var (
	waLinkPattern       = regexp.MustCompile(`(?i)https?://(?:wa\.me|api\.whatsapp\.com)[^\s"'<>]*`)
	queryPattern        = regexp.MustCompile(`\?[^\s"'<>]*`)
	bearerPattern       = regexp.MustCompile(`(?i)(bearer\s+)[^\s,;]+`)
	sensitiveKeyPattern = regexp.MustCompile(`(?i)(phone|telefone|email|e-mail|address|endereco|cep|zip|token|authorization|cookie|idempotency|password|senha|whatsapp|walink|note|observacao|body|headers?)`)
)

// sentryScope cria um hub por requisição sem anexar headers, query ou corpo ao
// evento. O SDK não é inicializado quando o DSN está vazio; nesse caso o
// middleware não faz nada.
func sentryScope(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hub := sentry.CurrentHub()
		if hub == nil || hub.Client() == nil {
			next.ServeHTTP(w, r)
			return
		}

		scoped := hub.Clone()
		scoped.Scope().SetTag("request_id", middleware.GetReqID(r.Context()))
		next.ServeHTTP(w, r.WithContext(sentry.SetHubOnContext(r.Context(), scoped)))
	})
}

// recoverer substitui middleware.Recoverer: captura o painel no Sentry e não
// grava stack trace no log estruturado.
func recoverer(reporter Reporter, logger *slog.Logger) func(http.Handler) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				value := recover()
				if value == nil {
					return
				}
				if value == http.ErrAbortHandler {
					panic(value)
				}

				reporter.Recover(r.Context(), value, requestTags(r))
				logger.Error("http panic",
					"request_id", middleware.GetReqID(r.Context()),
					"method", r.Method,
					"path", r.URL.Path,
					"error", "panic recovered",
				)
				w.Header().Set("Cache-Control", "no-store")
				writeJSON(w, http.StatusInternalServerError, errorEnvelope{Error: errorBody{
					Code:    codeInternalError,
					Message: "Não foi possível processar a solicitação.",
				}})
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// requestTags devolve apenas identificadores não sensíveis. RoutePattern vem
// do template do Chi, nunca de r.URL.Path (que pode conter publicId).
func requestTags(r *http.Request) map[string]string {
	route := "unknown"
	if routeContext := chi.RouteContext(r.Context()); routeContext != nil {
		if pattern := routeContext.RoutePattern(); pattern != "" {
			route = pattern
		}
	}

	return map[string]string{
		"request_id": middleware.GetReqID(r.Context()),
		"method":     r.Method,
		"route":      route,
	}
}

// SanitizeSentryEvent é o BeforeSend do backend. Nenhum evento carrega user,
// request original, query, headers, corpo, telefone, e-mail, CEP ou token.
func SanitizeSentryEvent(event *sentry.Event) *sentry.Event {
	if event == nil {
		return nil
	}

	event.User = sentry.User{}
	event.Request = nil
	event.Message = sanitizeSentryText(event.Message)
	event.Transaction = sanitizeSentryText(event.Transaction)
	event.Logger = sanitizeSentryText(event.Logger)
	event.ServerName = sanitizeSentryText(event.ServerName)
	event.Tags = sanitizeStringMap(event.Tags)
	event.Modules = sanitizeStringMap(event.Modules)
	event.Fingerprint = sanitizeStrings(event.Fingerprint)
	event.Contexts = sanitizeContexts(event.Contexts)

	for _, breadcrumb := range event.Breadcrumbs {
		if breadcrumb == nil {
			continue
		}
		breadcrumb.Message = sanitizeSentryText(breadcrumb.Message)
		breadcrumb.Data = sanitizeMap(breadcrumb.Data)
	}

	for index := range event.Exception {
		exception := &event.Exception[index]
		exception.Type = sanitizeSentryText(exception.Type)
		exception.Value = sanitizeSentryText(exception.Value)
		if exception.Stacktrace != nil {
			sanitizeStacktrace(exception.Stacktrace)
		}
	}
	for index := range event.Threads {
		if event.Threads[index].Stacktrace != nil {
			sanitizeStacktrace(event.Threads[index].Stacktrace)
		}
	}

	return event
}

func sanitizeSentryText(value string) string {
	result := scrubForLog(value)
	result = waLinkPattern.ReplaceAllString(result, "[wa-link]")
	result = bearerPattern.ReplaceAllString(result, "$1[redacted]")
	result = queryPattern.ReplaceAllString(result, "?[redacted]")
	return result
}

func sanitizeStringMap(values map[string]string) map[string]string {
	if values == nil {
		return nil
	}
	result := make(map[string]string, len(values))
	for key, value := range values {
		if isSensitiveKey(key) {
			continue
		}
		result[key] = sanitizeSentryText(value)
	}
	return result
}

func sanitizeMap(values map[string]any) map[string]any {
	if values == nil {
		return nil
	}
	result := make(map[string]any, len(values))
	for key, value := range values {
		if isSensitiveKey(key) {
			continue
		}
		result[key] = sanitizeValue(value)
	}
	return result
}

func sanitizeContexts(contexts map[string]sentry.Context) map[string]sentry.Context {
	if contexts == nil {
		return nil
	}
	result := make(map[string]sentry.Context, len(contexts))
	for name, context := range contexts {
		result[name] = sanitizeMap(context)
	}
	return result
}

func sanitizeValue(value any) any {
	switch typed := value.(type) {
	case string:
		return sanitizeSentryText(typed)
	case map[string]any:
		return sanitizeMap(typed)
	case map[string]string:
		return sanitizeStringMap(typed)
	case []any:
		result := make([]any, len(typed))
		for index, item := range typed {
			result[index] = sanitizeValue(item)
		}
		return result
	case []string:
		result := make([]string, len(typed))
		for index, item := range typed {
			result[index] = sanitizeSentryText(item)
		}
		return result
	default:
		return value
	}
}

func sanitizeStacktrace(stacktrace *sentry.Stacktrace) {
	for index := range stacktrace.Frames {
		frame := &stacktrace.Frames[index]
		frame.Function = sanitizeSentryText(frame.Function)
		frame.Module = sanitizeSentryText(frame.Module)
		frame.Filename = sanitizeSentryText(frame.Filename)
		frame.AbsPath = sanitizeSentryText(frame.AbsPath)
		frame.ContextLine = sanitizeSentryText(frame.ContextLine)
		frame.PreContext = sanitizeStrings(frame.PreContext)
		frame.PostContext = sanitizeStrings(frame.PostContext)
		for key, value := range frame.Vars {
			frame.Vars[key] = sanitizeValue(value)
		}
	}
}

func sanitizeStrings(values []string) []string {
	if values == nil {
		return nil
	}
	result := make([]string, len(values))
	for index, value := range values {
		result[index] = sanitizeSentryText(value)
	}
	return result
}

func isSensitiveKey(key string) bool {
	return sensitiveKeyPattern.MatchString(key)
}
