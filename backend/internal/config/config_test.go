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
