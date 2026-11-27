package httpapi

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/brocados-lanches/api/internal/domain"
	"github.com/getsentry/sentry-go"
	"github.com/go-chi/chi/v5"
)

type capturedError struct {
	err  error
	tags map[string]string
}

type capturedRecovery struct {
	value any
	tags  map[string]string
}

type fakeReporter struct {
	errors     []capturedError
	recoveries []capturedRecovery
}

func (f *fakeReporter) Capture(_ context.Context, err error, tags map[string]string) {
	f.errors = append(f.errors, capturedError{err: err, tags: tags})
}

func (f *fakeReporter) Recover(_ context.Context, value any, tags map[string]string) {
	f.recoveries = append(f.recoveries, capturedRecovery{value: value, tags: tags})
}

func TestSanitizeSentryEventRemovesPII(t *testing.T) {
	event := &sentry.Event{
		Message:     "falha para +5585999999999 dono@teste.local CEP 61700-000 https://wa.me/5585999999999?text=oi?token=segredo",
		Transaction: "/api/v1/orders?telefone=5585999999999",
		User: sentry.User{
			ID:    "cliente-1",
			Email: "dono@teste.local",
			Name:  "Maria Silva",
		},
		Request: &sentry.Request{
			URL:     "https://brocados.local/api/v1/orders",
			Headers: map[string]string{"Authorization": "Bearer segredo"},
			Cookies: "sessao=abc",
			Data:    `{"customer":{"phone":"+5585999999999"}}`,
		},
		Tags: map[string]string{
			"request_id": "req-123",
			"phone":      "+5585999999999",
			"route":      "/api/v1/orders",
		},
		Modules: map[string]string{
			"app":   "v1.2.3",
			"token": "segredo",
		},
		Fingerprint: []string{"{{ default }}"},
		Contexts: map[string]sentry.Context{
			"runtime": {"version": "go1.27", "authorization": "Bearer segredo"},
		},
		Breadcrumbs: []*sentry.Breadcrumb{{
			Message: "request para +5585999999999",
			Data: map[string]any{
				"status": 500,
				"body":   `{"phone":"+5585999999999"}`,
			},
		}},
		Exception: []sentry.Exception{{
			Type:  "mongo",
			Value: "erro em dono@teste.local",
			Stacktrace: &sentry.Stacktrace{Frames: []sentry.Frame{{
				Function: "handle",
				Vars:     map[string]any{"phone": "+5585999999999"},
			}}},
		}},
	}

	result := SanitizeSentryEvent(event)
	if result != event {
		t.Fatal("BeforeSend must return the sanitized event")
	}
	if !result.User.IsEmpty() {
		t.Errorf("user was not cleared: %+v", result.User)
	}
	if result.Request != nil {
		t.Errorf("request was not cleared: %+v", result.Request)
	}

	rendered := sanitizedEventText(result)
	for _, forbidden := range []string{
		"+5585999999999", "5585999999999", "dono@teste.local", "Maria Silva",
		"61700-000", "wa.me", "Bearer segredo", "sessao=abc", "segredo",
		`{"phone":`, "telefone=5585",
	} {
		if strings.Contains(rendered, forbidden) {
			t.Errorf("event leaked %q: %s", forbidden, rendered)
		}
	}
	for _, wanted := range []string{"falha para", "go1.27", "req-123", "/api/v1/orders", "handle", "v1.2.3"} {
		if !strings.Contains(rendered, wanted) {
			t.Errorf("sanitization removed useful context %q: %s", wanted, rendered)
		}
	}
}

