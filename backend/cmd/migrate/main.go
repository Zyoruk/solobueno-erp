package main

import (
	"fmt"
	"os"

	"github.com/solobueno/erp/internal/auth"
	"github.com/solobueno/erp/internal/shared/database"
)

func main() {
	fmt.Println("Solobueno ERP Migration Tool")

	if len(os.Args) < 2 {
		fmt.Println("Usage: migrate [up|down|status]")
		fmt.Println("Commands:")
		fmt.Println("  up      - Run all migrations (GORM AutoMigrate)")
		fmt.Println("  down    - Drop all tables (DANGEROUS)")
		fmt.Println("  status  - Show migration status")
		os.Exit(1)
	}

	command := os.Args[1]

	// Connect to database
	cfg := database.DefaultConfig()
	db, err := database.NewConnection(cfg)
	if err != nil {
		fmt.Printf("Failed to connect to database: %v\n", err)
		os.Exit(1)
	}

	switch command {
	case "up":
		fmt.Println("Running migrations...")
		if err := auth.AutoMigrate(db); err != nil {
			fmt.Printf("Migration failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Migrations completed successfully!")

	case "down":
		fmt.Println("WARNING: This will drop all auth tables!")
		fmt.Print("Type 'yes' to confirm: ")
		var confirm string
		fmt.Scanln(&confirm)
		if confirm != "yes" {
			fmt.Println("Aborted.")
			os.Exit(0)
		}
		if err := auth.DropAll(db); err != nil {
			fmt.Printf("Drop failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("All auth tables dropped.")

	case "status":
		fmt.Println("Checking migration status...")
		sqlDB, err := db.DB()
		if err != nil {
			fmt.Printf("Failed to get SQL DB: %v\n", err)
			os.Exit(1)
		}
		if err := sqlDB.Ping(); err != nil {
			fmt.Printf("Database connection failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Database connection: OK")

		// Check if tables exist
		tables := []string{"users", "tenants", "user_tenant_roles", "sessions", "password_reset_tokens", "auth_events"}
		for _, table := range tables {
			if db.Migrator().HasTable(table) {
				fmt.Printf("  Table '%s': EXISTS\n", table)
			} else {
				fmt.Printf("  Table '%s': MISSING\n", table)
			}
		}

	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}
