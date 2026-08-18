package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"vantaca-interview-project/Demo/mock/northwind/internal/mockapi"
)

const defaultAPIKey = "northwind_mock_local_key"

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if len(os.Args) > 1 {
		if len(os.Args) == 2 && os.Args[1] == "healthcheck" {
			if err := runHealthcheck(); err != nil {
				logger.Error("healthcheck failed", "error", err)
				os.Exit(1)
			}
			return
		}

		logger.Error("unknown command")
		os.Exit(2)
	}

	config, address, err := loadConfig()
	if err != nil {
		logger.Error("invalid configuration", "error", err)
		os.Exit(1)
	}

	mockServer, err := mockapi.NewServer(config, logger)
	if err != nil {
		logger.Error("create mock server", "error", err)
		os.Exit(1)
	}

	httpServer := &http.Server{
		Addr:              address,
		Handler:           mockServer.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      20 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	runContext, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("northwind mock listening", "address", address)
		serverErrors <- httpServer.ListenAndServe()
	}()

	select {
	case <-runContext.Done():
		logger.Info("shutdown requested")
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server stopped unexpectedly", "error", err)
			os.Exit(1)
		}
		return
	}

	shutdownContext, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownContext); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}
}

func loadConfig() (mockapi.Config, string, error) {
	port, err := configuredPort()
	if err != nil {
		return mockapi.Config{}, "", err
	}

	scenarioDelay, err := durationEnv("NORTHWIND_MOCK_SCENARIO_DELAY", 5*time.Second)
	if err != nil {
		return mockapi.Config{}, "", err
	}
	webhookBackoff, err := durationEnv("NORTHWIND_MOCK_WEBHOOK_BACKOFF", 100*time.Millisecond)
	if err != nil {
		return mockapi.Config{}, "", err
	}
	webhookAttempts, err := intEnv("NORTHWIND_MOCK_WEBHOOK_ATTEMPTS", 3)
	if err != nil {
		return mockapi.Config{}, "", err
	}

	return mockapi.Config{
		APIKey:          envOrDefault("NORTHWIND_MOCK_API_KEY", defaultAPIKey),
		WebhookURL:      os.Getenv("NORTHWIND_MOCK_WEBHOOK_URL"),
		WebhookAttempts: webhookAttempts,
		WebhookBackoff:  webhookBackoff,
		ScenarioDelay:   scenarioDelay,
		SwaggerUIOrigin: envOrDefault("SWAGGER_UI_ORIGIN", "http://localhost:18090"),
	}, ":" + port, nil
}

func configuredPort() (string, error) {
	port := envOrDefault("NORTHWIND_MOCK_PORT", "8081")
	parsedPort, err := strconv.Atoi(port)
	if err != nil || parsedPort < 1 || parsedPort > 65535 {
		return "", errors.New("NORTHWIND_MOCK_PORT must be an integer from 1 through 65535")
	}
	return port, nil
}

func runHealthcheck() error {
	port, err := configuredPort()
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: 2 * time.Second}
	return checkHealth(client, "http://127.0.0.1:"+port+"/healthz")
}

func checkHealth(client *http.Client, endpoint string) error {
	request, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("create healthcheck request: %w", err)
	}

	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("request health endpoint: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("health endpoint returned HTTP %d", response.StatusCode)
	}
	return nil
}

func envOrDefault(name string, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

func durationEnv(name string, fallback time.Duration) (time.Duration, error) {
	value := os.Getenv(name)
	if value == "" {
		return fallback, nil
	}

	duration, err := time.ParseDuration(value)
	if err != nil || duration <= 0 {
		return 0, fmt.Errorf("%s must be a positive Go duration", name)
	}
	return duration, nil
}

func intEnv(name string, fallback int) (int, error) {
	value := os.Getenv(name)
	if value == "" {
		return fallback, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 1 {
		return 0, fmt.Errorf("%s must be a positive integer", name)
	}
	return parsed, nil
}