func sanitizedEventText(event *sentry.Event) string {
	var builder strings.Builder
	builder.WriteString(event.Message)
	builder.WriteString(event.Transaction)
	for key, value := range event.Tags {
		builder.WriteString(key)
		builder.WriteString("=")
		builder.WriteString(value)
	}
	for key, value := range event.Modules {
		builder.WriteString(key)
		builder.WriteString("=")
		builder.WriteString(value)
	}
	builder.WriteString(strings.Join(event.Fingerprint, " "))
	for _, context := range event.Contexts {
		for key, value := range context {
			builder.WriteString(key)
			builder.WriteString("=")
			if text, ok := value.(string); ok {
				builder.WriteString(text)
			}
		}
	}
	for _, breadcrumb := range event.Breadcrumbs {
		builder.WriteString(breadcrumb.Message)
		for key, value := range breadcrumb.Data {
			builder.WriteString(key)
			builder.WriteString("=")
			if text, ok := value.(string); ok {
				builder.WriteString(text)
			}
		}
	}
	for _, exception := range event.Exception {
		builder.WriteString(exception.Type)
		builder.WriteString(exception.Value)
		if exception.Stacktrace != nil {
			for _, frame := range exception.Stacktrace.Frames {
				builder.WriteString(frame.Function)
				for key, value := range frame.Vars {
					builder.WriteString(key)
					builder.WriteString("=")
					if text, ok := value.(string); ok {
						builder.WriteString(text)
					}
				}
			}
		}
	}
	return builder.String()
}

func TestWriteErrorCapturesOnlyUnexpectedErrors(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantEvents int
	}{
		{
			name:       "internal error",
			err:        errors.New("mongo indisponível"),
			wantStatus: http.StatusInternalServerError,
			wantEvents: 1,
		},
		{
			name:       "business rule",
			err:        domain.ErrValidation,
			wantStatus: http.StatusUnprocessableEntity,
			wantEvents: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			reporter := &fakeReporter{}
			orders := &fakeOrderService{err: test.err}
			handler := NewRouter(Dependencies{
				Orders:   orders,
				Reporter: reporter,
				Logger:   slog.New(slog.NewTextHandler(io.Discard, nil)),
			})

			request := httptest.NewRequest(http.MethodPost, "/api/v1/orders/quote", strings.NewReader(validQuoteBody))
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)

			if response.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, test.wantStatus)
			}
			if len(reporter.errors) != test.wantEvents {
				t.Fatalf("captured %d errors, want %d", len(reporter.errors), test.wantEvents)
			}
			if test.wantEvents == 0 {
				return
			}
			captured := reporter.errors[0]
			if captured.tags["route"] != "/api/v1/orders/quote" {
				t.Errorf("route tag = %q", captured.tags["route"])
			}
			if captured.tags["request_id"] == "" {
				t.Error("missing request_id tag")
			}
			if captured.tags["method"] != http.MethodPost {
				t.Errorf("method tag = %q", captured.tags["method"])
			}
			if !strings.Contains(captured.err.Error(), "mongo indisponível") {
				t.Errorf("captured error = %q", captured.err.Error())
			}
		})
	}
}

func TestRecovererCapturesPanicWithoutLoggingStack(t *testing.T) {
	var logs bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logs, nil))
	reporter := &fakeReporter{}

	router := chi.NewRouter()
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(withReporter(r.Context(), reporter)))
		})
	})
	router.Use(sentryScope)
	router.Use(recoverer(reporter, logger))
	router.Get("/boom", func(http.ResponseWriter, *http.Request) {
		panic("quebrou")
	})

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/boom", nil))

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", response.Code)
	}
	if !strings.Contains(response.Body.String(), codeInternalError) {
		t.Errorf("unexpected body: %s", response.Body.String())
	}
	if got := response.Header().Get("Cache-Control"); got != "no-store" {
		t.Errorf("Cache-Control = %q, want no-store", got)
	}
	if len(reporter.recoveries) != 1 {
		t.Fatalf("recovered %d panics, want 1", len(reporter.recoveries))
	}
	if reporter.recoveries[0].value != "quebrou" {
		t.Errorf("recovered value = %v", reporter.recoveries[0].value)
	}
	if !strings.Contains(logs.String(), "http panic") || strings.Contains(logs.String(), "goroutine") {
		t.Errorf("unexpected panic log: %s", logs.String())
	}
}

func TestRequestTagsUseRouteTemplate(t *testing.T) {
	router := chi.NewRouter()
	router.Get("/api/v1/orders/{publicId}", func(w http.ResponseWriter, r *http.Request) {
		tags := requestTags(r)
		if tags["route"] != "/api/v1/orders/{publicId}" {
			t.Errorf("route tag = %q", tags["route"])
		}
		if strings.Contains(tags["route"], "segredo-publico") {
			t.Errorf("route tag leaked identifier: %q", tags["route"])
		}
		writeJSON(w, http.StatusOK, map[string]string{"ok": "true"})
	})

	router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/api/v1/orders/segredo-publico", nil))
}
