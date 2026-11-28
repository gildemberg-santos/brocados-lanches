package domain

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestPixSettingsValidate(t *testing.T) {
	t.Parallel()

	base := PixSettings{
		Active:       true,
		Key:          "123e4567-e89b-12d3-a456-426614174000",
		MerchantName: "Brocados Lanches",
		MerchantCity: "Aquiraz",
	}

	tests := []struct {
		name    string
		mutate  func(*PixSettings)
		field   string
		wantErr bool
	}{
		{name: "configuracao completa", mutate: func(*PixSettings) {}},
		{
			name: "desativado permite rascunho vazio",
			mutate: func(p *PixSettings) {
				p.Active = false
				p.Key = ""
				p.MerchantName = ""
				p.MerchantCity = ""
			},
		},
		{
			name: "chave obrigatoria",
			mutate: func(p *PixSettings) {
				p.Key = "  "
			},
			field: "key", wantErr: true,
		},
		{
			name: "chave com espaco interno",
			mutate: func(p *PixSettings) {
				p.Key = "123 e4567"
			},
			field: "key", wantErr: true,
		},
		{
			name: "cpf com mascara",
			mutate: func(p *PixSettings) {
				p.Key = "123.456.789-09"
			},
		},
		{
			name: "email valido",
			mutate: func(p *PixSettings) {
				p.Key = "contato@brocados.com.br"
			},
		},
		{
			name: "telefone E.164",
			mutate: func(p *PixSettings) {
				p.Key = "+5585999999999"
			},
		},
		{
			name: "chave no limite do subtemplate 26",
			mutate: func(p *PixSettings) {
				p.Key = strings.Repeat("a", 65) + "@exemplo.com"
			},
		},
		{
			name: "chave longa demais",
			mutate: func(p *PixSettings) {
				p.Key = strings.Repeat("a", 66) + "@exemplo.com"
			},
			field: "key", wantErr: true,
		},
		{
			name: "chave sem tipo reconhecido",
			mutate: func(p *PixSettings) {
				p.Key = "abc"
			},
			field: "key", wantErr: true,
		},
		{
			name: "chave so com simbolos",
			mutate: func(p *PixSettings) {
				p.Key = "@@@"
			},
			field: "key", wantErr: true,
		},
		{
			name: "telefone sem DDI",
			mutate: func(p *PixSettings) {
				p.Key = "+abc"
			},
			field: "key", wantErr: true,
		},
		{
			name: "nome obrigatorio",
			mutate: func(p *PixSettings) {
				p.MerchantName = ""
			},
			field: "merchantName", wantErr: true,
		},
		{
			name: "nome excede limite apos sanitizacao",
			mutate: func(p *PixSettings) {
				p.MerchantName = "12345678901234567890123456"
			},
			field: "merchantName", wantErr: true,
		},
		{
			name: "cidade com acento respeita limite sanitizado",
			mutate: func(p *PixSettings) {
				p.MerchantCity = "São Paulo"
			},
		},
		{
			name: "texto sem caracteres aceitos pelo EMV",
			mutate: func(p *PixSettings) {
				p.MerchantName = "日本語"
			},
			field: "merchantName", wantErr: true,
		},
		{
			name: "somente simbolos",
			mutate: func(p *PixSettings) {
				p.MerchantCity = "***"
			},
			field: "merchantCity", wantErr: true,
		},
		{
			name: "cidade longa demais",
			mutate: func(p *PixSettings) {
				p.MerchantCity = "1234567890123456"
			},
			field: "merchantCity", wantErr: true,
		},
		{
			name: "controle nao permitido",
			mutate: func(p *PixSettings) {
				p.MerchantCity = "Aqui\nraz"
			},
			field: "merchantCity", wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			settings := base
			test.mutate(&settings)
			err := settings.Validate()
			if !test.wantErr {
				if err != nil {
					t.Fatalf("Validate() = %v, want nil", err)
				}
				return
			}
			var validationError *ValidationError
			if !errors.As(err, &validationError) {
				t.Fatalf("Validate() = %v, want ValidationError", err)
			}
			if validationError.Field != test.field {
				t.Errorf("field = %q, want %q", validationError.Field, test.field)
			}
		})
	}
}

func TestPixSettingsPatchApply(t *testing.T) {
	t.Parallel()

	current := PixSettings{
		Active:       true,
		Key:          "chave-antiga@example.com",
		MerchantName: "Nome antigo",
		MerchantCity: "Aquiraz",
		CreatedAt:    time.Date(2026, 11, 27, 12, 0, 0, 0, time.UTC),
	}
	active := false
	key := "  nova-chave@example.com  "
	patch := PixSettingsPatch{Active: &active, Key: &key}

	if patch.Empty() {
		t.Fatal("Empty() = true, want false")
	}
	updated := patch.Apply(current)
	if updated.Active {
		t.Error("Active = true, want false")
	}
	if updated.Key != "nova-chave@example.com" {
		t.Errorf("Key = %q, want trimmed key", updated.Key)
	}
	if updated.MerchantName != current.MerchantName || updated.MerchantCity != current.MerchantCity {
		t.Error("Apply() replaced fields that were not patched")
	}
	if !current.Active || current.Key != "chave-antiga@example.com" {
		t.Error("Apply() mutated the original settings")
	}
	if !(PixSettingsPatch{}).Empty() {
		t.Error("empty patch reported fields to apply")
	}
}
