package cmd

import (
	"database/sql"
	"fmt"
	"os"

	db "github.com/levygit837-cyber/OrchestraOS/internal/core/db"
	"github.com/spf13/cobra"
)

var (
	// Global database connection (initialized on first use)
	database *sql.DB

	// rootCmd represents the base command when called without any subcommands
	rootCmd *cobra.Command
)

func init() {
	rootCmd = &cobra.Command{
		Use:   "orchestraos",
		Short: "OrchestraOS - Sistema de Orquestração de Agentes",
		//nolint:misspell // Portuguese word
		Long: `OrchestraOS é um sistema operacional de projeto onde agentes entendem contexto,
propõem próximos passos, executam tarefas, registram decisões e operam rotinas
com supervisão calibrada.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Skip DB init for certain commands
			if cmd.Name() == "version" || cmd.Name() == "help" {
				return nil
			}
			return initDB()
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			if database != nil {
				return database.Close()
			}
			return nil
		},
	}

	// Global flags
	rootCmd.PersistentFlags().String("db-host", getEnv("DB_HOST", "localhost"), "Database host")
	rootCmd.PersistentFlags().String("db-port", getEnv("DB_PORT", "5432"), "Database port")
	rootCmd.PersistentFlags().String("db-user", getEnv("DB_USER", "orchestraos"), "Database user")
	rootCmd.PersistentFlags().String("db-password", getEnv("DB_PASSWORD", "orchestraos"), "Database password")
	rootCmd.PersistentFlags().String("db-name", getEnv("DB_NAME", "orchestraos"), "Database name")
	// Add subcommands
	rootCmd.AddCommand(taskCmd)
	rootCmd.AddCommand(workUnitCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(eventCmd)
	rootCmd.AddCommand(agentSessionCmd)
	rootCmd.AddCommand(migrateCmd)
	rootCmd.AddCommand(versionCmd)
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func initDB() error {
	if database != nil {
		return nil
	}

	host, _ := rootCmd.Flags().GetString("db-host")
	port, _ := rootCmd.Flags().GetString("db-port")
	user, _ := rootCmd.Flags().GetString("db-user")
	password, _ := rootCmd.Flags().GetString("db-password")
	dbname, _ := rootCmd.Flags().GetString("db-name")

	cfg := db.Config{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		Database: dbname,
		SSLMode:  "disable",
	}

	conn, err := db.Open(cfg)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	database = conn
	return nil
}

func getDB() *sql.DB {
	return database
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// versionCmd
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("OrchestraOS CLI v0.1.0")
	},
}
