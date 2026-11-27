package domain

import "testing"

func TestWebhookSettingsValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  WebhookSettings
		wantErr bool
		field   string
	}{
		{
			name:   "inativa sem url",
			config: WebhookSettings{Active: false},
		},
		{
			name:    "ativa sem url",
			config:  WebhookSettings{Active: true},
			wantErr: true,
			field:   "url",
		},
		{
			name:   "url http qualquer host",
			config: WebhookSettings{URL: "http://127.0.0.1:5678/hook", Active: true},
		},
		{
			name:   "url https",
			config: WebhookSettings{URL: "https://hooks.exemplo.invalid/path", Active: true},
		},
		{
			name:   "esquema personalizado",
			config: WebhookSettings{URL: "file:///etc/passwd", Active: true},
		},
		{
			name:   "esquema nao http com host",
			config: WebhookSettings{URL: "ftp://integracao.exemplo.invalid/hook", Active: true},
		},
		{
			name:    "url sem esquema",
			config:  WebhookSettings{URL: "somente/caminho", Active: true},
			wantErr: true,
			field:   "url",
		},
		{
			name:    "inativa sem url com evento inválido",
			config:  WebhookSettings{Events: []string{"order.futuro"}},
			wantErr: true,
			field:   "events[0]",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.config.Validate()
			if test.wantErr {
				if err == nil {
					t.Fatal("esperava erro de validação")
				}
				var validation *ValidationError
				if !asValidationError(err, &validation) {
					t.Fatalf("erro = %T, esperado ValidationError", err)
				}
				if validation.Field != test.field {
					t.Fatalf("field = %q, want %q", validation.Field, test.field)
				}
				return
			}
			if err != nil {
				t.Fatalf("Validate = %v", err)
			}
		})
	}
}

func TestWebhookPatchApply(t *testing.T) {
	current := WebhookSettings{URL: "https://antigo.invalid", Active: true}
	url := "  https://novo.invalid/hook  "
	active := false

	updated := (WebhookSettingsPatch{URL: &url, Active: &active}).Apply(current)
	if updated.URL != "https://novo.invalid/hook" {
		t.Fatalf("url = %q", updated.URL)
	}
	if updated.Active {
		t.Fatal("active deveria ser false")
	}
	if current.URL != "https://antigo.invalid" {
		t.Fatal("Apply alterou a configuração original")
	}
	if !(WebhookSettingsPatch{}).Empty() {
		t.Fatal("patch vazio deveria ser detectado")
	}
}

func TestWebhookEventKeysAreStableAndCopied(t *testing.T) {
	keys := WebhookEventKeys()
	want := []string{
		"order.created", "order.confirmed", "order.preparing", "order.ready",
		"order.out_for_delivery", "order.picked_up", "order.delivered", "order.cancelled",
	}
	if len(keys) != len(want) {
		t.Fatalf("chaves = %v, want %v", keys, want)
	}
	for index, key := range want {
		if keys[index] != key || !IsWebhookEventKey(key) {
			t.Fatalf("chave[%d] = %q, want %q", index, keys[index], key)
		}
	}
	keys[0] = "alterada"
	if WebhookEventKeys()[0] != want[0] {
		t.Fatal("WebhookEventKeys expôs a fatia interna")
	}
	if IsWebhookEventKey("order.desconhecido") {
		t.Fatal("chave desconhecida foi aceita")
	}
}

func TestWebhookEventKeyKeepsGenericHeader(t *testing.T) {
	tests := []struct {
		name  string
		event WebhookEvent
		want  string
	}{
		{
			name:  "criação",
			event: WebhookEvent{Type: WebhookEventOrderCreated, Order: Order{Status: OrderStatusPending}},
			want:  WebhookEventKeyCreated,
		},
		{
			name:  "transição confirmada",
			event: WebhookEvent{Type: WebhookEventOrderUpdated, Order: Order{Status: OrderStatusConfirmed}},
			want:  WebhookEventKeyConfirmed,
		},
		{
			name:  "transição cancelada",
			event: WebhookEvent{Type: WebhookEventOrderUpdated, Order: Order{Status: OrderStatusCancelled}},
			want:  WebhookEventKeyCancelled,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.event.Key(); got != test.want {
				t.Fatalf("Key() = %q, want %q", got, test.want)
			}
			if test.event.Type == WebhookEventOrderCreated && test.want != WebhookEventOrderCreated {
				t.Fatal("cabeçalho de criação não deve ser alterado")
			}
			if test.event.Type == WebhookEventOrderUpdated && test.want == WebhookEventOrderUpdated {
				t.Fatal("cabeçalho de transição não deve ser usado como chave de status")
			}
		})
	}
}

