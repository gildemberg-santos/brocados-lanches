package seed

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestCentsFromReais(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int64
		wantErr bool
	}{
		{name: "inteiro", input: "23", want: 2300},
		{name: "uma casa decimal", input: "54.9", want: 5490},
		{name: "duas casas", input: "12.50", want: 1250},
		{name: "zero", input: "0", want: 0},
		{name: "um centavo", input: "0.01", want: 1},
		{name: "com espaços", input: " 7.05 ", want: 705},
		{name: "negativo", input: "-1", wantErr: true},
		{name: "vazio", input: "", wantErr: true},
		{name: "texto", input: "abc", wantErr: true},
		{name: "vírgula", input: "1,50", wantErr: true},
		{name: "três casas", input: "12.345", wantErr: true},
		{name: "ponto final", input: "1.", wantErr: true},
		{name: "sem parte inteira", input: ".5", wantErr: true},
		{name: "dois pontos", input: "1.2.3", wantErr: true},
		{name: "overflow", input: "92233720368547758.08", wantErr: true},
		{name: "limite int64", input: "92233720368547758.07", want: 9223372036854775807},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := CentsFromReais(test.input)
			if test.wantErr {
				if err == nil {
					t.Fatalf("CentsFromReais(%q) error = nil, want error", test.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("CentsFromReais(%q) error = %v", test.input, err)
			}
			if got != test.want {
				t.Fatalf("CentsFromReais(%q) = %d, want %d", test.input, got, test.want)
			}
		})
	}
}

func TestSlug(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{input: "Hambúrgueres", want: "hamburgueres"},
		{input: "Brocado's Burguer Cheddar", want: "brocados-burguer-cheddar"},
		{input: "Porto das Dunas", want: "porto-das-dunas"},
		{input: "Coca-Cola Lata 350ml", want: "coca-cola-lata-350ml"},
		{input: "Ação & Cia", want: "acao-cia"},
		{input: "  ", want: ""},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			if got := Slug(test.input); got != test.want {
				t.Fatalf("Slug(%q) = %q, want %q", test.input, got, test.want)
			}
		})
	}
}

func TestValidateTarget(t *testing.T) {
	tests := []struct {
		name        string
		database    string
		environment string
		wantErr     bool
	}{
		{name: "desenvolvimento", database: "brocados_dev", environment: "development"},
		{name: "teste", database: "brocados_test", environment: "test"},
		{name: "banco isolado de seed", database: "brocados_seed_test", environment: "test"},
		{name: "produção pelo nome", database: "brocados", environment: "development", wantErr: true},
		{name: "produção pelo ambiente", database: "brocados_dev", environment: "production", wantErr: true},
		{name: "banco vazio", database: "", environment: "development", wantErr: true},
		{name: "nome fora do padrão", database: "meu_banco", environment: "development", wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateTarget(test.database, test.environment)
			if test.wantErr && err == nil {
				t.Fatal("ValidateTarget() error = nil, want error")
			}
			if !test.wantErr && err != nil {
				t.Fatalf("ValidateTarget() error = %v", err)
			}
		})
	}
}

func TestValidateRejectsInvalidSeed(t *testing.T) {
	tests := []struct {
		name string
		data Data
	}{
		{
			name: "loja sem nome",
			data: Data{
				Categories: []Category{{ID: "cat_1", Name: "Bebidas"}},
				Products:   []Product{{CategoryID: "cat_1", Name: "Suco", Price: json.Number("5")}},
			},
		},
		{
			name: "categoria duplicada",
			data: Data{
				Store: Store{Name: "Loja"},
				Categories: []Category{
					{ID: "cat_1", Name: "Bebidas"},
					{ID: "cat_2", Name: "bebidas"},
				},
			},
		},
		{
			name: "produto sem categoria",
			data: Data{
				Store:      Store{Name: "Loja"},
				Categories: []Category{{ID: "cat_1", Name: "Bebidas"}},
				Products:   []Product{{CategoryID: "cat_9", Name: "Suco", Price: json.Number("5")}},
			},
		},
		{
			name: "preço inválido",
			data: Data{
				Store:      Store{Name: "Loja"},
				Categories: []Category{{ID: "cat_1", Name: "Bebidas"}},
				Products:   []Product{{CategoryID: "cat_1", Name: "Suco", Price: json.Number("5.555")}},
			},
		},
		{
			name: "taxa inválida",
			data: Data{
				Store:         Store{Name: "Loja"},
				Categories:    []Category{{ID: "cat_1", Name: "Bebidas"}},
				DeliveryZones: []DeliveryZone{{Name: "Centro", Fee: json.Number("abc")}},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := test.data.Validate(); err == nil {
				t.Fatal("Validate() error = nil, want error")
			}
		})
	}
}

