package config

import "testing"

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
}

func TestLoadOverrides(t *testing.T) {
	env := map[string]string{
		"APP_ENV":   "production",
		"PORT":      "9090",
		"MONGO_URI": "mongodb+srv://cluster.example.com",
		"MONGO_DB":  "brocados",
	}

	cfg, err := Load(func(key string) string { return env[key] })
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Env != "production" || cfg.Port != "9090" {
		t.Errorf("unexpected env/port: %+v", cfg)
	}
	if cfg.Addr() != ":9090" {
		t.Errorf("Addr() = %q, want :9090", cfg.Addr())
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
		{name: "partial bootstrap", env: map[string]string{"ADMIN_BOOTSTRAP_EMAIL": "dono@example.com"}},
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
