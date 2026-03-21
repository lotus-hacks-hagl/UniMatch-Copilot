package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"unimatch-be/internal/model"
)

func main() {
	_ = godotenv.Load()
	dbURL := os.Getenv("DATABASE_URL")
	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	var users []model.User
	db.Find(&users)
	for _, u := range users {
		fmt.Printf("Username: %s, ID: %s, Role: %s\n", u.Username, u.ID, u.Role)
	}
}
