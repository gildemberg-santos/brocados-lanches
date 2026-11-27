package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

const catalogFixture = `{
  "data": {
    "categories": [{
      "products": [{
        "id": "produto-1",
        "available": true,
        "optionGroups": [{
          "id": "grupo-1",
          "required": true,
          "minSelections": 1,
          "maxSelections": 1,
          "options": [
            {"id": "indisponivel", "available": false},
            {"id": "opcao-1", "available": true}
          ]
        }]
      }]
    }]
  }
}`

const zonesFixture = `{"data":[{"id":"zona-1","active":true}]}`

func writeFixture(w http.ResponseWriter, body string) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = io.WriteString(w, body)
}

func TestRequiredOptionsRespeitaDisponibilidadeENumeroMinimo(t *testing.T) {
	var menu catalogMenu
	if err := json.Unmarshal([]byte(catalogFixture), &menu); err != nil {
		t.Fatal(err)
	}
	groups := menu.Data.Categories[0].Products[0].OptionGroups

	selected, ok := requiredOptions(groups)
	if !ok || len(selected) != 1 || selected[0] != "opcao-1" {
		t.Fatalf("requiredOptions() = %v, %v", selected, ok)
	}

	groups[0].MinSelections = 2
	if _, ok := requiredOptions(groups); ok {
		t.Fatal("requiredOptions aceitou grupo sem opções suficientes")
	}
}

func TestDiscoverPayloadUsaCatalogoESalvaPerfilDesligado(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/menu":
			writeFixture(w, catalogFixture)
		case "/api/v1/delivery-zones":
			writeFixture(w, zonesFixture)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	payload, err := discoverPayload(context.Background(), server.Client(), server.URL)
	if err != nil {
		t.Fatal(err)
	}
	var request struct {
		Items []struct {
			ProductID string   `json:"productId"`
			OptionIDs []string `json:"optionIds"`
		} `json:"items"`
		Fulfillment         string `json:"fulfillment"`
		ZoneID              string `json:"zoneId"`
		SaveCustomerProfile bool   `json:"saveCustomerProfile"`
	}
	if err := json.Unmarshal(payload, &request); err != nil {
		t.Fatal(err)
	}
	if len(request.Items) != 1 || request.Items[0].ProductID != "produto-1" ||
		len(request.Items[0].OptionIDs) != 1 || request.Items[0].OptionIDs[0] != "opcao-1" {
		t.Fatalf("payload inesperado: %s", payload)
	}
	if request.Fulfillment != "delivery" || request.ZoneID != "zona-1" || request.SaveCustomerProfile {
		t.Fatalf("payload inesperado: %s", payload)
	}
}

