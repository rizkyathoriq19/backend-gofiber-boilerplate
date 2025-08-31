package main

import (
	"database/sql"
	"log"

	"boilerplate-be/internal/infrastructure/config"
	"boilerplate-be/internal/infrastructure/database"
	"boilerplate-be/internal/infrastructure/enum"
	"boilerplate-be/internal/infrastructure/helper"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	cfg := config.New()

	// Connect to database
	db, err := database.New(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("🌱 Starting database seeder...")

	// Seed users (matching your entity structure)
	if err := seedUsers(db); err != nil {
		log.Fatalf("Failed to seed users: %v", err)
	}

	log.Println("✅ Database seeding completed successfully!")
}

func seedUsers(db *sql.DB) error {
	log.Println("Seeding users...")

	users := []struct {
		name     string
		email    string
		password string
		role     enum.UserRole
	}{
		{
			name:     "Super Admin",
			email:    "admin@example.com",
			password: "admin123",
			role:     enum.UserRoleAdmin,
		},
		{
			name:     "John Doe",
			email:    "john@example.com",
			password: "password123",
			role:     enum.UserRoleUser,
		},
		{
			name:     "Jane Smith",
			email:    "jane@example.com",
			password: "password123",
			role:     enum.UserRoleUser,
		},
		{
			name:     "Bob Johnson",
			email:    "bob@example.com",
			password: "password123",
			role:     enum.UserRoleUser,
		},
		{
			name:     "Alice Wilson",
			email:    "alice@example.com",
			password: "password123",
			role:     enum.UserRoleUser,
		},
	}

	query := `
		INSERT INTO users (name, email, password, role, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		ON CONFLICT (email) DO UPDATE SET
			name = EXCLUDED.name,
			updated_at = NOW()
		RETURNING id, email
	`

	for _, userData := range users {
		// Hash password
		hashedPassword, err := helper.HashPassword(userData.password)
		if err != nil {
			return err
		}

		// Insert/Update user
		var userID, email string
		err = db.QueryRow(query, userData.name, userData.email, hashedPassword, userData.role).Scan(&userID, &email)
		if err != nil {
			return err
		}

		log.Printf("✅ User %s (%s) seeded with ID: %s", email, userData.role, userID)
	}

	return nil
}