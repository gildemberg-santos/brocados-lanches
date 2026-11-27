// Comando loadtest executa o teste de carga HTTP descrito em
// docs/09-qualidade-e-testes.md §11: 50 leituras concorrentes de /menu, pedidos
// concorrentes com chaves distintas e dois pedidos com a mesma chave.
//
// Uso:
//
//	go run ./cmd/loadtest -url http://localhost:8080
//	go run ./cmd/loadtest -url http://localhost:8080 -payload /tmp/pedido.json
//
// Nenhum dado pessoal é impresso: o relatório traz só status, contagens e
// latências. O payload gerado é fictício e usa o catálogo publicado pela API.
package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/brocados-lanches/api/internal/domain"
)

const (
	defaultMenuConcurrency   = 50
	defaultOrdersConcurrency = 10
	defaultDuplicates        = 2
)

type config struct {
	baseURL           string
	menuConcurrency   int
	ordersConcurrency int
	duplicates        int
	timeout           time.Duration
	payloadPath       string
}

type result struct {
	status  int
	elapsed time.Duration
	err     error
}

type stats struct {
	total     int
	success   int
	statuses  map[int]int
	latencies []time.Duration
}

func newStats() stats {
	return stats{statuses: make(map[int]int)}
}

func (s *stats) add(result result) {
	s.total++
	if result.err == nil {
		s.statuses[result.status]++
		if result.status >= 200 && result.status < 300 {
			s.success++
		}
		s.latencies = append(s.latencies, result.elapsed)
	}
}

func (s stats) percentile(fraction float64) time.Duration {
	if len(s.latencies) == 0 {
		return 0
	}
	ordered := append([]time.Duration(nil), s.latencies...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i] < ordered[j] })
	index := int(float64(len(ordered)-1) * fraction)
	return ordered[index]
}

func (s stats) errors() int {
	return s.total - s.success
}

func (s stats) statusSummary() string {
	codes := make([]int, 0, len(s.statuses))
	for code := range s.statuses {
		codes = append(codes, code)
	}
	sort.Ints(codes)
	parts := make([]string, 0, len(codes))
	for _, code := range codes {
		parts = append(parts, fmt.Sprintf("%d=%d", code, s.statuses[code]))
	}
	return fmt.Sprintf("[%s]", joinStrings(parts, " "))
}

func joinStrings(values []string, separator string) string {
	if len(values) == 0 {
		return ""
	}
	result := values[0]
	for _, value := range values[1:] {
		result += separator + value
	}
	return result
}

func printStats(out io.Writer, label string, s stats) error {
	_, err := fmt.Fprintf(
		out,
		"%s: total=%d sucesso=%d erros=%d p50=%s p95=%s p99=%s status=%s\n",
		label,
		s.total,
		s.success,
		s.errors(),
		s.percentile(0.50),
		s.percentile(0.95),
		s.percentile(0.99),
		s.statusSummary(),
	)
	return err
}

func doRequest(
	ctx context.Context,
	client *http.Client,
	method string,
	url string,
	body []byte,
	idempotencyKey string,
) result {
	started := time.Now()
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	request, err := http.NewRequestWithContext(ctx, method, url, reader)
	if err != nil {
		return result{err: err, elapsed: time.Since(started)}
	}
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	if idempotencyKey != "" {
		request.Header.Set("Idempotency-Key", idempotencyKey)
	}

	response, err := client.Do(request)
	if err != nil {
		return result{err: err, elapsed: time.Since(started)}
	}
	defer func() { _ = response.Body.Close() }()
	_, err = io.Copy(io.Discard, response.Body)
	return result{status: response.StatusCode, err: err, elapsed: time.Since(started)}
}

// exercise dispara todas as requisições ao mesmo tempo. bodyFactory e
// keyFactory recebem o índice do worker; ambos podem ser nil.
func exercise(
	ctx context.Context,
	client *http.Client,
	method string,
	url string,
	concurrency int,
	bodyFactory func(int) ([]byte, error),
	keyFactory func(int) string,
) (stats, error) {
	collected := make(chan result, concurrency)
	var waitGroup sync.WaitGroup
	waitGroup.Add(concurrency)

	for index := 0; index < concurrency; index++ {
		go func(index int) {
			defer waitGroup.Done()
			var body []byte
			if bodyFactory != nil {
				created, err := bodyFactory(index)
				if err != nil {
					collected <- result{err: err}
					return
				}
				body = created
			}
			key := ""
			if keyFactory != nil {
				key = keyFactory(index)
			}
			collected <- doRequest(ctx, client, method, url, body, key)
		}(index)
	}

	waitGroup.Wait()
	close(collected)

	summary := newStats()
	var firstError error
	for item := range collected {
		if item.err != nil && firstError == nil {
			firstError = item.err
		}
		summary.add(item)
	}
	return summary, firstError
}

