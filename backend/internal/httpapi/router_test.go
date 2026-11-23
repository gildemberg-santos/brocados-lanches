package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type fakePinger struct {
	err error
}

func (f fakePinger) Ping(context.Context) error {
	return f.err
}

func newTestRouter(pinger Pinger) http.Handler {
	return NewRouter(Dependencies{
		Pinger: pinger,
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	})
}

func TestHealthz(t *testing.T) {
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)

	newTestRouter(nil).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", response.Code)
	}
	if got := response.Header().Get("Content-Type"); got != "application/json; charset=utf-8" {
		t.Errorf("Content-Type = %q", got)
	}

	var body healthResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Status != "ok" {
		t.Errorf("status = %q, want ok", body.Status)
	}
}

func TestReadyz(t *testing.T) {
	tests := []struct {
		name       string
		pinger     Pinger
		wantStatus int
		wantHealth string
		wantMongo  string
	}{
		{
			name:       "ready",
			pinger:     fakePinger{},
			wantStatus: http.StatusOK,
			wantHealth: "ok",
			wantMongo:  "ok",
		},
		{
			name:       "mongo error",
			pinger:     fakePinger{err: errors.New("mongo down")},
			wantStatus: http.StatusServiceUnavailable,
			wantHealth: "degraded",
			wantMongo:  "error",
		},
		{
			name:       "nil pinger",
			pinger:     nil,
			wantStatus: http.StatusServiceUnavailable,
			wantHealth: "degraded",
			wantMongo:  "unavailable",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/readyz", nil)

			newTestRouter(test.pinger).ServeHTTP(response, request)

			if response.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, test.wantStatus)
			}

			var body readyResponse
			if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if body.Status != test.wantHealth {
				t.Errorf("status = %q, want %q", body.Status, test.wantHealth)
			}
			if body.Checks["mongo"] != test.wantMongo {
				t.Errorf("mongo = %q, want %q", body.Checks["mongo"], test.wantMongo)
			}
		})
	}
}

func TestAPIRoot(t *testing.T) {
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/", nil)

	newTestRouter(fakePinger{}).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", response.Code)
	}

	var body struct {
		Data struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Data.Version != "v1" {
		t.Errorf("version = %q, want v1", body.Data.Version)
	}
}

func TestUnknownRoute(t *testing.T) {
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/does-not-exist", nil)

	newTestRouter(fakePinger{}).ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", response.Code)
	}
}

func TestRequestLoggerCorrelatesAndHidesPII(t *testing.T) {
	var logs bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logs, nil))
	handler := NewRouter(Dependencies{
		Orders: &fakeOrderService{err: errors.New("mongo down para +5585999999999, dono@teste.local, CEP 61700-000, Cookie: segredo")},
		Logger: logger,
	})

	body := strings.Replace(validQuoteBody, "sem cebola", "Maria Silva da Rua das Flores", 1)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/orders/quote?nome=Maria%20Silva", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer segredo-do-token")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	requestID := response.Header().Get("X-Request-Id")
	if requestID == "" {
		t.Fatal("missing X-Request-Id response header")
	}
	if !strings.Contains(logs.String(), "request_id="+requestID) {
		t.Errorf("logs do not correlate with response: %s", logs.String())
	}

	for _, forbidden := range []string{
		"Maria Silva",
		"Rua das Flores",
		"segredo-do-token",
		"+5585999999999",
		"dono@teste.local",
		"61700-000",
		"segredo",
	} {
		if strings.Contains(logs.String(), forbidden) {
			t.Errorf("log leaked %q: %s", forbidden, logs.String())
		}
	}
	if !strings.Contains(logs.String(), "mongo down") {
		t.Errorf("sanitization removed useful internal context: %s", logs.String())
	}
}