func TestDiscoverPayloadFalhaSemZonaAtiva(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/menu":
			writeFixture(w, catalogFixture)
		case "/api/v1/delivery-zones":
			writeFixture(w, `{"data":[{"id":"zona-inativa","active":false}]}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	if _, err := discoverPayload(context.Background(), server.Client(), server.URL); err == nil {
		t.Fatal("discoverPayload deveria falhar sem zona ativa")
	}
}

func TestPercentile(t *testing.T) {
	summary := newStats()
	for index, milliseconds := range []int{50, 10, 40, 30, 20} {
		summary.add(result{
			status:  http.StatusOK,
			elapsed: time.Duration(milliseconds) * time.Millisecond,
		})
		if index == 4 {
			summary.add(result{err: fmt.Errorf("transporte falhou")})
		}
	}
	if got := summary.percentile(0); got != 10*time.Millisecond {
		t.Fatalf("p0 = %s", got)
	}
	if got := summary.percentile(0.5); got != 30*time.Millisecond {
		t.Fatalf("p50 = %s", got)
	}
	if got := summary.percentile(1); got != 50*time.Millisecond {
		t.Fatalf("p100 = %s", got)
	}
	if summary.errors() != 1 {
		t.Fatalf("erros = %d", summary.errors())
	}
	if summary.statusSummary() != "[200=5]" {
		t.Fatalf("status = %s", summary.statusSummary())
	}
}

func TestExerciseExecutaTodasAsRequisicoesComChavesDaFabrica(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("Idempotency-Key")
		if key == "" {
			http.Error(w, "sem chave", http.StatusBadRequest)
			return
		}
		writeFixture(w, `{"data":{"publicId":"abc"}}`)
	}))
	defer server.Close()

	summary, err := exercise(
		context.Background(),
		server.Client(),
		http.MethodPost,
		server.URL,
		8,
		func(int) ([]byte, error) { return []byte(`{}`), nil },
		func(index int) string { return fmt.Sprintf("chave-%d", index) },
	)
	if err != nil {
		t.Fatal(err)
	}
	if summary.total != 8 || summary.success != 8 {
		t.Fatalf("total=%d sucesso=%d", summary.total, summary.success)
	}
}

func TestRunGeraCargaESeparaChavesDuplicadas(t *testing.T) {
	var mutex sync.Mutex
	keys := make(map[string]int)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/menu":
			writeFixture(w, catalogFixture)
		case "/api/v1/delivery-zones":
			writeFixture(w, zonesFixture)
		case "/api/v1/orders":
			key := r.Header.Get("Idempotency-Key")
			if key == "" {
				http.Error(w, "sem chave", http.StatusBadRequest)
				return
			}
			mutex.Lock()
			keys[key]++
			mutex.Unlock()
			w.WriteHeader(http.StatusCreated)
			writeFixture(w, `{"data":{"publicId":"abc"}}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	var output bytes.Buffer
	cfg := config{
		baseURL:           server.URL + "/",
		menuConcurrency:   5,
		ordersConcurrency: 4,
		duplicates:        2,
		timeout:           5 * time.Second,
	}
	if err := run(context.Background(), cfg, &output); err != nil {
		t.Fatal(err)
	}
	report := output.String()
	for _, expected := range []string{
		"GET /menu: total=5 sucesso=5 erros=0",
		"POST /orders: total=6 sucesso=6 erros=0",
		"p50=", "p95=", "p99=",
	} {
		if !strings.Contains(report, expected) {
			t.Fatalf("relatório %q não contém %q", report, expected)
		}
	}

	mutex.Lock()
	defer mutex.Unlock()
	if len(keys) != 5 {
		t.Fatalf("chaves distintas = %d, esperado 5 (4 únicas + 1 compartilhada): %v", len(keys), keys)
	}
	for key, count := range keys {
		if strings.HasSuffix(key, "duplicata") && count != 2 {
			t.Fatalf("chave duplicada %s usada %d vezes", key, count)
		}
	}
}

func TestRunUsaPayloadFornecidoEReportaErroDoServidor(t *testing.T) {
	payloadFile := t.TempDir() + "/pedido.json"
	payload := []byte(`{"items":[],"fulfillment":"pickup"}`)
	if err := os.WriteFile(payloadFile, payload, 0o600); err != nil {
		t.Fatal(err)
	}

	var menuHits int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/menu":
			menuHits++
			writeFixture(w, `{"data":{"categories":[]}}`)
		case "/api/v1/orders":
			http.Error(w, "falha interna", http.StatusInternalServerError)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	var output bytes.Buffer
	cfg := config{
		baseURL:           server.URL,
		menuConcurrency:   1,
		ordersConcurrency: 1,
		duplicates:        1,
		timeout:           5 * time.Second,
		payloadPath:       payloadFile,
	}
	err := run(context.Background(), cfg, &output)
	if err == nil || !strings.Contains(err.Error(), "5xx") {
		t.Fatalf("run() erro = %v", err)
	}
	if menuHits != 1 {
		t.Fatalf("menu foi consultado %d vezes, esperado 1", menuHits)
	}
	if !strings.Contains(output.String(), "status=[500=2]") {
		t.Fatalf("relatório = %s", output.String())
	}
}
