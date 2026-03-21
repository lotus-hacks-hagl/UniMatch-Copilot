package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	// Load .env
	err := godotenv.Load()
	if err != nil {
		log.Println("Note: .env file not found, using system environment variables")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	// Connect to DB
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Read seed.sql
	content, err := os.ReadFile("migrations/002_seed.sql")
	if err != nil {
		log.Fatalf("Failed to read migrations/002_seed.sql: %v", err)
	}

	// Execute SQL
	_, err = db.Exec(string(content))
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value") {
			fmt.Println("Note: Some data might already exist. Seeding skipped for those records.")
		} else {
			log.Fatalf("Failed to execute seed SQL: %v", err)
		}
	}

	fmt.Println("Successfully seeded initial data from migrations/002_seed.sql")
}
