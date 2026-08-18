package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	_ "github.com/microsoft/go-mssqldb"

	"vantaca-interview-project/Demo/api/internal/config"
)

var databaseNamePattern = regexp.MustCompile(`^[A-Za-z0-9_]+$`)

func Bootstrap(ctx context.Context, cfg config.Config) (*sql.DB, error) {
	if !databaseNamePattern.MatchString(cfg.DatabaseName) {
		return nil, errors.New("DB_NAME may contain only letters, digits, and underscore")
	}

	master, err := openWithRetry(ctx, connectionString(cfg, "master"))
	if err != nil {
		return nil, fmt.Errorf("connect to SQL Server master: %w", err)
	}

	createStatement := fmt.Sprintf(
		"IF DB_ID(N'%s') IS NULL CREATE DATABASE [%s]",
		cfg.DatabaseName,
		cfg.DatabaseName,
	)
	if _, err := master.ExecContext(ctx, createStatement); err != nil {
		_ = master.Close()
		return nil, fmt.Errorf("create application database: %w", err)
	}
	if err := master.Close(); err != nil {
		return nil, fmt.Errorf("close master database: %w", err)
	}

	db, err := openWithRetry(ctx, connectionString(cfg, cfg.DatabaseName))
	if err != nil {
		return nil, fmt.Errorf("connect to application database: %w", err)
	}
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(30 * time.Minute)

	if err := applyMigrations(ctx, db, cfg.MigrationsDir); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := applySeeds(ctx, db, cfg.SeedsDir); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}

func connectionString(cfg config.Config, databaseName string) string {
	query := url.Values{}
	query.Set("database", databaseName)
	query.Set("encrypt", "true")
	query.Set("TrustServerCertificate", "true")
	query.Set("connection timeout", "5")

	return (&url.URL{
		Scheme:   "sqlserver",
		User:     url.UserPassword(cfg.DatabaseUser, cfg.DatabasePassword),
		Host:     cfg.DatabaseHost + ":" + cfg.DatabasePort,
		RawQuery: query.Encode(),
	}).String()
}

func openWithRetry(ctx context.Context, dsn string) (*sql.DB, error) {
	var lastErr error
	for {
		db, err := sql.Open("sqlserver", dsn)
		if err == nil {
			pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
			err = db.PingContext(pingCtx)
			cancel()
			if err == nil {
				return db, nil
			}
			_ = db.Close()
		}
		lastErr = err

		timer := time.NewTimer(time.Second)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}
			return nil, errors.Join(lastErr, ctx.Err())
		case <-timer.C:
		}
	}
}

func applyMigrations(ctx context.Context, db *sql.DB, directory string) error {
	if _, err := db.ExecContext(ctx, `
IF OBJECT_ID('dbo.schema_migrations', 'U') IS NULL
BEGIN
    CREATE TABLE dbo.schema_migrations (
        migration_name NVARCHAR(255) NOT NULL PRIMARY KEY,
        applied_at DATETIMEOFFSET(7) NOT NULL DEFAULT SYSUTCDATETIME()
    );
END`); err != nil {
		return fmt.Errorf("create migration ledger: %w", err)
	}

	files, err := sqlFiles(directory)
	if err != nil {
		return fmt.Errorf("list migrations: %w", err)
	}
	for _, file := range files {
		name := filepath.Base(file)
		var count int
		if err := db.QueryRowContext(ctx,
			"SELECT COUNT(1) FROM dbo.schema_migrations WHERE migration_name = @name",
			sql.Named("name", name),
		).Scan(&count); err != nil {
			return fmt.Errorf("check migration %s: %w", name, err)
		}
		if count > 0 {
			continue
		}

		script, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("begin migration %s: %w", name, err)
		}
		if _, err := tx.ExecContext(ctx, string(script)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("execute migration %s: %w", name, err)
		}
		if _, err := tx.ExecContext(ctx,
			"INSERT INTO dbo.schema_migrations (migration_name) VALUES (@name)",
			sql.Named("name", name),
		); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("record migration %s: %w", name, err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %s: %w", name, err)
		}
	}
	return nil
}

func applySeeds(ctx context.Context, db *sql.DB, directory string) error {
	files, err := sqlFiles(directory)
	if err != nil {
		return fmt.Errorf("list seeds: %w", err)
	}
	for _, file := range files {
		script, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("read seed %s: %w", filepath.Base(file), err)
		}
		if _, err := db.ExecContext(ctx, string(script)); err != nil {
			return fmt.Errorf("apply seed %s: %w", filepath.Base(file), err)
		}
	}
	return nil
}

func sqlFiles(directory string) ([]string, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, err
	}
	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".sql") {
			continue
		}
		files = append(files, filepath.Join(directory, entry.Name()))
	}
	sort.Strings(files)
	if len(files) == 0 {
		return nil, fs.ErrNotExist
	}
	return files, nil
}
