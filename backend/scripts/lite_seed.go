//go:build ignore

package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/datatypes"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"unimatch-be/internal/model"
)

func main() {
	_ = godotenv.Load()
	dbURL := os.Getenv("DATABASE_URL")

	dbURL = "postgres://postgres:password@localhost:5432/unimatch_be?sslmode=disable"

	// If we are running on host but DATABASE_URL points to 'db' (Docker service name),
	// we try to swap it to 'localhost' for local execution convenience.
	// This is a common dev-friendly hack.
	dialector := postgres.Open(dbURL)
	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil && (dbURL == "postgres://postgres:password@db:5432/unimatch_be?sslmode=disable" || os.Getenv("ENV") != "production") {
		log.Println("⚠️  Failed to connect using Docker hostname 'db', trying 'localhost'...")
		localURL := "postgres://postgres:password@localhost:5432/unimatch_be?sslmode=disable"
		db, err = gorm.Open(postgres.Open(localURL), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
	}

	if err != nil {
		log.Fatalf("❌ failed to connect database: %v", err)
	}

	// ─── TRUNCATE ALL TABLES ─────────────────────────────────────────────────────
	log.Println("🗑️  Truncating relevant tables...")
	db.Exec("TRUNCATE TABLE recommendations, cases, students, universities, activity_logs, users RESTART IDENTITY CASCADE")

	// ─── USERS ───────────────────────────────────────────────────────────────────
	log.Println("👤 Seeding admin and teacher...")
	hashPw := func(pw string) string {
		h, _ := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
		return string(h)
	}

	admin := model.User{
		Username:     "admin@unimatch.com",
		PasswordHash: hashPw("admin123"),
		Role:         "admin",
		IsVerified:   true,
	}
	db.Create(&admin)

	teacher := model.User{
		Username:     "teacher@unimatch.com",
		PasswordHash: hashPw("teacher123"),
		Role:         "teacher",
		IsVerified:   true,
	}
	db.Create(&teacher)

	// ─── UNIVERSITIES ─────────────────────────────────────────────────────────────
	log.Println("🏛️  Seeding 3 universities...")

	unis := []model.University{
		{
			Name:              "Massachusetts Institute of Technology",
			Country:           "USA",
			QsRank:            &[]int{1}[0],
			AcceptanceRate:    &[]float64{0.04}[0],
			TuitionUsdPerYear: &[]int{58000}[0],
			IeltsMin:          &[]float64{7.5}[0],
			AvailableMajors:   pq.StringArray([]string{"Engineering", "Computer Science", "Physics"}),
			CrawlStatus:       model.CrawlStatusNeverCrawled,
		},
		{
			Name:              "University of Oxford",
			Country:           "UK",
			QsRank:            &[]int{3}[0],
			AcceptanceRate:    &[]float64{0.17}[0],
			TuitionUsdPerYear: &[]int{37000}[0],
			IeltsMin:          &[]float64{7.5}[0],
			AvailableMajors:   pq.StringArray([]string{"Philosophy", "Medicine", "Economics"}),
			CrawlStatus:       model.CrawlStatusNeverCrawled,
		},
		{
			Name:              "National University of Singapore",
			Country:           "Singapore",
			QsRank:            &[]int{8}[0],
			AcceptanceRate:    &[]float64{0.17}[0],
			TuitionUsdPerYear: &[]int{18000}[0],
			IeltsMin:          &[]float64{6.5}[0],
			AvailableMajors:   pq.StringArray([]string{"Computer Science", "Business", "Engineering"}),
			CrawlStatus:       model.CrawlStatusNeverCrawled,
		},
	}

	for i := range unis {
		db.Create(&unis[i])
	}

	// ─── STUDENTS & CASES ─────────────────────────────────────────────────────────
	log.Println("🎓 Seeding sample students and cases...")

	student := model.Student{
		FullName:           "John Doe",
		GpaRaw:             3.8,
		GpaScale:           4.0,
		GpaNormalized:      9.5,
		IntendedMajor:      "Computer Science",
		BudgetUsdPerYear:   30000,
		PreferredCountries: pq.StringArray([]string{"USA", "Singapore"}),
		TargetIntake:       "Fall 2025",
	}
	db.Create(&student)

	caseRecord := model.Case{
		StudentID:      student.ID,
		AssignedToID:   &teacher.ID,
		Status:         model.CaseStatusDone,
		AiConfidence:   &[]float64{0.92}[0],
		ProfileSummary: datatypes.JSON(`{"strengths":["Excellent GPA","Strong CS background"]}`),
	}
	db.Create(&caseRecord)

	// Recommendation
	rec := model.Recommendation{
		CaseID:                   caseRecord.ID,
		UniversityID:             unis[0].ID,
		UniversityName:           unis[0].Name,
		Tier:                     "reach",
		RankOrder:                1,
		AdmissionLikelihoodScore: 45,
		StudentFitScore:          95,
		Reason:                   "Perfect alignment with candidate's research interests despite being highly selective.",
	}
	db.Create(&rec)

	log.Println("✅ Lite seeding complete!")
}