func randomPrefix() (string, error) {
	buffer := make([]byte, 8)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return hex.EncodeToString(buffer), nil
}

func loadPayload(path string) ([]byte, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("ler payload: %w", err)
	}
	if !json.Valid(content) {
		return nil, fmt.Errorf("payload %q não é JSON válido", path)
	}
	return content, nil
}

type catalogMenu struct {
	Data struct {
		Categories []struct {
			Products []struct {
				ID           string `json:"id"`
				Available    bool   `json:"available"`
				OptionGroups []struct {
					ID            string `json:"id"`
					Required      bool   `json:"required"`
					MinSelections int    `json:"minSelections"`
					MaxSelections int    `json:"maxSelections"`
					Options       []struct {
						ID        string `json:"id"`
						Available bool   `json:"available"`
					} `json:"options"`
				} `json:"optionGroups"`
			} `json:"products"`
		} `json:"categories"`
	} `json:"data"`
}

type catalogZones struct {
	Data []struct {
		ID     string `json:"id"`
		Active bool   `json:"active"`
	} `json:"data"`
}

func getJSON(ctx context.Context, client *http.Client, url string, target any) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("GET %s devolveu %d", url, response.StatusCode)
	}
	return json.NewDecoder(response.Body).Decode(target)
}

func falsePointer() *bool {
	value := false
	return &value
}

// discoverPayload monta um pedido fictício com o primeiro produto disponível
// cujos grupos obrigatórios possam ser atendidos e a primeira zona ativa.
func discoverPayload(ctx context.Context, client *http.Client, baseURL string) ([]byte, error) {
	var menu catalogMenu
	if err := getJSON(ctx, client, baseURL+"/api/v1/menu", &menu); err != nil {
		return nil, fmt.Errorf("carregar catálogo: %w", err)
	}
	var zones catalogZones
	if err := getJSON(ctx, client, baseURL+"/api/v1/delivery-zones", &zones); err != nil {
		return nil, fmt.Errorf("carregar zonas: %w", err)
	}

	var zoneID string
	for _, zone := range zones.Data {
		if zone.Active {
			zoneID = zone.ID
			break
		}
	}
	if zoneID == "" {
		return nil, fmt.Errorf("nenhuma zona ativa encontrada no catálogo")
	}

	for _, category := range menu.Data.Categories {
		for _, product := range category.Products {
			if !product.Available || product.ID == "" {
				continue
			}
			optionIDs, ok := requiredOptions(product.OptionGroups)
			if !ok {
				continue
			}
			request := domain.OrderRequest{
				Items: []domain.QuoteItem{{
					ProductID: product.ID,
					Quantity:  1,
					OptionIDs: optionIDs,
				}},
				Fulfillment: domain.FulfillmentDelivery,
				ZoneID:      zoneID,
				Customer: domain.OrderCustomer{
					Name:  "Cliente de carga",
					Phone: "+5585999999999",
				},
				Address: &domain.OrderAddress{
					Street:   "Rua de teste",
					Number:   "1",
					District: "Centro",
					City:     "Aquiraz",
					State:    "CE",
					ZipCode:  "61700000",
				},
				Payment:             domain.OrderPayment{Method: domain.PaymentMethodPix},
				SaveCustomerProfile: falsePointer(),
			}
			encoded, err := json.Marshal(request)
			if err != nil {
				return nil, err
			}
			return encoded, nil
		}
	}
	return nil, fmt.Errorf("nenhum produto disponível com opções válidas encontrado no catálogo")
}

func requiredOptions(groups []struct {
	ID            string `json:"id"`
	Required      bool   `json:"required"`
	MinSelections int    `json:"minSelections"`
	MaxSelections int    `json:"maxSelections"`
	Options       []struct {
		ID        string `json:"id"`
		Available bool   `json:"available"`
	} `json:"options"`
}) ([]string, bool) {
	selected := make([]string, 0)
	for _, group := range groups {
		minimum := group.MinSelections
		if group.Required && minimum == 0 {
			minimum = 1
		}
		if minimum == 0 {
			continue
		}
		if group.MaxSelections > 0 && minimum > group.MaxSelections {
			return nil, false
		}
		chosen := 0
		for _, option := range group.Options {
			if chosen >= minimum {
				break
			}
			if option.Available && option.ID != "" {
				selected = append(selected, option.ID)
				chosen++
			}
		}
		if chosen < minimum {
			return nil, false
		}
	}
	return selected, true
}

