package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

const DefaultTenantID = "018f0000-0000-7000-8000-000000000001"

// Config is shared by every thin-slice service. Services only read the values
// they need, but validation is common so a bad deployment fails fast.
type Config struct {
	Service            string
	HTTPAddr           string
	DatabaseURL        string
	KafkaBrokers       []string
	RedisAddr          string
	TenantID           string
	TenantSalt         string
	InternalToken      string
	RuleCatalogVersion string
	EngineVersion      string
	DocForensics       bool
	LogLevel           string
	IdentityURL        string
	IngestionURL       string
	AlertURL           string
	AuditURL           string
	CookieSecure       bool
}

func Load(service string) Config {
	return Config{
		Service:            service,
		HTTPAddr:           env("HTTP_ADDR", ":8080"),
		DatabaseURL:        env("DATABASE_URL", ""),
		KafkaBrokers:       splitCSV(env("KAFKA_BROKERS", "")),
		RedisAddr:          env("REDIS_ADDR", ""),
		TenantID:           env("TENANT_ID", DefaultTenantID),
		TenantSalt:         env("TENANT_SALT", ""),
		InternalToken:      env("INTERNAL_TOKEN", ""),
		RuleCatalogVersion: env("RULE_CATALOG_VERSION", "2026.11.1"),
		EngineVersion:      env("ENGINE_VERSION", "rulesengine-1.0.0"),
		DocForensics:       envBool("DOC_FORENSICS_ENABLED", false),
		LogLevel:           env("LOG_LEVEL", "info"),
		IdentityURL:        env("IDENTITY_URL", "http://identity:8080"),
		IngestionURL:       env("INGESTION_URL", "http://ingestion:8080"),
		AlertURL:           env("ALERT_URL", "http://alert:8080"),
		AuditURL:           env("AUDIT_URL", "http://audit:8080"),
		CookieSecure:       envBool("COOKIE_SECURE", false),
	}
}

func (c Config) Validate() error {
	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	if c.InternalToken == "" {
		return fmt.Errorf("INTERNAL_TOKEN is required")
	}
	if c.TenantSalt == "" {
		return fmt.Errorf("TENANT_SALT is required")
	}
	if _, err := uuid.Parse(c.TenantID); err != nil {
		return fmt.Errorf("TENANT_ID must be a uuid: %w", err)
	}
	return nil
}

func env(key, def string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return def
}

func envBool(key string, def bool) bool {
	v, ok := os.LookupEnv(key)
	if !ok {
		return def
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return def
	}
	return b
}

func splitCSV(v string) []string {
	if strings.TrimSpace(v) == "" {
		return nil
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if s := strings.TrimSpace(p); s != "" {
			out = append(out, s)
		}
	}
	return out
}
