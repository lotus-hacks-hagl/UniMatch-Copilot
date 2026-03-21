package database

import (
	"time"

	"unimatch-be/internal/model"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewPostgres(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	// AutoMigrate — order matters for foreign keys
	if err := db.AutoMigrate(
		&model.User{},
		&model.Student{},
		&model.University{},
		&model.Case{},
		&model.Recommendation{},
		&model.ActivityLog{},
	); err != nil {
		return nil, err
	}

	// Create indexes for common query patterns
	db.Exec("CREATE INDEX IF NOT EXISTS idx_cases_status ON cases(status)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_cases_created_at ON cases(created_at DESC)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_universities_crawl_status ON universities(crawl_status)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_activity_logs_created_at ON activity_logs(created_at DESC)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_recommendations_case_id ON recommendations(case_id)")

	return db, nil
}
