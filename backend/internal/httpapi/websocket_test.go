package httpapi

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/brocados-lanches/api/internal/domain"
	"github.com/brocados-lanches/api/internal/realtime"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/go-chi/chi/v5"
)

func wsTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func wsServerURL(httpURL string) string {
	return strings.Replace(httpURL, "http://", "ws://", 1)
}

func newWSTestServer(hub *realtime.Hub) *httptest.Server {
	router := chi.NewRouter()
	router.Get("/api/v1/orders/{publicId}/events", handlePublicOrderEvents(hub, wsTestLogger()))
	return httptest.NewServer(router)
}

func dialWSTest(t *testing.T, url string, headers http.Header) *websocket.Conn {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	conn, _, err := websocket.Dial(ctx, url, &websocket.DialOptions{HTTPHeader: headers})
	if err != nil {
		t.Fatalf("websocket.Dial(%q) error = %v", url, err)
	}
	t.Cleanup(func() { _ = conn.CloseNow() })
	return conn
}

func readWSEvent(t *testing.T, conn *websocket.Conn) map[string]any {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var event map[string]any
	if err := wsjson.Read(ctx, conn, &event); err != nil {
		t.Fatalf("wsjson.Read error = %v", err)
	}
	return event
}

func TestPublicOrderEventsStreamEnviaSomenteDadosMinimos(t *testing.T) {
	hub := realtime.NewHub()
	defer hub.Close()
	server := newWSTestServer(hub)
	defer server.Close()

	headers := http.Header{"Origin": []string{server.URL}}
	conn := dialWSTest(t, wsServerURL(server.URL)+"/api/v1/orders/publico-1234567890/events", headers)
	if connected := readWSEvent(t, conn); connected["type"] != realtime.EventConnected {
		t.Fatalf("primeiro evento = %v, want connected", connected)
	}

	hub.Publish(realtime.TopicForPublic("publico-1234567890"), realtime.Event{
		Type: realtime.EventOrderUpdated, OrderNumber: 42, Status: "ready",
		UpdatedAt: "2026-11-26T20:00:00Z",
	})
	event := readWSEvent(t, conn)
	if event["status"] != "ready" || event["orderNumber"] != float64(42) {
		t.Fatalf("evento = %v, want status ready e orderNumber 42", event)
	}
	for _, forbidden := range []string{"orderId", "customer", "name", "phone", "address", "note", "feeCents", "totalCents", "whatsappUrl"} {
		if _, found := event[forbidden]; found {
			t.Errorf("evento público expôs %q: %v", forbidden, event)
		}
	}

	hub.Publish(realtime.TopicForPublic("outro-pedido-123"), realtime.Event{Type: realtime.EventOrderUpdated, Status: "ready"})
	hub.Publish(realtime.TopicAdmin, realtime.Event{Type: realtime.EventOrderUpdated, Status: "ready"})
	quietContext, cancelQuiet := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancelQuiet()
	var unexpected map[string]any
	if err := wsjson.Read(quietContext, conn, &unexpected); err == nil {
		t.Fatalf("recebeu evento de outro tópico: %v", unexpected)
	}
}

func TestPublicOrderEventsRejeitaIDInvalidoOrigemCrossSite(t *testing.T) {
	hub := realtime.NewHub()
	defer hub.Close()

	invalidRouter := chi.NewRouter()
	invalidRouter.Get("/api/v1/orders/{publicId}/events", handlePublicOrderEvents(hub, wsTestLogger()))
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/orders/inv@lido/events", nil)
	invalidRouter.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status para publicId inválido = %d, want 404", recorder.Code)
	}

	server := newWSTestServer(hub)
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, _, err := websocket.Dial(ctx, wsServerURL(server.URL)+"/api/v1/orders/publico-1234567890/events", &websocket.DialOptions{
		HTTPHeader: http.Header{"Origin": []string{"http://site-estranho.example"}},
	})
	if err == nil {
		t.Fatal("handshake cross-site deveria falhar")
	}
}

