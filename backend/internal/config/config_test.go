package config

import (
	"testing"
	"time"
)

func TestLoadDefaults(t *testing.T) {
	cfg, err := Load(func(string) string { return "" })
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Env != "development" {
		t.Errorf("Env = %q, want development", cfg.Env)
	}
	if cfg.Port != "8080" {
		t.Errorf("Port = %q, want 8080", cfg.Port)
	}
	if cfg.MongoURI != "mongodb://localhost:27017" {
		t.Errorf("MongoURI = %q, want default", cfg.MongoURI)
	}
	if cfg.MongoDB != "brocados_dev" {
		t.Errorf("MongoDB = %q, want brocados_dev", cfg.MongoDB)
	}
	if cfg.Addr() != ":8080" {
		t.Errorf("Addr() = %q, want :8080", cfg.Addr())
	}
	if cfg.OrderAccessSecret == "" {
		t.Error("OrderAccessSecret is empty in development")
	}
	if cfg.AdminBaseURL != "http://localhost:5173" {
		t.Errorf("AdminBaseURL = %q, want Vite development URL", cfg.AdminBaseURL)
	}
	if cfg.AdminCookieName != "brocados_admin" {
		t.Errorf("AdminCookieName = %q, want brocados_admin", cfg.AdminCookieName)
	}
	if cfg.AdminCookieSecure {
		t.Error("AdminCookieSecure = true in development, want false")
	}
	if cfg.AdminCookieNameEffective() != "brocados_admin" {
		t.Errorf("AdminCookieNameEffective() = %q", cfg.AdminCookieNameEffective())
	}
	if cfg.HasAdminBootstrap() {
		t.Error("HasAdminBootstrap() = true without environment variables")
	}
	if cfg.SentryDSN != "" {
		t.Errorf("SentryDSN = %q, want empty (disabled)", cfg.SentryDSN)
	}
	if cfg.SentryEnvironment != "development" {
		t.Errorf("SentryEnvironment = %q, want development", cfg.SentryEnvironment)
	}
	if cfg.SentryFlushTimeout != 2*time.Second {
		t.Errorf("SentryFlushTimeout = %s, want 2s", cfg.SentryFlushTimeout)
	}
}

func TestLoadOverrides(t *testing.T) {
	env := map[string]string{
		"APP_ENV":                      "production",
		"PORT":                         "9090",
		"MONGO_URI":                    "mongodb+srv://cluster.example.com",
		"MONGO_DB":                     "brocados",
		"ORDER_ACCESS_SECRET":          "segredo-de-producao-com-mais-de-32-caracteres",
		"ADMIN_BASE_URL":               "https://admin.exemplo.com",
		"SENTRY_DSN":                   "https://key@sentry.example.com/1",
		"SENTRY_ENVIRONMENT":           "producao",
		"SENTRY_RELEASE":               "abc123",
		"SENTRY_FLUSH_TIMEOUT_SECONDS": "5",
	}

	cfg, err := Load(func(key string) string { return env[key] })
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Env != "production" || cfg.Port != "9090" {
		t.Errorf("unexpected env/port: %+v", cfg)
	}
	if cfg.AdminBaseURL != "https://admin.exemplo.com" {
		t.Errorf("AdminBaseURL = %q, want configured URL", cfg.AdminBaseURL)
	}
	if cfg.Addr() != ":9090" {
		t.Errorf("Addr() = %q, want :9090", cfg.Addr())
	}
	if cfg.SentryDSN == "" || cfg.SentryEnvironment != "producao" || cfg.SentryRelease != "abc123" {
		t.Errorf("unexpected sentry config: %+v", cfg)
	}
	if cfg.SentryFlushTimeout != 5*time.Second {
		t.Errorf("SentryFlushTimeout = %s, want 5s", cfg.SentryFlushTimeout)
	}
}

func TestLoadAcceptsMongoDBURIAlias(t *testing.T) {
	cfg, err := Load(func(key string) string {
		if key == "MONGODB_URI" {
			return "mongodb+srv://user:pass@cluster.example.com"
		}
		return ""
	})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.MongoURI != "mongodb+srv://user:pass@cluster.example.com" {
		t.Errorf("MongoURI = %q, want alias value", cfg.MongoURI)
	}
}

