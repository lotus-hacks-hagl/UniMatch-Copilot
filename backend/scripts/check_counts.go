package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
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

	var userCount, uniCount, studentCount, caseCount, recCount int64
	db.Table("users").Count(&userCount)
	db.Table("universities").Count(&uniCount)
	db.Table("students").Count(&studentCount)
	db.Table("cases").Count(&caseCount)
	db.Table("recommendations").Count(&recCount)

	fmt.Printf("Users: %d\n", userCount)
	fmt.Printf("Universities: %d\n", uniCount)
	fmt.Printf("Students: %d\n", studentCount)
	fmt.Printf("Cases: %d\n", caseCount)
	fmt.Printf("Recommendations: %d\n", recCount)
}