func TestAdminOrderEventsExigeSessaoEAutorizacao(t *testing.T) {
	hub := realtime.NewHub()
	defer hub.Close()
	router := NewRouter(Dependencies{
		AdminAuth:   newAdminOperatorAuth(),
		AdminCookie: AdminCookieConfig{Name: "brocados_admin"},
		Events:      hub,
		Logger:      wsTestLogger(),
	})
	server := httptest.NewServer(router)
	defer server.Close()
	url := wsServerURL(server.URL) + "/api/v1/admin/orders/events"

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if _, _, err := websocket.Dial(ctx, url, &websocket.DialOptions{
		HTTPHeader: http.Header{"Origin": []string{server.URL}},
	}); err == nil {
		t.Fatal("handshake sem sessão deveria falhar")
	}

	headers := http.Header{
		"Origin": []string{server.URL},
		"Cookie": []string{"brocados_admin=token-opaco"},
	}
	conn := dialWSTest(t, url, headers)
	if connected := readWSEvent(t, conn); connected["type"] != realtime.EventConnected {
		t.Fatalf("primeiro evento = %v, want connected", connected)
	}

	hub.Publish(realtime.TopicAdmin, realtime.Event{
		Type: realtime.EventOrderUpdated, OrderID: "665f1c2a9b1e4a0012a3b4f1",
		OrderNumber: 42, Status: "ready", UpdatedAt: "2026-11-26T20:00:00Z",
	})
	event := readWSEvent(t, conn)
	if event["orderId"] != "665f1c2a9b1e4a0012a3b4f1" || event["status"] != "ready" {
		t.Fatalf("evento admin = %v", event)
	}
}

func TestPublishOrderEventSeparaCanais(t *testing.T) {
	hub := realtime.NewHub()
	defer hub.Close()
	admin, ok := hub.Subscribe(realtime.TopicAdmin)
	if !ok {
		t.Fatal("Subscribe(admin) = false")
	}
	defer admin.Close()
	public, ok := hub.Subscribe(realtime.TopicForPublic("opaco"))
	if !ok {
		t.Fatal("Subscribe(public) = false")
	}
	defer public.Close()

	order := adminOrderDetailFixture()
	order.UpdatedAt = time.Date(2026, 11, 26, 20, 0, 0, 0, time.UTC)
	publishOrderEvent(hub, realtime.EventOrderUpdated, order)

	adminEvent := <-admin.Events
	if adminEvent.Type != realtime.EventOrderUpdated || adminEvent.OrderID != order.ID || adminEvent.PublicID != order.PublicID {
		t.Fatalf("evento admin = %+v", adminEvent)
	}
	publicEvent := <-public.Events
	if publicEvent.OrderID != "" || publicEvent.Type != realtime.EventOrderUpdated || publicEvent.Status != order.Status {
		t.Fatalf("evento público = %+v", publicEvent)
	}
	if publicEvent.UpdatedAt != "2026-11-26T20:00:00Z" {
		t.Fatalf("updatedAt = %q", publicEvent.UpdatedAt)
	}
}

func TestPublishOrderDeletedAvisaApenasAdmin(t *testing.T) {
	hub := realtime.NewHub()
	defer hub.Close()
	admin, ok := hub.Subscribe(realtime.TopicAdmin)
	if !ok {
		t.Fatal("Subscribe(admin) = false")
	}
	defer admin.Close()
	public, ok := hub.Subscribe(realtime.TopicForPublic("opaco"))
	if !ok {
		t.Fatal("Subscribe(public) = false")
	}
	defer public.Close()

	order := adminOrderDetailFixture()
	order.UpdatedAt = time.Date(2026, 11, 27, 10, 0, 0, 0, time.UTC)
	publishOrderDeleted(hub, order)

	event := <-admin.Events
	if event.Type != realtime.EventOrderDeleted || event.OrderID != order.ID ||
		event.PublicID != order.PublicID || event.OrderNumber != order.Number {
		t.Fatalf("evento admin = %+v", event)
	}
	select {
	case publicEvent := <-public.Events:
		t.Fatalf("canal público não deve receber exclusão: %+v", publicEvent)
	default:
	}
}