func TestBuildIsDeterministic(t *testing.T) {
	data := fakeData()

	first, err := Build(data)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	second, err := Build(data)
	if err != nil {
		t.Fatalf("Build() second error = %v", err)
	}

	if !reflect.DeepEqual(first, second) {
		t.Fatal("Build() returned different plans for the same data")
	}
	if first.Store.Name != "Brocados Lanches" || !first.Store.IsOpen {
		t.Fatalf("Build() store = %+v", first.Store)
	}
	if len(first.Categories) != 2 || first.Categories[0].Slug != "hamburgueres" {
		t.Fatalf("Build() categories = %+v", first.Categories)
	}
	if len(first.Products) != 2 {
		t.Fatalf("Build() products = %v", len(first.Products))
	}
	if first.Products[0].PriceCents != 2300 {
		t.Fatalf("Build() first price = %d, want 2300", first.Products[0].PriceCents)
	}
	if first.Products[1].PriceCents != 5490 {
		t.Fatalf("Build() second price = %d, want 5490", first.Products[1].PriceCents)
	}
	if first.Products[1].CategorySlug != "combos" {
		t.Fatalf("Build() second category = %q, want combos", first.Products[1].CategorySlug)
	}
	if first.Zones[0].FeeCents != 500 {
		t.Fatalf("Build() zone fee = %d, want 500", first.Zones[0].FeeCents)
	}
}

func TestLoadRepositorySeedFile(t *testing.T) {
	path := filepath.Join("..", "..", "..", "design-system", "seed.json")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("seed file not found: %v", err)
	}

	data, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	plan, err := Build(data)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if len(plan.Categories) != 4 {
		t.Fatalf("categories = %d, want 4", len(plan.Categories))
	}
	if len(plan.Products) != 5 {
		t.Fatalf("products = %d, want 5", len(plan.Products))
	}
	if len(plan.Zones) != 3 {
		t.Fatalf("zones = %d, want 3", len(plan.Zones))
	}
	if plan.Products[0].PriceCents != 2300 {
		t.Fatalf("first product price = %d, want 2300", plan.Products[0].PriceCents)
	}
	if plan.Products[2].PriceCents != 5490 {
		t.Fatalf("combo price = %d, want 5490", plan.Products[2].PriceCents)
	}
	if plan.Zones[2].FeeCents != 1200 {
		t.Fatalf("last zone fee = %d, want 1200", plan.Zones[2].FeeCents)
	}
}

func TestLoadRejectsInvalidJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "seed.json")
	if err := os.WriteFile(path, []byte("{invalid"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if _, err := Load(path); err == nil {
		t.Fatal("Load() error = nil, want error")
	}
}

func TestApplyRequiresDatabase(t *testing.T) {
	if _, err := Apply(context.Background(), nil, Plan{Store: StoreSettings{Name: "Loja"}}, false); err == nil {
		t.Fatal("Apply() error = nil, want error")
	}
}

func fakeData() Data {
	return Data{
		Store: Store{
			Name:         "Brocados Lanches",
			IsOpen:       true,
			DeliveryTime: "35-50 min",
			WhatsApp:     "(85) 99999-9999",
			Instagram:    "@brocadoslanches",
			Address:      "Av. Principal, 1200 - Centro, Aquiraz - CE",
		},
		Categories: []Category{
			{ID: "cat_1", Name: "Hambúrgueres", Active: true},
			{ID: "cat_2", Name: "Combos", Active: true},
		},
		Products: []Product{
			{ID: "prod_1", CategoryID: "cat_1", Name: "Brocado's Burguer Cheddar", Price: json.Number("23"), Available: true, Featured: true},
			{ID: "prod_2", CategoryID: "cat_2", Name: "Combo Casal Brocado", Price: json.Number("54.9"), Available: true},
		},
		DeliveryZones: []DeliveryZone{
			{ID: "z1", Name: "Centro", Fee: json.Number("5"), Active: true},
		},
	}
}
