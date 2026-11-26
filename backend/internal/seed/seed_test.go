package seed

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"unicode"

	"github.com/brocados-lanches/api/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
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
		{
			name: "imagem sem texto alternativo",
			data: Data{
				Store:      Store{Name: "Loja"},
				Categories: []Category{{ID: "cat_1", Name: "Bebidas"}},
				Products:   []Product{{CategoryID: "cat_1", Name: "Suco", Price: json.Number("5"), Image: "/products/suco.png"}},
			},
		},
		{
			name: "imagem fora do diretório local",
			data: Data{
				Store:      Store{Name: "Loja"},
				Categories: []Category{{ID: "cat_1", Name: "Bebidas"}},
				Products:   []Product{{CategoryID: "cat_1", Name: "Suco", Price: json.Number("5"), Image: "../remoto.png", ImageAlt: "Suco"}},
			},
		},
		{
			name: "imagem duplicada",
			data: Data{
				Store:      Store{Name: "Loja"},
				Categories: []Category{{ID: "cat_1", Name: "Bebidas"}},
				Products: []Product{
					{CategoryID: "cat_1", Name: "Suco", Price: json.Number("5"), Image: "/products/suco.png", ImageAlt: "Suco"},
					{CategoryID: "cat_1", Name: "Refrigerante", Price: json.Number("6"), Image: "/products/suco.png", ImageAlt: "Refrigerante"},
				},
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

func TestValidateRejectsInvalidStoreSettings(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Data)
	}{
		{name: "whatsapp com máscara", mutate: func(data *Data) { data.Store.WhatsApp = "(85) 99999-9999" }},
		{name: "sem deliveryTime", mutate: func(data *Data) { data.Store.DeliveryTime = "" }},
		{name: "sem endereço", mutate: func(data *Data) { data.Store.Address = "" }},
		{name: "dia inválido", mutate: func(data *Data) { data.Store.OpeningHours[0].Day = "domingo" }},
		{
			name: "horário inválido",
			mutate: func(data *Data) {
				data.Store.OpeningHours[0].Intervals[0].Open = "18h"
			},
		},
		{
			name: "dia duplicado",
			mutate: func(data *Data) {
				data.Store.OpeningHours = append(data.Store.OpeningHours, domain.OpeningHours{Day: "sunday", Closed: true})
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			data := fakeData()
			test.mutate(&data)
			if err := data.Validate(); err == nil {
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
	if !reflect.DeepEqual(first.Store.OpeningHours, data.Store.OpeningHours) {
		t.Fatalf("Build() opening hours = %+v, want %+v", first.Store.OpeningHours, data.Store.OpeningHours)
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
	if first.Products[0].ImageURL != "/products/brocados-burguer-cheddar.png" || first.Products[0].ImageAlt != "Brocado's Burguer Cheddar" {
		t.Fatalf("Build() first image = %q / %q", first.Products[0].ImageURL, first.Products[0].ImageAlt)
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

	if len(plan.Categories) != 5 {
		t.Fatalf("categories = %d, want 5", len(plan.Categories))
	}
	if len(plan.Products) != 37 {
		t.Fatalf("products = %d, want 37", len(plan.Products))
	}
	if len(plan.Zones) != 30 {
		t.Fatalf("zones = %d, want 30", len(plan.Zones))
	}
	if !data.Store.IsOpen || data.Store.WhatsApp != "+5585992358708" {
		t.Fatalf("store contact = %+v, want open store with +5585992358708", data.Store)
	}
	if len(plan.Store.OpeningHours) != 7 {
		t.Fatalf("opening hours = %d, want 7", len(plan.Store.OpeningHours))
	}
	if plan.Store.OpeningHours[0].Day != "sunday" || plan.Store.OpeningHours[2].Day != "tuesday" || !plan.Store.OpeningHours[2].Closed {
		t.Fatalf("unexpected opening hours = %+v", plan.Store.OpeningHours)
	}
	if got := plan.Store.OpeningHours[0].Intervals[0]; got.Open != "17:45" || got.Close != "23:30" {
		t.Fatalf("sunday interval = %+v, want 17:45-23:30", got)
	}
	if got := plan.Store.OpeningHours[5].Intervals[0]; got.Open != "17:30" || got.Close != "23:30" {
		t.Fatalf("friday interval = %+v, want 17:30-23:30", got)
	}
	if plan.Products[0].PriceCents != 4999 {
		t.Fatalf("first product price = %d, want 4999", plan.Products[0].PriceCents)
	}
	if plan.Products[2].PriceCents != 8999 {
		t.Fatalf("combo price = %d, want 8999", plan.Products[2].PriceCents)
	}
	if plan.Zones[0].Slug != "araca-do-biel" || plan.Zones[0].FeeCents != 600 {
		t.Fatalf("first zone = %+v, want araca-do-biel with 600 cents", plan.Zones[0])
	}
	if plan.Zones[29].Name != "Riviera" || plan.Zones[29].FeeCents != 1200 {
		t.Fatalf("last zone = %+v, want Riviera with 1200 cents", plan.Zones[29])
	}

	featured := make([]string, 0, 3)
	for _, product := range plan.Products {
		if !product.Featured {
			continue
		}
		if !product.Available {
			t.Errorf("featured product %q is not available", product.Name)
		}
		featured = append(featured, product.Slug)
	}
	wantFeatured := []string{"combo-3-hamburguer", "burger-frescurinha", "batata-g"}
	if !reflect.DeepEqual(featured, wantFeatured) {
		t.Fatalf("featured products = %v, want %v", featured, wantFeatured)
	}
}

func TestSeedCategoryNamesHaveNoEmoji(t *testing.T) {
	path := filepath.Join("..", "..", "..", "design-system", "seed.json")
	data, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(data.Categories) == 0 {
		t.Fatal("o seed não trouxe categorias para verificar")
	}

	for _, category := range data.Categories {
		for _, character := range category.Name {
			if unicode.IsLetter(character) || unicode.IsDigit(character) || unicode.IsSpace(character) {
				continue
			}
			t.Errorf("categoria %q contém caractere não textual %q", category.Name, character)
		}
	}
}

func TestRepositorySeedImagesExist(t *testing.T) {
	path := filepath.Join("..", "..", "..", "design-system", "seed.json")
	data, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	plan, err := Build(data)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	seen := make(map[string]bool, len(plan.Products))
	for _, product := range plan.Products {
		if product.ImageURL == "" || product.ImageAlt == "" {
			t.Fatalf("product %q is missing image or alt", product.Name)
		}
		if seen[product.ImageURL] {
			t.Fatalf("duplicate image %q", product.ImageURL)
		}
		seen[product.ImageURL] = true

		imagePath := filepath.Join("..", "..", "..", "frontend", "public", strings.TrimPrefix(product.ImageURL, "/"))
		file, err := os.Stat(imagePath)
		if err != nil {
			t.Errorf("image %q for product %q not found: %v", product.ImageURL, product.Name, err)
			continue
		}
		if file.Size() == 0 {
			t.Errorf("image %q for product %q is empty", product.ImageURL, product.Name)
		}
	}
	if len(seen) != 37 {
		t.Fatalf("images = %d, want 37", len(seen))
	}
}

func TestSeedDeliveryZonesMatchProvidedList(t *testing.T) {
	expected := map[string]int64{
		"Araçá do biel":                 600,
		"Canto verde":                   600,
		"Vila pagã":                     500,
		"Estrada Nova":                  500,
		"Pau pombo":                     500,
		"Iguape":                        800,
		"Novo Iguape":                   800,
		"Patacas":                       700,
		"Araçá do manduca":              800,
		"Área verde":                    400,
		"Engenho velho":                 500,
		"Sítio extremo":                 700,
		"Presídio ( até a Praia)":       800,
		"Canoa":                         400,
		"Araçá da Mônica":               800,
		"Presídio ( até a rua da bica)": 700,
		"Smart city":                    700,
		"Fagundes":                      500,
		"Praia bella":                   900,
		"Tapera":                        300,
		"Portelinha":                    500,
		"Sítios Novos":                  600,
		"Vila Isabel":                   900,
		"Lagoa de cima até o açaí":      1000,
		"Catu":                          700,
		"Machuca":                       1500,
		"Divineia":                      1500,
		"Centro de Aquiraz":             1600,
		"Barro preto":                   1000,
		"Riviera":                       1200,
	}

	path := filepath.Join("..", "..", "..", "design-system", "seed.json")
	data, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	plan, err := Build(data)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if len(plan.Zones) != len(expected) {
		t.Fatalf("zones = %d, want %d", len(plan.Zones), len(expected))
	}

	seenSlugs := make(map[string]string, len(plan.Zones))
	for _, zone := range plan.Zones {
		want, ok := expected[zone.Name]
		if !ok {
			t.Errorf("unexpected zone %q", zone.Name)
			continue
		}
		if zone.FeeCents != want {
			t.Errorf("zone %q fee = %d, want %d", zone.Name, zone.FeeCents, want)
		}
		if !zone.Active {
			t.Errorf("zone %q is not active", zone.Name)
		}
		if previous, exists := seenSlugs[zone.Slug]; exists {
			t.Errorf("duplicate slug %q for %q and %q", zone.Slug, previous, zone.Name)
		}
		seenSlugs[zone.Slug] = zone.Name
		delete(expected, zone.Name)
	}
	for name := range expected {
		t.Errorf("missing zone %q", name)
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

func TestForEnvironment(t *testing.T) {
	data := fakeData()
	plan, err := Build(data)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	plan.Store.IsOpen = false
	plan.Store.OpeningHours = []domain.OpeningHours{{
		Day:       "sunday",
		Intervals: []domain.OpeningInterval{{Open: "18:00", Close: "23:30"}},
	}}

	testPlan := ForEnvironment(plan, "test")
	if !testPlan.Store.IsOpen || len(testPlan.Store.OpeningHours) != 0 {
		t.Fatalf("ForEnvironment(test) store = %+v", testPlan.Store)
	}

	developmentPlan := ForEnvironment(plan, "development")
	if developmentPlan.Store.IsOpen || len(developmentPlan.Store.OpeningHours) != 1 {
		t.Fatalf("ForEnvironment(development) store = %+v", developmentPlan.Store)
	}
}

func TestOpeningHoursBSON(t *testing.T) {
	hours := []domain.OpeningHours{
		{
			Day:    "sunday",
			Closed: false,
			Intervals: []domain.OpeningInterval{
				{Open: "17:45", Close: "23:30"},
			},
		},
		{Day: "tuesday", Closed: true, Intervals: []domain.OpeningInterval{}},
	}

	want := bson.A{
		bson.D{
			{Key: "day", Value: "sunday"},
			{Key: "closed", Value: false},
			{Key: "intervals", Value: bson.A{
				bson.D{{Key: "open", Value: "17:45"}, {Key: "close", Value: "23:30"}},
			}},
		},
		bson.D{
			{Key: "day", Value: "tuesday"},
			{Key: "closed", Value: true},
			{Key: "intervals", Value: bson.A{}},
		},
	}

	if got := openingHoursBSON(hours); !reflect.DeepEqual(got, want) {
		t.Fatalf("openingHoursBSON() = %+v, want %+v", got, want)
	}
}

func fakeData() Data {
	return Data{
		Store: Store{
			Name:         "Brocados Lanches",
			IsOpen:       true,
			DeliveryTime: "35-50 min",
			WhatsApp:     "+5585999999999",
			Instagram:    "@brocadoslanches",
			Address:      "Av. Principal, 1200 - Centro, Aquiraz - CE",
			OpeningHours: []domain.OpeningHours{{
				Day:       "sunday",
				Closed:    false,
				Intervals: []domain.OpeningInterval{{Open: "18:00", Close: "23:30"}},
			}},
		},
		Categories: []Category{
			{ID: "cat_1", Name: "Hambúrgueres", Active: true},
			{ID: "cat_2", Name: "Combos", Active: true},
		},
		Products: []Product{
			{ID: "prod_1", CategoryID: "cat_1", Name: "Brocado's Burguer Cheddar", Price: json.Number("23"), Image: "/products/brocados-burguer-cheddar.png", ImageAlt: "Brocado's Burguer Cheddar", Available: true, Featured: true},
			{ID: "prod_2", CategoryID: "cat_2", Name: "Combo Casal Brocado", Price: json.Number("54.9"), Image: "/products/combo-casal-brocado.png", ImageAlt: "Combo Casal Brocado", Available: true},
		},
		DeliveryZones: []DeliveryZone{
			{ID: "z1", Name: "Centro", Fee: json.Number("5"), Active: true},
		},
	}
}
