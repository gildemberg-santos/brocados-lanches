package main

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/brocados-lanches/api/internal/config"
	"github.com/brocados-lanches/api/internal/seed"
)

func repositorySeedPath() string {
	return filepath.Join("..", "..", "..", "..", "design-system", "seed.json")
}

func TestParseOptions(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want options
	}{
		{
			name: "padrões",
			args: nil,
			want: options{file: defaultFile},
		},
		{
			name: "aplicar com confirmação",
			args: []string{"-file", "/tmp/seed.json", "-apply", "-confirm", confirmation},
			want: options{file: "/tmp/seed.json", apply: true, confirm: confirmation},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := parseOptions(test.args)
			if err != nil {
				t.Fatalf("parseOptions() error = %v", err)
			}
			if got != test.want {
				t.Fatalf("parseOptions() = %+v, want %+v", got, test.want)
			}
		})
	}
}

func TestParseOptionsRejectsUnknownFlag(t *testing.T) {
	if _, err := parseOptions([]string{"-unknown"}); err == nil {
		t.Fatal("parseOptions() error = nil, want error")
	}
}

func TestValidateTarget(t *testing.T) {
	tests := []struct {
		name    string
		cfg     config.Config
		confirm string
		wantErr bool
	}{
		{
			name:    "produção com confirmação",
			cfg:     config.Config{Env: "production", MongoDB: "brocados"},
			confirm: confirmation,
		},
		{
			name:    "produção sem confirmação",
			cfg:     config.Config{Env: "production", MongoDB: "brocados"},
			confirm: "",
			wantErr: true,
		},
		{
			name:    "produção com banco incorreto",
			cfg:     config.Config{Env: "production", MongoDB: "brocados_dev"},
			confirm: confirmation,
			wantErr: true,
		},
		{
			name:    "desenvolvimento",
			cfg:     config.Config{Env: "development", MongoDB: "brocados_dev"},
			confirm: confirmation,
		},
		{
			name:    "desenvolvimento com banco incorreto",
			cfg:     config.Config{Env: "development", MongoDB: "brocados"},
			confirm: confirmation,
			wantErr: true,
		},
		{
			name:    "teste",
			cfg:     config.Config{Env: "test", MongoDB: "brocados_test"},
			confirm: confirmation,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateTarget(test.cfg, test.confirm)
			if (err != nil) != test.wantErr {
				t.Fatalf("validateTarget() error = %v, wantErr %t", err, test.wantErr)
			}
		})
	}
}

func TestRunDryRunValidatesRepositorySeed(t *testing.T) {
	getenv := func(string) string { return "" }
	if err := run([]string{"-file", repositorySeedPath()}, getenv); err != nil {
		t.Fatalf("run() error = %v", err)
	}
}

func TestRunRequiresExplicitConfirmation(t *testing.T) {
	env := map[string]string{
		"APP_ENV":             "production",
		"MONGO_URI":           "mongodb://localhost:27017",
		"MONGO_DB":            "brocados",
		"ORDER_ACCESS_SECRET": "segredo-de-producao-com-mais-de-32-caracteres",
		"MIGRATION_CONFIRM":   "wrong",
	}
	getenv := func(key string) string { return env[key] }

	err := run([]string{"-file", repositorySeedPath(), "-apply", "-confirm", "wrong"}, getenv)
	if err == nil || !strings.Contains(err.Error(), "confirmação") {
		t.Fatalf("run() error = %v, want confirmation error", err)
	}
}

func TestMigrationMarker(t *testing.T) {
	marker := migrationMarker()
	if len(marker) != 2 {
		t.Fatalf("migrationMarker() = %+v, want version and source", marker)
	}
	if marker[0].Key != "version" || marker[0].Value != version {
		t.Fatalf("marker version = %+v", marker[0])
	}
	if marker[1].Key != "source" || marker[1].Value != defaultFile {
		t.Fatalf("marker source = %+v", marker[1])
	}
}

func TestKillSwitchValue(t *testing.T) {
	open := true
	closed := false

	tests := []struct {
		name    string
		current *bool
		want    bool
	}{
		{name: "valor ausente usa padrão do seed", current: nil, want: true},
		{name: "loja aberta", current: &open, want: true},
		{name: "loja fechada", current: &closed, want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := killSwitchValue(test.current); got != test.want {
				t.Fatalf("killSwitchValue() = %t, want %t", got, test.want)
			}
		})
	}
}

func TestLoadKillSwitchRequiresDatabase(t *testing.T) {
	if _, err := loadKillSwitch(context.Background(), nil); err == nil {
		t.Fatal("loadKillSwitch() error = nil, want error")
	}
}

func TestDeactivateStaleRequiresDatabase(t *testing.T) {
	if _, err := deactivateStale(context.Background(), nil, seed.Plan{}); err == nil {
		t.Fatal("deactivateStale() error = nil, want error")
	}
}

func TestDeactivateCollectionRejectsEmptyPlan(t *testing.T) {
	if _, err := deactivateCollection(context.Background(), nil, seed.CollectionCategories, nil, time.Now(), nil); err == nil {
		t.Fatal("deactivateCollection() error = nil, want error")
	}
}

func TestSlugHelpersUseRepositorySeed(t *testing.T) {
	data, err := seed.Load(repositorySeedPath())
	if err != nil {
		t.Fatalf("seed.Load() error = %v", err)
	}
	plan, err := seed.Build(data)
	if err != nil {
		t.Fatalf("seed.Build() error = %v", err)
	}

	if got := len(categorySlugs(plan)); got != 5 {
		t.Fatalf("categorySlugs() = %d, want 5", got)
	}
	if got := len(productSlugs(plan)); got != 37 {
		t.Fatalf("productSlugs() = %d, want 37", got)
	}
	if got := len(zoneSlugs(plan)); got != 30 {
		t.Fatalf("zoneSlugs() = %d, want 30", got)
	}
}
