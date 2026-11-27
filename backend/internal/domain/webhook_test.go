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

func asValidationError(err error, target **ValidationError) bool {
	if validation, ok := err.(*ValidationError); ok {
		*target = validation
		return true
	}
	return false
}
