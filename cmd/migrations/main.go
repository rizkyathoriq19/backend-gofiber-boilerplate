package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"boilerplate-be/internal/infrastructure/config"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	cfg := config.New()

	var direction = flag.String("direction", "up", "Migration direction: up or down")
	var steps = flag.Int("steps", 0, "Number of migration steps (0 = all)")
	var create = flag.String("create", "", "Create new migration with given name")
	flag.Parse()

	// Handle create migration
	if *create != "" {
		if err := createMigration(*create); err != nil {
			log.Fatalf("Failed to create migration: %v", err)
		}
		return
	}

	// Connect to database
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize migration driver
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatalf("Failed to create migration driver: %v", err)
	}

	// Create migrate instance
	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres",
		driver,
	)
	if err != nil {
		log.Fatalf("Failed to create migrate instance: %v", err)
	}

	// Run migrations
	switch *direction {
	case "up":
		if *steps == 0 {
			if err := m.Up(); err != nil && err != migrate.ErrNoChange {
				log.Fatalf("Migration up failed: %v", err)
			}
		} else {
			if err := m.Steps(*steps); err != nil {
				log.Fatalf("Migration steps failed: %v", err)
			}
		}
		log.Println("Migration up completed successfully")
	case "down":
		if *steps == 0 {
			if err := m.Down(); err != nil && err != migrate.ErrNoChange {
				log.Fatalf("Migration down failed: %v", err)
			}
		} else {
			if err := m.Steps(-*steps); err != nil {
				log.Fatalf("Migration steps failed: %v", err)
			}
		}
		log.Println("Migration down completed successfully")
	case "force":
		version := *steps
		if version == 0 {
			log.Fatalf("Force requires a version number. Use -steps=<version>")
		}
		if err := m.Force(version); err != nil {
			log.Fatalf("Migration force failed: %v", err)
		}
		log.Printf("Migration forced to version %d", version)
	case "version":
		version, dirty, err := m.Version()
		if err != nil {
			log.Fatalf("Failed to get migration version: %v", err)
		}
		log.Printf("Current migration version: %d, dirty: %v", version, dirty)
	default:
		log.Fatalf("Invalid direction: %s. Use 'up', 'down', 'force', or 'version'", *direction)
	}
}

func createMigration(name string) error {
	// Get migrations directory
	migrationsDir := "migrations"
	
	// Create migrations directory if it doesn't exist
	if err := os.MkdirAll(migrationsDir, 0755); err != nil {
		return fmt.Errorf("failed to create migrations directory: %w", err)
	}

	// Get next sequence number
	nextSeq, err := getNextSequenceNumber(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to get next sequence number: %w", err)
	}

	// Generate timestamp
	timestamp := time.Now().Format("20060102150405")
	
	// Clean migration name
	cleanName := strings.ReplaceAll(strings.ToLower(name), " ", "_")
	
	// Create file names
	upFile := fmt.Sprintf("%s/%03d_%s_%s.up.sql", migrationsDir, nextSeq, timestamp, cleanName)
	downFile := fmt.Sprintf("%s/%03d_%s_%s.down.sql", migrationsDir, nextSeq, timestamp, cleanName)

	// Create UP migration file
	upTemplate := generateUpTemplate(cleanName)
	if err := os.WriteFile(upFile, []byte(upTemplate), 0644); err != nil {
		return fmt.Errorf("failed to create up migration file: %w", err)
	}

	// Create DOWN migration file
	downTemplate := generateDownTemplate(cleanName)
	if err := os.WriteFile(downFile, []byte(downTemplate), 0644); err != nil {
		return fmt.Errorf("failed to create down migration file: %w", err)
	}

	fmt.Printf("Migration created successfully:\n")
	fmt.Printf("  UP:   %s\n", upFile)
	fmt.Printf("  DOWN: %s\n", downFile)
	
	return nil
}

func getNextSequenceNumber(migrationsDir string) (int, error) {
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		// If directory doesn't exist, start with 1
		if os.IsNotExist(err) {
			return 1, nil
		}
		return 0, err
	}

	maxSeq := 0
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		
		name := file.Name()
		if strings.HasSuffix(name, ".sql") {
			// Extract sequence number from filename (first 3 digits)
			if len(name) >= 3 {
				seqStr := name[:3]
				if seq, err := strconv.Atoi(seqStr); err == nil && seq > maxSeq {
					maxSeq = seq
				}
			}
		}
	}

	return maxSeq + 1, nil
}

func generateUpTemplate(name string) string {
	template := fmt.Sprintf(`-- Migration: %s
-- Created at: %s

-- Add your UP migration SQL here
-- Example:
-- CREATE TABLE example_table (
--     id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
--     name VARCHAR(255) NOT NULL,
--     created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
-- );

`, strings.ReplaceAll(name, "_", " "), time.Now().Format("2006-01-02 15:04:05"))

	// Add common templates based on migration name
	if strings.Contains(name, "create") && strings.Contains(name, "table") {
		tableName := extractTableName(name)
		template += fmt.Sprintf(`-- CREATE TABLE %s (
--     id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
--     created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
--     updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
-- );

`, tableName)
	} else if strings.Contains(name, "add") && strings.Contains(name, "column") {
		template += `-- ALTER TABLE table_name 
-- ADD COLUMN column_name data_type;

`
	} else if strings.Contains(name, "index") {
		template += `-- CREATE INDEX idx_table_column ON table_name(column_name);

`
	}

	return template
}

func generateDownTemplate(name string) string {
	template := fmt.Sprintf(`-- Rollback migration: %s
-- Created at: %s

-- Add your DOWN migration SQL here (reverse of UP)
-- Example:
-- DROP TABLE IF EXISTS example_table;

`, strings.ReplaceAll(name, "_", " "), time.Now().Format("2006-01-02 15:04:05"))

	// Add common rollback templates
	if strings.Contains(name, "create") && strings.Contains(name, "table") {
		tableName := extractTableName(name)
		template += fmt.Sprintf(`-- DROP TABLE IF EXISTS %s;

`, tableName)
	} else if strings.Contains(name, "add") && strings.Contains(name, "column") {
		template += `-- ALTER TABLE table_name 
-- DROP COLUMN IF EXISTS column_name;

`
	} else if strings.Contains(name, "index") {
		template += `-- DROP INDEX IF EXISTS idx_table_column;

`
	}

	return template
}

func extractTableName(migrationName string) string {
	parts := strings.Split(migrationName, "_")
	for i, part := range parts {
		if part == "table" && i > 0 {
			return parts[i-1]
		}
	}
	return "table_name"
}