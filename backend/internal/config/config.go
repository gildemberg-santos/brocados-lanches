package config

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

// Config reúne as configurações lidas do ambiente.
type Config struct {
	Env      string
	Port     string
	MongoURI string
	MongoDB  string
}

// Load lê e valida a configuração a partir de getenv.
func Load(getenv func(string) string) (Config, error) {
	cfg := Config{
		Env:      value(getenv("APP_ENV"), "development"),
		Port:     value(getenv("PORT"), "8080"),
		MongoURI: value(getenv("MONGO_URI"), "mongodb://localhost:27017"),
		MongoDB:  value(getenv("MONGO_DB"), "brocados_dev"),
	}

	if err := validateEnv(cfg.Env); err != nil {
		return Config{}, err
	}
	if err := validatePort(cfg.Port); err != nil {
		return Config{}, err
	}
	if !isMongoURI(cfg.MongoURI) {
		return Config{}, fmt.Errorf("MONGO_URI must start with mongodb:// or mongodb+srv://")
	}
	if strings.TrimSpace(cfg.MongoDB) == "" {
		return Config{}, fmt.Errorf("MONGO_DB is required")
	}

	return cfg, nil
}

// Addr devolve o endereço HTTP no formato :porta.
func (c Config) Addr() string {
	return net.JoinHostPort("", c.Port)
}

func value(raw, fallback string) string {
	if strings.TrimSpace(raw) == "" {
		return fallback
	}
	return strings.TrimSpace(raw)
}

func validateEnv(env string) error {
	switch env {
	case "development", "test", "production":
		return nil
	default:
		return fmt.Errorf("APP_ENV must be development, test or production")
	}
}

func validatePort(port string) error {
	number, err := strconv.Atoi(port)
	if err != nil || number < 1 || number > 65535 {
		return fmt.Errorf("PORT must be a number between 1 and 65535")
	}
	return nil
}

func isMongoURI(uri string) bool {
	return strings.HasPrefix(uri, "mongodb://") || strings.HasPrefix(uri, "mongodb+srv://")
}
