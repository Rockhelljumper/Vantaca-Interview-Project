package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Port                 string
	DatabaseHost         string
	DatabasePort         string
	DatabaseUser         string
	DatabasePassword     string
	DatabaseName         string
	MigrationsDir        string
	SeedsDir             string
	NorthwindBaseURL     string
	NorthwindAPIKey      string
	NorthwindTimeout     time.Duration
	FreshnessThreshold   time.Duration
	DemoTenantID         string
	DemoCustomerID       string
	DemoLinkID           string
	DemoControlsEnabled  bool
	DemoAllowUnsignedWeb bool
	DemoAdminKey         string
	SwaggerUIOrigin      string
}

func Load() (Config, error) {
	cfg := Config{
		Port:             valueOrDefault("API_PORT", "8080"),
		DatabaseHost:     valueOrDefault("DB_HOST", "localhost"),
		DatabasePort:     valueOrDefault("DB_PORT", "1433"),
		DatabaseUser:     valueOrDefault("DB_USER", "sa"),
		DatabasePassword: os.Getenv("DB_PASSWORD"),
		DatabaseName:     valueOrDefault("DB_NAME", "VantacaNorthwindDemo"),
		MigrationsDir:    valueOrDefault("MIGRATIONS_DIR", "../database/migrations"),
		SeedsDir:         valueOrDefault("SEEDS_DIR", "../database/seeds"),
		NorthwindBaseURL: strings.TrimRight(valueOrDefault("NORTHWIND_BASE_URL", "http://localhost:8081/v1"), "/"),
		NorthwindAPIKey:  valueOrDefault("NORTHWIND_API_KEY", "northwind_mock_local_key"),
		DemoTenantID:     valueOrDefault("DEMO_TENANT_ID", "tenant_demo"),
		DemoCustomerID:   valueOrDefault("DEMO_CUSTOMER_ID", "customer_demo"),
		DemoLinkID:       valueOrDefault("DEMO_LINK_ID", "11111111-1111-4111-8111-111111111111"),
		DemoAdminKey:     valueOrDefault("DEMO_ADMIN_KEY", "demo_admin_local_only"),
		SwaggerUIOrigin:  valueOrDefault("SWAGGER_UI_ORIGIN", "http://localhost:18090"),
	}

	var err error
	cfg.NorthwindTimeout, err = durationValue("NORTHWIND_TIMEOUT", 2*time.Second)
	if err != nil {
		return Config{}, err
	}
	cfg.FreshnessThreshold, err = durationValue("FRESHNESS_THRESHOLD", 5*time.Minute)
	if err != nil {
		return Config{}, err
	}
	cfg.DemoControlsEnabled, err = boolValue("DEMO_CONTROLS_ENABLED", true)
	if err != nil {
		return Config{}, err
	}
	cfg.DemoAllowUnsignedWeb, err = boolValue("DEMO_ALLOW_UNSIGNED_WEBHOOKS", true)
	if err != nil {
		return Config{}, err
	}

	if cfg.DatabasePassword == "" {
		return Config{}, errors.New("DB_PASSWORD is required")
	}
	if cfg.NorthwindAPIKey == "" {
		return Config{}, errors.New("NORTHWIND_API_KEY is required")
	}
	if cfg.DemoTenantID == "" || cfg.DemoCustomerID == "" || cfg.DemoLinkID == "" {
		return Config{}, errors.New("demo tenant, customer, and link identifiers are required")
	}

	return cfg, nil
}

func valueOrDefault(name string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}
	return value
}

func durationValue(name string, fallback time.Duration) (time.Duration, error) {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return fallback, nil
	}
	value, err := time.ParseDuration(raw)
	if err != nil || value <= 0 {
		return 0, fmt.Errorf("%s must be a positive Go duration", name)
	}
	return value, nil
}

func boolValue(name string, fallback bool) (bool, error) {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.ParseBool(raw)
	if err != nil {
		return false, fmt.Errorf("%s must be true or false", name)
	}
	return value, nil
}