// run executa as três fases e escreve o relatório. Foi separada de main para
// permitir teste com httptest sem depender de API ou MongoDB externos.
func run(ctx context.Context, cfg config, out io.Writer) error {
	if cfg.menuConcurrency < 1 || cfg.ordersConcurrency < 1 || cfg.duplicates < 1 {
		return fmt.Errorf("concorrência e duplicatas precisam ser maiores que zero")
	}
	client := &http.Client{Timeout: cfg.timeout}
	baseURL := trimTrailingSlash(cfg.baseURL)

	menuStats, menuErr := exercise(
		ctx, client, http.MethodGet, baseURL+"/api/v1/menu", cfg.menuConcurrency, nil, nil,
	)
	if err := printStats(out, "GET /menu", menuStats); err != nil {
		return fmt.Errorf("escrever relatório: %w", err)
	}

	var payload []byte
	var err error
	if cfg.payloadPath != "" {
		payload, err = loadPayload(cfg.payloadPath)
	} else {
		payload, err = discoverPayload(ctx, client, baseURL)
	}
	if err != nil {
		return err
	}

	prefix, err := randomPrefix()
	if err != nil {
		return fmt.Errorf("gerar prefixo de idempotência: %w", err)
	}
	sharedKey := fmt.Sprintf("loadtest-%s-duplicata", prefix)
	orderCount := cfg.ordersConcurrency + cfg.duplicates
	orderStats, orderErr := exercise(
		ctx,
		client,
		http.MethodPost,
		baseURL+"/api/v1/orders",
		orderCount,
		func(int) ([]byte, error) { return payload, nil },
		func(index int) string {
			if index >= cfg.ordersConcurrency {
				return sharedKey
			}
			return fmt.Sprintf("loadtest-%s-%d", prefix, index)
		},
	)
	if err := printStats(out, "POST /orders", orderStats); err != nil {
		return fmt.Errorf("escrever relatório: %w", err)
	}

	if menuErr != nil {
		return fmt.Errorf("falha de transporte em /menu: %w", menuErr)
	}
	if orderErr != nil {
		return fmt.Errorf("falha de transporte em /orders: %w", orderErr)
	}
	if menuStats.statuses[http.StatusTooManyRequests] > 0 ||
		orderStats.statuses[http.StatusTooManyRequests] > 0 {
		if _, err := fmt.Fprintln(out, "aviso: 429 recebido; o limite por IP pode ter sido atingido"); err != nil {
			return fmt.Errorf("escrever aviso: %w", err)
		}
	}
	var serverErrors []int
	for code := range orderStats.statuses {
		if code >= 500 {
			serverErrors = append(serverErrors, code)
		}
	}
	if len(serverErrors) > 0 {
		return fmt.Errorf("servidor devolveu erro 5xx em /orders: %v", serverErrors)
	}
	return nil
}

func trimTrailingSlash(value string) string {
	for len(value) > 1 && value[len(value)-1] == '/' {
		value = value[:len(value)-1]
	}
	return value
}

func main() {
	var cfg config
	flag.StringVar(&cfg.baseURL, "url", "http://localhost:8080", "URL base da API")
	flag.IntVar(&cfg.menuConcurrency, "menu-concurrency", defaultMenuConcurrency, "requisições concorrentes em GET /menu")
	flag.IntVar(&cfg.ordersConcurrency, "orders-concurrency", defaultOrdersConcurrency, "pedidos concorrentes com chaves distintas")
	flag.IntVar(&cfg.duplicates, "duplicate-orders", defaultDuplicates, "pedidos concorrentes com a mesma chave")
	flag.DurationVar(&cfg.timeout, "timeout", 15*time.Second, "timeout de cada requisição")
	flag.StringVar(&cfg.payloadPath, "payload", "", "arquivo JSON com o corpo do pedido; vazio descobre do catálogo")
	flag.Parse()

	if err := run(context.Background(), cfg, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "loadtest:", err)
		os.Exit(1)
	}
}
