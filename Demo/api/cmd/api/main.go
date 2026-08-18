package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"vantaca-interview-project/Demo/api/internal/application"
	"vantaca-interview-project/Demo/api/internal/config"
	"vantaca-interview-project/Demo/api/internal/database"
	"vantaca-interview-project/Demo/api/internal/httpapi"
	appLogging "vantaca-interview-project/Demo/api/internal/logging"
	"vantaca-interview-project/Demo/api/internal/northwind"
)

func main() {
	consoleHandler := appLogging.NewRedactingHandler(slog.NewJSONHandler(os.Stdout, nil))
	startupLogger := slog.New(consoleHandler)
	logger := startupLogger
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		logger.Error("invalid configuration", "error", err)
		os.Exit(1)
	}

	if len(os.Args) > 1 {
		if len(os.Args) == 2 && os.Args[1] == "healthcheck" {
			if err := healthcheck(cfg.Port); err != nil {
				logger.Error("healthcheck failed", "error", err)
				os.Exit(1)
			}
			return
		}
		logger.Error("unknown command")
		os.Exit(2)
	}

	bootstrapCtx, bootstrapCancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer bootstrapCancel()
	db, err := database.Bootstrap(bootstrapCtx, cfg)
	if err != nil {
		logger.Error("bootstrap database", "error", err)
		os.Exit(1)
	}
	repository := database.NewRepository(db, cfg.DemoLinkID)
	defer repository.Close()

	databaseHandler := appLogging.NewDatabaseHandler(db, appLogging.DatabaseHandlerOptions{
		Application:  "vantaca-api",
		MinimumLevel: slog.LevelInfo,
		WriteTimeout: 500 * time.Millisecond,
		OnError: func(logErr error) {
			// This logger has no database sink, so reporting a persistence
			// failure cannot recurse back into the failed handler.
			startupLogger.Error("database log persistence failed", "error", logErr)
		},
	})
	logger = slog.New(appLogging.NewFanoutHandler(consoleHandler, databaseHandler))
	slog.SetDefault(logger)

	partner, err := northwind.NewClient(cfg.NorthwindBaseURL, cfg.NorthwindAPIKey, cfg.NorthwindTimeout)
	if err != nil {
		logger.Error("create Northwind client", "error", err)
		os.Exit(1)
	}
	syncService := application.NewSyncService(repository, partner, logger)
	coordinator := application.NewRefreshCoordinator(syncService, logger)
	transferService := application.NewTransferService(repository, partner, logger)

	initialSyncCtx, initialSyncCancel := context.WithTimeout(context.Background(), 30*time.Second)
	initialResult, initialErr := syncService.SyncAll(initialSyncCtx, cfg.DemoTenantID, "")
	initialSyncCancel()
	if initialErr != nil {
		accounts, readErr := repository.ListAccounts(context.Background(), cfg.DemoTenantID)
		if readErr != nil || len(accounts) == 0 {
			logger.Error("initial Northwind synchronization failed and no SQL snapshot exists", "error", initialErr)
			os.Exit(1)
		}
		logger.Warn("initial Northwind synchronization partially failed; serving last SQL snapshot", "error", initialErr)
	} else {
		logger.Info(
			"initial synchronization complete",
			"accounts", initialResult.AccountsSeen,
			"account_changes", initialResult.AccountsChanged,
			"transaction_changes", initialResult.TransactionChanges,
		)
	}

	api := httpapi.NewServer(httpapi.Options{
		TenantID:              cfg.DemoTenantID,
		CustomerID:            cfg.DemoCustomerID,
		FreshnessThreshold:    cfg.FreshnessThreshold,
		DemoControlsEnabled:   cfg.DemoControlsEnabled,
		AllowUnsignedWebhooks: cfg.DemoAllowUnsignedWeb,
		DemoAdminKey:          cfg.DemoAdminKey,
		SwaggerUIOrigin:       cfg.SwaggerUIOrigin,
	}, repository, repository, syncService, coordinator, transferService, logger)

	httpServer := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           api.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	runCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("Vantaca demo API listening", "port", cfg.Port)
		serverErrors <- httpServer.ListenAndServe()
	}()

	select {
	case <-runCtx.Done():
		logger.Info("shutdown requested")
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			logger.Error("API server stopped unexpectedly", "error", err)
			os.Exit(1)
		}
		return
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}
	coordinator.Wait()
}

func healthcheck(port string) error {
	client := &http.Client{Timeout: 2 * time.Second}
	response, err := client.Get("http://127.0.0.1:" + port + "/healthz")
	if err != nil {
		return fmt.Errorf("request health endpoint: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("health endpoint returned HTTP %d", response.StatusCode)
	}
	return nil
}
