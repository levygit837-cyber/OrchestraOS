package cmd

import (
	"fmt"

	db "github.com/levygit837-cyber/OrchestraOS/internal/core/db"
	"github.com/pressly/goose/v3"
	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Database migrations",
	Long:  `Run database migrations using goose.`,
}

var migrateUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Run all pending migrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		dbConn, err := db.Open(getDBConfig())
		if err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}
		defer func() { _ = dbConn.Close() }()

		goose.SetBaseFS(nil)
		if err := goose.SetDialect("postgres"); err != nil {
			return fmt.Errorf("failed to set dialect: %w", err)
		}

		if err := goose.Up(dbConn, "migrations"); err != nil {
			return fmt.Errorf("failed to run migrations: %w", err)
		}

		fmt.Println("Migrations completed successfully")
		return nil
	},
}

var migrateStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show migration status",
	RunE: func(cmd *cobra.Command, args []string) error {
		dbConn, err := db.Open(getDBConfig())
		if err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}
		defer func() { _ = dbConn.Close() }()

		goose.SetBaseFS(nil)
		if err := goose.SetDialect("postgres"); err != nil {
			return fmt.Errorf("failed to set dialect: %w", err)
		}

		return goose.Status(dbConn, "migrations")
	},
}

var migrateResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset all migrations (WARNING: destructive)",
	RunE: func(cmd *cobra.Command, args []string) error {
		dbConn, err := db.Open(getDBConfig())
		if err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}
		defer func() { _ = dbConn.Close() }()

		goose.SetBaseFS(nil)
		if err := goose.SetDialect("postgres"); err != nil {
			return fmt.Errorf("failed to set dialect: %w", err)
		}

		if err := goose.Reset(dbConn, "migrations"); err != nil {
			return fmt.Errorf("failed to reset migrations: %w", err)
		}

		fmt.Println("Migrations reset successfully")
		return nil
	},
}

func getDBConfig() db.Config {
	host, _ := rootCmd.Flags().GetString("db-host")
	port, _ := rootCmd.Flags().GetString("db-port")
	user, _ := rootCmd.Flags().GetString("db-user")
	password, _ := rootCmd.Flags().GetString("db-password")
	dbname, _ := rootCmd.Flags().GetString("db-name")

	return db.Config{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		Database: dbname,
		SSLMode:  "disable",
	}
}

func init() {
	migrateCmd.AddCommand(migrateUpCmd)
	migrateCmd.AddCommand(migrateStatusCmd)
	migrateCmd.AddCommand(migrateResetCmd)
}
