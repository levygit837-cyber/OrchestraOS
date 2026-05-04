package migrations

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pressly/goose/v3"
)

// Run runs all pending migrations
func Run(db *sql.DB) error {
	dir, err := migrationsDir()
	if err != nil {
		return err
	}

	if err := configureGoose(); err != nil {
		return err
	}

	if err := goose.Up(db, dir); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// RunTo runs migrations up to a specific version
func RunTo(db *sql.DB, version int64) error {
	dir, err := migrationsDir()
	if err != nil {
		return err
	}

	if err := configureGoose(); err != nil {
		return err
	}

	if err := goose.UpTo(db, dir, version); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// Status returns the current migration status
func Status(db *sql.DB) error {
	dir, err := migrationsDir()
	if err != nil {
		return err
	}

	if err := configureGoose(); err != nil {
		return err
	}

	return goose.Status(db, dir)
}

// Reset rolls back all migrations
func Reset(db *sql.DB) error {
	dir, err := migrationsDir()
	if err != nil {
		return err
	}

	if err := configureGoose(); err != nil {
		return err
	}

	return goose.Reset(db, dir)
}

func configureGoose() error {
	goose.SetBaseFS(nil)
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set dialect: %w", err)
	}
	return nil
}

func migrationsDir() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	start := wd
	for {
		candidate := filepath.Join(wd, "migrations")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate, nil
		}

		parent := filepath.Dir(wd)
		if parent == wd {
			return "", fmt.Errorf("migrations directory not found from %s", start)
		}
		wd = parent
	}
}
