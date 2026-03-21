package main

import (
	"log"
	"os"

	"unimatch-be/internal/model"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	_ = godotenv.Load()
	dbURL := os.Getenv("DATABASE_URL")
	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin@123"), bcrypt.DefaultCost)
	testUser := model.User{
		Username:     "testadmin@unimatch.com",
		PasswordHash: string(hashedPassword),
		Role:         "admin",
		IsVerified:   true,
	}

	db.Where("username = ?", testUser.Username).Delete(&model.User{})
	if err := db.Create(&testUser).Error; err != nil {
		log.Fatal(err)
	}
	log.Println("Test user created: testadmin@unimatch.com / admin@123")
}