func TestLoadPrefersMongoURIToAlias(t *testing.T) {
	env := map[string]string{
		"MONGO_URI":   "mongodb://canonical:27017",
		"MONGODB_URI": "mongodb://alias:27017",
		"MONGO_DB":    "brocados",
	}
	cfg, err := Load(func(key string) string { return env[key] })
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.MongoURI != "mongodb://canonical:27017" {
		t.Errorf("MongoURI = %q, want MONGO_URI to take precedence", cfg.MongoURI)
	}
}

func TestLoadRejectsInvalidValues(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
	}{
		{name: "invalid env", env: map[string]string{"APP_ENV": "staging"}},
		{name: "port not numeric", env: map[string]string{"PORT": "abc"}},
		{name: "port zero", env: map[string]string{"PORT": "0"}},
		{name: "port too high", env: map[string]string{"PORT": "65536"}},
		{name: "invalid mongo uri", env: map[string]string{"MONGO_URI": "http://localhost:27017"}},
		{name: "invalid mongodb uri alias", env: map[string]string{"MONGODB_URI": "http://localhost:27017"}},
		{name: "invalid cookie name", env: map[string]string{"ADMIN_SESSION_COOKIE": "cookie com espaco"}},
		{name: "admin base url without scheme", env: map[string]string{"ADMIN_BASE_URL": "admin.exemplo.com"}},
		{name: "admin base url with unsupported scheme", env: map[string]string{"ADMIN_BASE_URL": "ftp://admin.exemplo.com"}},
		{
			name: "production admin base url without https",
			env: map[string]string{
				"APP_ENV":             "production",
				"ORDER_ACCESS_SECRET": "segredo-de-producao-com-mais-de-32-caracteres",
				"ADMIN_BASE_URL":      "http://admin.exemplo.com",
			},
		},
		{name: "partial bootstrap", env: map[string]string{"ADMIN_BOOTSTRAP_EMAIL": "dono@example.com"}},
		{name: "invalid sentry flush timeout", env: map[string]string{"SENTRY_FLUSH_TIMEOUT_SECONDS": "abc"}},
		{name: "sentry flush timeout too high", env: map[string]string{"SENTRY_FLUSH_TIMEOUT_SECONDS": "31"}},
		{
			name: "missing production order access secret",
			env:  map[string]string{"APP_ENV": "production"},
		},
		{
			name: "short production order access secret",
			env: map[string]string{
				"APP_ENV":             "production",
				"ORDER_ACCESS_SECRET": "curto",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := Load(func(key string) string { return test.env[key] })
			if err == nil {
				t.Fatal("Load() error = nil, want error")
			}
		})
	}
}

func TestLoadAdminCookieInProduction(t *testing.T) {
	cfg, err := Load(func(key string) string {
		if key == "APP_ENV" {
			return "production"
		}
		if key == "ORDER_ACCESS_SECRET" {
			return "segredo-de-producao-com-mais-de-32-caracteres"
		}
		return ""
	})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if !cfg.AdminCookieSecure {
		t.Error("AdminCookieSecure = false in production, want true")
	}
	if cfg.AdminCookieNameEffective() != "__Host-brocados_admin" {
		t.Errorf("AdminCookieNameEffective() = %q, want __Host-brocados_admin", cfg.AdminCookieNameEffective())
	}
	if cfg.AdminBaseURL != "" {
		t.Errorf("AdminBaseURL = %q, want empty until the production domain is defined", cfg.AdminBaseURL)
	}
}

func TestLoadAdminBootstrap(t *testing.T) {
	env := map[string]string{
		"ADMIN_BOOTSTRAP_NAME":     "Dono",
		"ADMIN_BOOTSTRAP_EMAIL":    "dono@example.com",
		"ADMIN_BOOTSTRAP_PASSWORD": "senha-forte-com-12-caracteres",
	}
	cfg, err := Load(func(key string) string { return env[key] })
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if !cfg.HasAdminBootstrap() {
		t.Error("HasAdminBootstrap() = false, want true")
	}
}
