package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: migrate <up|down>")
		os.Exit(1)
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost port=5432 user=netmon password=netmon_dev_password dbname=netmon sslmode=disable"
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to ping database: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()

	switch os.Args[1] {
	case "up":
		if err := runMigrations(ctx, db); err != nil {
			fmt.Fprintf(os.Stderr, "Migration failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Migrations completed successfully")
	case "down":
		fmt.Println("Down migrations not implemented yet")
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}

func runMigrations(ctx context.Context, db *sql.DB) error {
	// Read migration file
	migrationPath := os.Getenv("MIGRATION_PATH")
	if migrationPath == "" {
		migrationPath = "../../migrations/001_initial_schema.sql"
	}
	migrationSQL, err := os.ReadFile(migrationPath)
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	// Execute migration
	_, err = db.ExecContext(ctx, string(migrationSQL))
	if err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	return nil
}