func TestCreateOrderPublicaEventoApenasNaPrimeiraGravacao(t *testing.T) {
	for _, test := range []struct {
		name     string
		replayed bool
		want     bool
	}{
		{name: "pedido novo", want: true},
		{name: "replay idempotente", replayed: true, want: false},
	} {
		t.Run(test.name, func(t *testing.T) {
			hub := realtime.NewHub()
			defer hub.Close()
			admin, ok := hub.Subscribe(realtime.TopicAdmin)
			if !ok {
				t.Fatal("Subscribe(admin) = false")
			}
			defer admin.Close()
			public, ok := hub.Subscribe(realtime.TopicForPublic("public-123"))
			if !ok {
				t.Fatal("Subscribe(public) = false")
			}
			defer public.Close()

			orders := &fakeOrderService{
				created:  createdOrderFixture(),
				replayed: test.replayed,
				store:    domain.StoreSettings{WhatsApp: "(85) 99999-9999"},
			}
			router := NewRouter(Dependencies{
				Pinger: fakePinger{},
				Orders: orders,
				Events: hub,
				Logger: wsTestLogger(),
			})
			response := performCreateOrderRequest(router, testOrderHeaderKey, validOrderBody)
			if response.Code != http.StatusCreated && response.Code != http.StatusOK {
				t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
			}

			select {
			case event := <-admin.Events:
				if !test.want {
					t.Fatalf("replay publicou evento: %+v", event)
				}
				if event.Type != realtime.EventOrderCreated || event.PublicID != "public-123" || event.Status != domain.OrderStatusPending {
					t.Fatalf("evento admin = %+v", event)
				}
			default:
				if test.want {
					t.Fatal("criação não publicou no tópico admin")
				}
			}
			select {
			case event := <-public.Events:
				if !test.want {
					t.Fatalf("replay publicou evento público: %+v", event)
				}
				if event.OrderID != "" || event.Type != realtime.EventOrderCreated {
					t.Fatalf("evento público = %+v", event)
				}
			default:
				if test.want {
					t.Fatal("criação não publicou no tópico público")
				}
			}
		})
	}
}

func TestAdminUpdateOrderStatusPublicaEvento(t *testing.T) {
	hub := realtime.NewHub()
	defer hub.Close()
	orders := &fakeAdminOrdersHandlerService{order: adminOrderDetailFixture()}
	router := NewRouter(Dependencies{
		AdminAuth:   newAdminOperatorAuth(),
		AdminCookie: AdminCookieConfig{Name: "brocados_admin"},
		AdminOrders: orders,
		Events:      hub,
		Logger:      wsTestLogger(),
	})
	admin, ok := hub.Subscribe(realtime.TopicAdmin)
	if !ok {
		t.Fatal("Subscribe(admin) = false")
	}
	defer admin.Close()
	public, ok := hub.Subscribe(realtime.TopicForPublic("opaco"))
	if !ok {
		t.Fatal("Subscribe(public) = false")
	}
	defer public.Close()

	recorder := httptest.NewRecorder()
	request := newAdminRequest(http.MethodPatch, "/api/v1/admin/orders/665f1c2a9b1e4a0012a3b4f1/status", `{"status":"confirmed"}`)
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status PATCH = %d, body = %s", recorder.Code, recorder.Body.String())
	}

	select {
	case event := <-admin.Events:
		if event.Status != "confirmed" || event.OrderID == "" {
			t.Fatalf("evento admin = %+v", event)
		}
	default:
		t.Fatal("PATCH não publicou no tópico admin")
	}
	select {
	case event := <-public.Events:
		if event.Status != "confirmed" || event.OrderID != "" {
			t.Fatalf("evento público = %+v", event)
		}
	default:
		t.Fatal("PATCH não publicou no tópico público")
	}

	var body struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil || len(body.Data) == 0 {
		t.Fatalf("resposta PATCH inválida: %s", recorder.Body.String())
	}
}
