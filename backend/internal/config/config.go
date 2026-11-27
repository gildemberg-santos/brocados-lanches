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

	OrderAccessSecret string

	AdminCookieName        string
	AdminCookieSecure      bool
	AdminBootstrapName     string
	AdminBootstrapEmail    string
	AdminBootstrapPassword string
}

// Load lê e valida a configuração a partir de getenv.
func Load(getenv func(string) string) (Config, error) {
	cfg := Config{
		Env:      value(getenv("APP_ENV"), "development"),
		Port:     value(getenv("PORT"), "8080"),
		MongoURI: mongoURI(getenv),
		MongoDB:  value(getenv("MONGO_DB"), "brocados_dev"),
	}

	cfg.OrderAccessSecret = value(getenv("ORDER_ACCESS_SECRET"), "")
	if cfg.OrderAccessSecret == "" {
		if cfg.Env == "production" {
			return Config{}, fmt.Errorf("ORDER_ACCESS_SECRET is required in production")
		}
		cfg.OrderAccessSecret = "brocados-development-order-access-secret"
	}

	cfg.AdminCookieName = value(getenv("ADMIN_SESSION_COOKIE"), "brocados_admin")
	cfg.AdminCookieSecure = cfg.Env == "production" ||
		strings.EqualFold(value(getenv("ADMIN_COOKIE_SECURE"), "false"), "true")
	cfg.AdminBootstrapName = value(getenv("ADMIN_BOOTSTRAP_NAME"), "")
	cfg.AdminBootstrapEmail = value(getenv("ADMIN_BOOTSTRAP_EMAIL"), "")
	cfg.AdminBootstrapPassword = getenv("ADMIN_BOOTSTRAP_PASSWORD")

	if err := validateEnv(cfg.Env); err != nil {
		return Config{}, err
	}
	if err := validatePort(cfg.Port); err != nil {
		return Config{}, err
	}
	if !isMongoURI(cfg.MongoURI) {
		return Config{}, fmt.Errorf("MONGO_URI or MONGODB_URI must start with mongodb:// or mongodb+srv://")
	}
	if strings.TrimSpace(cfg.MongoDB) == "" {
		return Config{}, fmt.Errorf("MONGO_DB is required")
	}
	if cfg.Env == "production" && len(cfg.OrderAccessSecret) < 32 {
		return Config{}, fmt.Errorf("ORDER_ACCESS_SECRET must have at least 32 characters in production")
	}
	if !validCookieName(cfg.AdminCookieName) {
		return Config{}, fmt.Errorf("ADMIN_SESSION_COOKIE has invalid characters")
	}
	if err := validateBootstrap(cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

// AdminCookieNameEffective devolve o nome final do cookie. Em produção o
// prefixo __Host- é obrigatório e exige Secure, Path=/ e ausência de Domain.
func (c Config) AdminCookieNameEffective() string {
	if c.AdminCookieSecure && !strings.HasPrefix(strings.ToLower(c.AdminCookieName), "__host-") {
		return "__Host-" + c.AdminCookieName
	}
	return c.AdminCookieName
}

// HasAdminBootstrap informa se todas as variáveis de bootstrap foram fornecidas.
func (c Config) HasAdminBootstrap() bool {
	return c.AdminBootstrapName != "" && c.AdminBootstrapEmail != "" && c.AdminBootstrapPassword != ""
}

// Addr devolve o endereço HTTP no formato :porta.
func (c Config) Addr() string {
	return net.JoinHostPort("", c.Port)
}

// mongoURI aceita MONGO_URI (nome canônico do projeto) e MONGODB_URI
// (nome usado por plataformas como Render e Atlas). MONGO_URI tem prioridade.
func mongoURI(getenv func(string) string) string {
	if uri := value(getenv("MONGO_URI"), ""); uri != "" {
		return uri
	}
	return value(getenv("MONGODB_URI"), "mongodb://localhost:27017")
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

func validCookieName(name string) bool {
	if name == "" {
		return false
	}
	for _, char := range name {
		switch {
		case char >= 'a' && char <= 'z', char >= 'A' && char <= 'Z', char >= '0' && char <= '9':
		case strings.ContainsRune("!#$%&'*+-.^_`|~", char):
		default:
			return false
		}
	}
	return true
}

func validateBootstrap(cfg Config) error {
	fields := []string{cfg.AdminBootstrapName, cfg.AdminBootstrapEmail, cfg.AdminBootstrapPassword}
	provided := 0
	for _, field := range fields {
		if field != "" {
			provided++
		}
	}
	if provided == 0 || provided == len(fields) {
		return nil
	}
	return fmt.Errorf("ADMIN_BOOTSTRAP_NAME, ADMIN_BOOTSTRAP_EMAIL and ADMIN_BOOTSTRAP_PASSWORD must be set together")
}