func TestWebhookEventSelection(t *testing.T) {
	legacy := WebhookSettings{}
	if len(legacy.EnabledEvents()) != len(WebhookEventKeys()) {
		t.Fatal("configuração legada sem eventos deve equivaler a todos habilitados")
	}
	for _, key := range WebhookEventKeys() {
		if !legacy.EventEnabled(key) {
			t.Fatalf("configuração legada não habilitou %q", key)
		}
	}
	if !legacy.EventEnabled("order.futuro") {
		t.Fatal("configuração legada deveria preservar o envio genérico de status futuros")
	}

	selected := WebhookSettings{Events: []string{WebhookEventKeyCreated, WebhookEventKeyCancelled}}
	if !selected.EventEnabled(WebhookEventKeyCreated) || !selected.EventEnabled(WebhookEventKeyCancelled) {
		t.Fatal("eventos selecionados não foram habilitados")
	}
	if selected.EventEnabled(WebhookEventKeyConfirmed) || selected.EventEnabled("order.futuro") {
		t.Fatal("evento fora da seleção foi habilitado")
	}

	none := WebhookSettings{Events: []string{}}
	if len(none.EnabledEvents()) != 0 || none.EventEnabled(WebhookEventKeyCreated) || none.EventEnabled("order.futuro") {
		t.Fatal("lista vazia deveria desabilitar todos os eventos")
	}

	enabled := legacy.EnabledEvents()
	enabled[0] = "alterada"
	if !legacy.EventEnabled(WebhookEventKeys()[0]) {
		t.Fatal("EnabledEvents expôs a configuração gravada")
	}
}

func TestWebhookSettingsValidateEvents(t *testing.T) {
	tests := []struct {
		name    string
		events  []string
		wantErr bool
		field   string
	}{
		{name: "sem seleção", events: nil},
		{name: "seleção vazia", events: []string{}},
		{name: "chaves válidas", events: []string{WebhookEventKeyCreated, WebhookEventKeyDelivered}},
		{name: "chave desconhecida", events: []string{"order.future"}, wantErr: true, field: "events[0]"},
		{name: "chave repetida", events: []string{WebhookEventKeyCreated, WebhookEventKeyCreated}, wantErr: true, field: "events[1]"},
		{name: "cabeçalho genérico não é chave", events: []string{WebhookEventOrderUpdated}, wantErr: true, field: "events[0]"},
		{name: "muitas chaves", events: []string{"a", "b", "c", "d", "e", "f", "g", "h", "i"}, wantErr: true, field: "events"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := (WebhookSettings{URL: "https://hooks.exemplo.invalid", Active: true, Events: test.events}).Validate()
			if !test.wantErr {
				if err != nil {
					t.Fatalf("Validate = %v", err)
				}
				return
			}
			var validation *ValidationError
			if !asValidationError(err, &validation) || validation.Field != test.field {
				t.Fatalf("erro = %v, want ValidationError no campo %s", err, test.field)
			}
		})
	}
}

func TestWebhookPatchApplyEvents(t *testing.T) {
	current := WebhookSettings{Events: []string{WebhookEventKeyCreated}}
	selected := []string{WebhookEventKeyConfirmed, WebhookEventKeyCreated}
	patch := WebhookSettingsPatch{Events: &selected}

	updated := patch.Apply(current)
	if len(updated.Events) != 2 || updated.Events[0] != WebhookEventKeyConfirmed {
		t.Fatalf("events = %v", updated.Events)
	}
	selected[0] = "alterada"
	if updated.Events[0] != WebhookEventKeyConfirmed {
		t.Fatal("Apply reutilizou a fatia recebida")
	}
	if len(current.Events) != 1 || current.Events[0] != WebhookEventKeyCreated {
		t.Fatal("Apply alterou a configuração original")
	}
	if !(WebhookSettingsPatch{}).Empty() {
		t.Fatal("patch sem eventos deveria ser detectado como vazio")
	}
	emptySelection := []string{}
	if (WebhookSettingsPatch{Events: &emptySelection}).Empty() {
		t.Fatal("seleção vazia é uma alteração explícita")
	}
}

func asValidationError(err error, target **ValidationError) bool {
	if validation, ok := err.(*ValidationError); ok {
		*target = validation
		return true
	}
	return false
}
