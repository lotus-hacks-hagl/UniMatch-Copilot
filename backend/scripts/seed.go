//go:build ignore

package main

import (
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
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
	if dbURL == "" {
		dbURL = "postgres://postgres:password@localhost:5432/unimatch_be?sslmode=disable"
	}

	dialector := postgres.Open(dbURL)
	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		log.Println("⚠️  Failed to connect to DB, trying localhost fallback...")
		localURL := "postgres://postgres:password@localhost:5432/unimatch_be?sslmode=disable"
		db, err = gorm.Open(postgres.Open(localURL), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
	}

	if err != nil {
		log.Fatalf("❌ failed to connect database: %v", err)
	}

	gofakeit.Seed(42)
	rng := rand.New(rand.NewSource(42))

	// ─── TRUNCATE ALL TABLES ─────────────────────────────────────────────────────
	log.Println("🗑️  Truncating all tables...")
	db.Exec("TRUNCATE TABLE recommendations, cases, students, universities, activity_logs, users RESTART IDENTITY CASCADE")
	log.Println("✅ Tables cleared")

	// ─── USERS ───────────────────────────────────────────────────────────────────
	log.Println("👤 Seeding users...")

	hashPw := func(pw string) string {
		h, _ := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
		return string(h)
	}

	adminUser := model.User{
		Username:     "admin",
		PasswordHash: hashPw("admin@123"),
		Role:         "admin",
		IsVerified:   true,
	}
	db.Create(&adminUser)

	teacherNames := []string{
		"nguyen.thao", "tran.minh", "le.huong", "pham.duc", "hoang.lan",
		"vu.thanh", "do.quang", "bui.ngoc", "dang.khanh", "ngo.anh",
		"ly.tuan", "mai.thu", "dinh.nam", "cao.phuong", "luong.viet",
	}

	var teachers []model.User
	for _, name := range teacherNames {
		t := model.User{
			Username:     name + "@unimatch.com",
			PasswordHash: hashPw("teacher123"),
			Role:         "teacher",
			IsVerified:   true,
		}
		db.Create(&t)
		teachers = append(teachers, t)
	}
	log.Printf("✅ Created 1 admin + %d teachers", len(teachers))

	// ─── UNIVERSITIES ─────────────────────────────────────────────────────────────
	log.Println("🏛️  Seeding universities...")

	type uniDef struct {
		Name           string
		Country        string
		QsRank         int
		AcceptanceRate float64
		Tuition        int
		IeltsMin       float64
		SatMin         int
		Majors         []string
	}

	topUnis := []uniDef{
		{"Massachusetts Institute of Technology", "USA", 1, 0.04, 57986, 7.5, 1540, []string{"Engineering", "Computer Science", "Physics"}},
		{"University of Cambridge", "UK", 2, 0.21, 35000, 7.5, 1500, []string{"Law", "Medicine", "Engineering"}},
		{"University of Oxford", "UK", 3, 0.17, 37000, 7.5, 1520, []string{"Philosophy", "Medicine", "Economics"}},
		{"Harvard University", "USA", 4, 0.04, 54768, 7.5, 1560, []string{"Business", "Law", "Medicine"}},
		{"Stanford University", "USA", 5, 0.04, 60700, 7.5, 1540, []string{"Computer Science", "Engineering", "Business"}},
		{"Imperial College London", "UK", 6, 0.15, 38000, 7.0, 1480, []string{"Engineering", "Medicine", "Business"}},
		{"ETH Zürich", "Switzerland", 7, 0.27, 1400, 7.0, 1450, []string{"Engineering", "Computer Science", "Architecture"}},
		{"National University of Singapore", "Singapore", 8, 0.17, 18000, 6.5, 1400, []string{"Computer Science", "Business", "Engineering"}},
		{"UCL", "UK", 9, 0.63, 26000, 7.0, 1420, []string{"Architecture", "Medicine", "Arts & Design"}},
		{"University of Chicago", "USA", 10, 0.06, 60000, 7.5, 1530, []string{"Economics", "Business", "Law"}},
		{"Nanyang Technological University", "Singapore", 11, 0.33, 17000, 6.5, 1380, []string{"Engineering", "Business", "Computer Science"}},
		{"University of Toronto", "Canada", 18, 0.43, 42000, 6.5, 1350, []string{"Computer Science", "Medicine", "Business"}},
		{"University of Melbourne", "Australia", 14, 0.70, 35000, 7.0, 1320, []string{"Business", "Arts & Design", "Engineering"}},
		{"University of British Columbia", "Canada", 34, 0.52, 38000, 6.5, 1350, []string{"Engineering", "Business", "Arts & Design"}},
		{"University of Edinburgh", "UK", 27, 0.45, 27000, 7.0, 1380, []string{"Computer Science", "Philosophy", "Medicine"}},
		{"University of Hong Kong", "China", 26, 0.22, 18000, 6.5, 1350, []string{"Business", "Law", "Medicine"}},
		{"Seoul National University", "South Korea", 31, 0.15, 5000, 6.5, 1300, []string{"Engineering", "Medicine", "Business"}},
		{"University of Sydney", "Australia", 18, 0.30, 36000, 6.5, 1320, []string{"Medicine", "Business", "Engineering"}},
		{"McGill University", "Canada", 30, 0.46, 28000, 7.0, 1380, []string{"Medicine", "Engineering", "Computer Science"}},
		{"Australian National University", "Australia", 30, 0.35, 33000, 6.5, 1300, []string{"Sciences", "Engineering", "Law"}},
	}

	countries := []string{"USA", "UK", "Canada", "Australia", "Singapore", "Germany", "Netherlands", "Japan", "South Korea", "France"}
	majors := []string{"Computer Science", "Business Administration", "Engineering", "Medicine", "Law", "Arts & Design", "Psychology", "Economics", "Architecture", "Data Science"}
	crawlStatuses := []string{model.CrawlStatusPending, model.CrawlStatusOK, model.CrawlStatusFailed, model.CrawlStatusNeverCrawled}

	var universities []model.University

	// Insert curated top unis
	for _, u := range topUnis {
		deadline := time.Now().AddDate(0, rng.Intn(8)+1, 0)
		uni := model.University{
			Name:                u.Name,
			Country:             u.Country,
			QsRank:              &[]int{u.QsRank}[0],
			AcceptanceRate:      &[]float64{u.AcceptanceRate}[0],
			TuitionUsdPerYear:   &[]int{u.Tuition}[0],
			IeltsMin:            &[]float64{u.IeltsMin}[0],
			AvailableMajors:     pq.StringArray(u.Majors),
			ApplicationDeadline: &deadline,
			CrawlStatus:         model.CrawlStatusOK,
		}
		db.Create(&uni)
		universities = append(universities, uni)
	}

	// Generate 40 more fake universities
	for i := 0; i < 30; i++ {
		country := countries[rng.Intn(len(countries))]
		rankVal := rng.Intn(900) + 100
		acceptance := 0.10 + rng.Float64()*0.65
		tuition := 10000 + rng.Intn(55000)
		ieltsMin := 5.5 + float64(rng.Intn(5))*0.5

		deadline := time.Now().AddDate(0, rng.Intn(12)+1, 0)
		uni := model.University{
			Name:                gofakeit.Company() + " University",
			Country:             country,
			QsRank:              &rankVal,
			AcceptanceRate:      &acceptance,
			TuitionUsdPerYear:   &tuition,
			IeltsMin:            &ieltsMin,
			AvailableMajors:     pq.StringArray(sampleStrings(majors, 2+rng.Intn(4), rng)),
			ApplicationDeadline: &deadline,
			CrawlStatus:         crawlStatuses[rng.Intn(len(crawlStatuses))],
		}
		db.Create(&uni)
		universities = append(universities, uni)
	}
	log.Printf("✅ Created %d universities", len(universities))

	// ─── STUDENTS ─────────────────────────────────────────────────────────────────
	log.Println("🎓 Seeding 200 students...")

	intakes := []string{"Fall 2025", "Spring 2026", "Fall 2026", "Spring 2027", "Fall 2027"}
	gpaScales := []float64{4.0, 10.0, 100.0}
	targetCountries := [][]string{
		{"USA"}, {"UK"}, {"Canada"}, {"Australia"}, {"Singapore"},
		{"USA", "UK"}, {"UK", "Canada"}, {"Australia", "Singapore"},
		{"USA", "Canada", "Australia"}, {"UK", "Netherlands", "Germany"},
	}

	var students []model.Student
	for i := 0; i < 200; i++ {
		gscale := gpaScales[rng.Intn(len(gpaScales))]
		var graw float64
		switch gscale {
		case 4.0:
			graw = 2.5 + rng.Float64()*1.5
		case 10.0:
			graw = 5.0 + rng.Float64()*5.0
		case 100.0:
			graw = 55.0 + rng.Float64()*45.0
		}
		gnorm := (graw / gscale) * 10.0

		ielts := 5.5 + float64(rng.Intn(8))*0.5
		sat := 1000 + rng.Intn(600)
		hasSat := rng.Float32() < 0.6
		hasIelts := rng.Float32() < 0.8

		var ieltsPtr *float64
		var satPtr *int
		if hasIelts {
			ieltsPtr = &ielts
		}
		if hasSat {
			satPtr = &sat
		}
		// ensure at least one test score
		if !hasIelts && !hasSat {
			ieltsPtr = &ielts
		}

		countries := targetCountries[rng.Intn(len(targetCountries))]
		budgets := []int{15000, 20000, 25000, 30000, 35000, 40000, 50000, 60000, 70000}

		s := model.Student{
			FullName:               gofakeit.Name(),
			GpaRaw:                 graw,
			GpaScale:               gscale,
			GpaNormalized:          gnorm,
			IeltsOverall:           ieltsPtr,
			SatTotal:               satPtr,
			IntendedMajor:          majors[rng.Intn(len(majors))],
			BudgetUsdPerYear:       budgets[rng.Intn(len(budgets))],
			PreferredCountries:     countries,
			TargetIntake:           intakes[rng.Intn(len(intakes))],
			ScholarshipRequired:    rng.Float32() < 0.35,
			Extracurriculars:       gofakeit.Sentence(8),
			Achievements:           gofakeit.Sentence(6),
			PersonalStatementNotes: gofakeit.Paragraph(2, 3, 10, " "),
		}
		db.Create(&s)
		students = append(students, s)
	}
	log.Printf("✅ Created %d students", len(students))

	// ─── CASES + RECOMMENDATIONS ──────────────────────────────────────────────────
	log.Println("📋 Seeding 200 cases with recommendations...")

	statuses := []string{
		model.CaseStatusDone, model.CaseStatusDone, model.CaseStatusDone, model.CaseStatusDone, // 40%
		model.CaseStatusProcessing, model.CaseStatusProcessing, // 20%
		model.CaseStatusPending, model.CaseStatusPending, // 15% (approx)
		model.CaseStatusHumanReview, model.CaseStatusHumanReview, // 20%
		model.CaseStatusFailed, // 10%
	}

	escalationReasons := []string{
		"High SAT score but low GPA requires manual verification",
		"Uncommon major-country combination — needs specialist review",
		"Budget constraints may conflict with target university tuition",
		"Borderline IELTS score for top-tier UK universities",
		"Student has significant gap year — advisor input required",
		"Attendance at non-standard high school — transcript unclear",
		"Scholarship dependency changes risk profile significantly",
		"Multiple failed attempts at standardized tests noted",
		"Target intake too soon — application timeline is critical",
		"Research background exceptional — needs special matching",
	}

	tierReasons := map[string][]string{
		"safe": {
			"Strong GPA aligns well within acceptance threshold",
			"Test scores comfortably meet minimum requirements",
			"Budget fully covers expected annual costs",
			"Major offered as a core program at this institution",
		},
		"match": {
			"Competitive profile with moderate admission probability",
			"Academic metrics fall near the 50th percentile of admitted students",
			"Strong extracurricular profile may compensate for borderline GPA",
			"IELTS score is on the boundary — preparation recommended",
		},
		"reach": {
			"Top-tier institution with highly selective admission process",
			"Test scores are below the typical admitted student profile",
			"Availability of scholarships makes this financially risky",
			"Admission rate below 10% makes this aspirational but possible",
		},
	}

	for i, student := range students {
		status := statuses[rng.Intn(len(statuses))]

		// Randomize created_at over last 90 days
		daysAgo := rng.Intn(90)
		hoursAgo := rng.Intn(24)
		createdAt := time.Now().AddDate(0, 0, -daysAgo).Add(-time.Duration(hoursAgo) * time.Hour)

		confidence := 0.50 + rng.Float64()*0.50 // stored as 0-1 fraction; dashboard multiplies by 100
		teacher := teachers[i%len(teachers)]

		var escalationReason *string
		var processingStarted, processingFinished *time.Time

		if status == model.CaseStatusHumanReview {
			reason := escalationReasons[rng.Intn(len(escalationReasons))]
			escalationReason = &reason
			start := createdAt.Add(4 * time.Minute)
			end := createdAt.Add(time.Duration(20+rng.Intn(60)) * time.Minute)
			processingStarted = &start
			processingFinished = &end
		} else if status == model.CaseStatusDone || status == model.CaseStatusFailed {
			start := createdAt.Add(2 * time.Minute)
			end := createdAt.Add(time.Duration(5+rng.Intn(30)) * time.Minute)
			processingStarted = &start
			processingFinished = &end
		} else if status == model.CaseStatusProcessing {
			start := createdAt.Add(1 * time.Minute)
			processingStarted = &start
		}

		profileSummary := datatypes.JSON(`{"strengths":["Strong academic record","Relevant extracurriculars","Clear goals"],"weaknesses":["Limited research experience","Standardized test scores on boundary"]}`)

		c := model.Case{
			StudentID:            student.ID,
			AssignedToID:         &teacher.ID,
			Status:               status,
			AiConfidence:         &confidence,
			EscalationReason:     escalationReason,
			ProfileSummary:       profileSummary,
			ProcessingStartedAt:  processingStarted,
			ProcessingFinishedAt: processingFinished,
		}
		c.CreatedAt = createdAt
		c.UpdatedAt = createdAt

		db.Create(&c)

		// ─── Recommendations ─────────────────────────────────────────────
		if status == model.CaseStatusDone || status == model.CaseStatusHumanReview {
			numRecs := 3 + rng.Intn(3) // 3–5 recs
			shuffledUnis := shuffleUnis(universities, rng)
			tierOrder := []string{"safe", "match", "reach"}

			for rank := 0; rank < numRecs && rank < len(shuffledUnis); rank++ {
				uni := shuffledUnis[rank]
				tier := tierOrder[rank%3]
				admLikelihood := 40 + rng.Float64()*55
				fitScore := 50 + rng.Float64()*50
				reason := tierReasons[tier][rng.Intn(len(tierReasons[tier]))]

				uniID := uni.ID
				_ = uniID // avoid unused var if uni is zero
				var universityID uuid.UUID
				if uni.ID != (uuid.UUID{}) {
					universityID = uni.ID
				}
				rec := model.Recommendation{
					CaseID:                   c.ID,
					UniversityID:             universityID,
					UniversityName:           uni.Name,
					Tier:                     tier,
					RankOrder:                rank + 1,
					AdmissionLikelihoodScore: int(admLikelihood),
					StudentFitScore:          int(fitScore),
					Reason:                   reason,
				}
				db.Create(&rec)
			}
		}
		_ = i
	}

	log.Println("✅ Seeding complete!")
	log.Printf("   📊 1 admin + %d teachers", len(teachers))
	log.Printf("   🏛️  %d universities", len(universities))
	log.Printf("   🎓 %d students", len(students))
	log.Printf("   📋 200 cases with recommendations")
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func toJSON(arr []string) []byte {
	if len(arr) == 0 {
		return []byte(`[]`)
	}
	result := `[`
	for i, s := range arr {
		if i > 0 {
			result += ","
		}
		result += `"` + s + `"`
	}
	result += `]`
	return []byte(result)
}

func sampleStrings(arr []string, n int, rng *rand.Rand) []string {
	if n > len(arr) {
		n = len(arr)
	}
	perm := rng.Perm(len(arr))
	result := make([]string, n)
	for i := 0; i < n; i++ {
		result[i] = arr[perm[i]]
	}
	return result
}

func shuffleUnis(unis []model.University, rng *rand.Rand) []model.University {
	result := make([]model.University, len(unis))
	copy(result, unis)
	rng.Shuffle(len(result), func(i, j int) {
		result[i], result[j] = result[j], result[i]
	})
	return result
}
