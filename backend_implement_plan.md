# UniMatch Copilot — Backend Implementation Plan

> **Stack:** Go 1.22+ · Gin · GORM v2 · PostgreSQL · godotenv  
> **Role:** BE là orchestrator. Không biết Claude, Neo4j tồn tại. Chỉ giao tiếp với AI Service qua HTTP.  
> **Thời gian ước tính:** ~5h (hackathon 24h, 1 người BE)

---

## 1. Tech Stack & Dependencies

```bash
go mod init unimatch-be

# Core
go get github.com/gin-gonic/gin
go get gorm.io/gorm
go get gorm.io/driver/postgres

# Utilities
go get github.com/google/uuid
go get github.com/joho/godotenv
go get github.com/go-playground/validator/v10

# PDF generation
go get github.com/jung-kurt/gofpdf
```

**Không dùng Redis** cho hackathon — đơn giản hóa, GORM query thẳng PG là đủ.

---

## 2. Folder Structure (N-Layer Architecture)

```
unimatch-be/
├── main.go                          # Entry point, wire dependencies
├── config/
│   └── config.go                    # Load env vars
├── internal/
│   ├── handler/
│   │   ├── cases_handler.go         # Cases CRUD + report
│   │   ├── universities_handler.go  # Universities CRUD + crawl
│   │   ├── dashboard_handler.go     # Stats, analytics
│   │   └── internal_handler.go      # POST /internal/jobs/done (callback)
│   ├── service/
│   │   ├── interfaces.go            # Service interfaces
│   │   ├── cases_service.go
│   │   ├── universities_service.go
│   │   └── dashboard_service.go
│   ├── repository/
│   │   ├── interfaces.go            # Repository interfaces
│   │   ├── cases_repository.go
│   │   ├── universities_repository.go
│   │   ├── activity_repository.go
│   │   └── dashboard_repository.go
│   ├── model/
│   │   ├── base.go                  # BaseModel với UUID
│   │   ├── student.go
│   │   ├── case.go
│   │   ├── university.go
│   │   ├── recommendation.go
│   │   └── activity_log.go
│   ├── dto/
│   │   ├── cases_dto.go
│   │   ├── universities_dto.go
│   │   └── dashboard_dto.go
│   ├── middleware/
│   │   ├── cors.go
│   │   └── logger.go
│   └── router/
│       └── router.go
├── pkg/
│   ├── database/
│   │   └── postgres.go              # GORM init + AutoMigrate
│   ├── response/
│   │   └── response.go              # Standard JSON response
│   ├── apperror/
│   │   └── apperror.go
│   └── client/
│       └── ai_client.go             # HTTP client gọi AI Service
└── migrations/
    ├── 001_init.sql
    └── 002_seed.sql
```

**Nguyên tắc layer:**
- `handler` → nhận HTTP, validate, gọi `service`, trả response
- `service` → business logic, gọi `repository`
- `repository` → GORM queries, trả model
- Không layer nào được import layer cùng cấp hoặc layer trên nó

---

## 3. Config

**`config/config.go`:**
```go
type Config struct {
    Port          string
    DatabaseURL   string // postgres://user:pass@host:5432/dbname
    AIServiceURL  string // http://localhost:9000
    PublicBaseURL string // http://localhost:8080
    Env           string // development | production
}

func Load() *Config {
    godotenv.Load()
    return &Config{
        Port:          getEnv("PORT", "8080"),
        DatabaseURL:   mustEnv("DATABASE_URL"),
        AIServiceURL:  getEnv("AI_SERVICE_URL", "http://localhost:9000"),
        PublicBaseURL: getEnv("PUBLIC_BASE_URL", "http://localhost:8080"),
        Env:           getEnv("ENV", "development"),
    }
}
```

**`.env`:**
```
PORT=8080
DATABASE_URL=postgres://postgres:password@localhost:5432/unimatch_be
AI_SERVICE_URL=http://localhost:9000
PUBLIC_BASE_URL=http://localhost:8080
ENV=development
```

---

## 4. GORM Models

### `internal/model/base.go`
```go
type Base struct {
    ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (b *Base) BeforeCreate(tx *gorm.DB) error {
    if b.ID == uuid.Nil {
        b.ID = uuid.New()
    }
    return nil
}
```

### `internal/model/student.go`
```go
type Student struct {
    Base
    FullName               string         `json:"full_name" gorm:"not null"`
    GpaRaw                 float64        `json:"gpa_raw"`
    GpaScale               float64        `json:"gpa_scale" gorm:"default:10"`
    GpaNormalized          float64        `json:"gpa_normalized"`
    IeltsOverall           *float64       `json:"ielts_overall"`
    IeltsBreakdown         datatypes.JSON `json:"ielts_breakdown" gorm:"type:jsonb"`
    SatTotal               *int           `json:"sat_total"`
    ToeflTotal             *int           `json:"toefl_total"`
    IntendedMajor          string         `json:"intended_major"`
    BudgetUsdPerYear       int            `json:"budget_usd_per_year"`
    PreferredCountries     pq.StringArray `json:"preferred_countries" gorm:"type:text[]"`
    TargetIntake           string         `json:"target_intake"`
    ScholarshipRequired    bool           `json:"scholarship_required"`
    Extracurriculars       string         `json:"extracurriculars"`
    Achievements           string         `json:"achievements"`
    PersonalStatementNotes string         `json:"personal_statement_notes"`
}
```

### `internal/model/case.go`
```go
// Status constants
const (
    CaseStatusPending     = "pending"
    CaseStatusProcessing  = "processing"
    CaseStatusDone        = "done"
    CaseStatusHumanReview = "human_review"
    CaseStatusFailed      = "failed"
)

type Case struct {
    Base
    StudentID            uuid.UUID      `json:"student_id" gorm:"type:uuid;not null"`
    Student              *Student       `json:"student,omitempty" gorm:"foreignKey:StudentID"`
    Status               string         `json:"status" gorm:"default:'pending'"`
    AiJobID              *uuid.UUID     `json:"ai_job_id" gorm:"type:uuid"`
    AiConfidence         *float64       `json:"ai_confidence"`
    EscalationReason     *string        `json:"escalation_reason"`
    ProfileSummary       datatypes.JSON `json:"profile_summary" gorm:"type:jsonb"`
    ReportData           datatypes.JSON `json:"report_data,omitempty" gorm:"type:jsonb"`
    ReportGeneratedAt    *time.Time     `json:"report_generated_at"`
    ProcessingStartedAt  *time.Time     `json:"processing_started_at"`
    ProcessingFinishedAt *time.Time     `json:"processing_finished_at"`
    Recommendations      []Recommendation `json:"recommendations,omitempty" gorm:"foreignKey:CaseID"`
}

func (Case) TableName() string { return "cases" }
```

### `internal/model/university.go`
```go
const (
    CrawlStatusOK          = "ok"
    CrawlStatusPending     = "pending"
    CrawlStatusChanged     = "changed"
    CrawlStatusFailed      = "failed"
    CrawlStatusNeverCrawled = "never_crawled"
)

type University struct {
    Base
    Name                    string         `json:"name" gorm:"not null"`
    Country                 string         `json:"country"`
    QsRank                  *int           `json:"qs_rank"`
    GroupTag                string         `json:"group_tag"`
    IeltsMin                *float64       `json:"ielts_min"`
    SatRequired             bool           `json:"sat_required"`
    GpaExpectationNormalized *float64      `json:"gpa_expectation_normalized"`
    TuitionUsdPerYear       *int           `json:"tuition_usd_per_year"`
    ScholarshipAvailable    bool           `json:"scholarship_available"`
    ScholarshipNotes        string         `json:"scholarship_notes"`
    AvailableMajors         pq.StringArray `json:"available_majors" gorm:"type:text[]"`
    ApplicationDeadline     *time.Time     `json:"application_deadline"`
    AcceptanceRate          *float64       `json:"acceptance_rate"`
    CrawlStatus             string         `json:"crawl_status" gorm:"default:'never_crawled'"`
    LastCrawledAt           *time.Time     `json:"last_crawled_at"`
    CrawlJobID              *uuid.UUID     `json:"crawl_job_id" gorm:"type:uuid"`
    CounselorNotes          string         `json:"counselor_notes"`
}
```

### `internal/model/recommendation.go`
```go
type Recommendation struct {
    Base
    CaseID                  uuid.UUID      `json:"case_id" gorm:"type:uuid;not null"`
    UniversityID            uuid.UUID      `json:"university_id" gorm:"type:uuid"`
    UniversityName          string         `json:"university_name"`
    Tier                    string         `json:"tier"` // safe|match|reach
    AdmissionLikelihoodScore int           `json:"admission_likelihood_score"`
    StudentFitScore         int            `json:"student_fit_score"`
    Reason                  string         `json:"reason"`
    Risks                   datatypes.JSON `json:"risks" gorm:"type:jsonb"`
    Improvements            datatypes.JSON `json:"improvements" gorm:"type:jsonb"`
    RankOrder               int            `json:"rank_order"`
}
```

### `internal/model/activity_log.go`
```go
const (
    EventCaseCreated       = "case_created"
    EventProcessingStarted = "processing_started"
    EventAutoApproved      = "auto_approved"
    EventEscalated         = "escalated"
    EventCrawlStarted      = "crawl_started"
    EventCrawlChange       = "crawl_change"
    EventCrawlDone         = "crawl_done"
    EventReportGenerated   = "report_generated"
)

type ActivityLog struct {
    Base
    CaseID       *uuid.UUID     `json:"case_id,omitempty" gorm:"type:uuid"`
    UniversityID *uuid.UUID     `json:"university_id,omitempty" gorm:"type:uuid"`
    EventType    string         `json:"event_type"`
    Description  string         `json:"description"`
    Metadata     datatypes.JSON `json:"metadata" gorm:"type:jsonb"`
}
```

---

## 5. Database Setup (GORM AutoMigrate)

**`pkg/database/postgres.go`:**
```go
func NewPostgres(dsn string) (*gorm.DB, error) {
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Warn),
    })
    if err != nil {
        return nil, err
    }

    sqlDB, _ := db.DB()
    sqlDB.SetMaxOpenConns(25)
    sqlDB.SetMaxIdleConns(10)
    sqlDB.SetConnMaxLifetime(5 * time.Minute)

    // AutoMigrate — thứ tự quan trọng (foreign keys)
    err = db.AutoMigrate(
        &model.Student{},
        &model.University{},
        &model.Case{},
        &model.Recommendation{},
        &model.ActivityLog{},
    )
    return db, err
}
```

**Indexes cần tạo thêm** (GORM tag hoặc raw SQL sau migrate):
```go
// Trong AutoMigrate hoặc init function
db.Exec("CREATE INDEX IF NOT EXISTS idx_cases_status ON cases(status)")
db.Exec("CREATE INDEX IF NOT EXISTS idx_cases_created_at ON cases(created_at DESC)")
db.Exec("CREATE INDEX IF NOT EXISTS idx_universities_crawl_status ON universities(crawl_status)")
db.Exec("CREATE INDEX IF NOT EXISTS idx_activity_log_created_at ON activity_log(created_at DESC)")
```

---

## 6. Standard Packages

### `pkg/response/response.go`
```go
type Response struct {
    Success   bool        `json:"success"`
    Data      interface{} `json:"data,omitempty"`
    Error     *ErrorInfo  `json:"error,omitempty"`
    Meta      *Meta       `json:"meta,omitempty"`
}

type ErrorInfo struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Details string `json:"details,omitempty"`
}

type Meta struct {
    Page       int   `json:"page"`
    Limit      int   `json:"limit"`
    Total      int64 `json:"total"`
    TotalPages int   `json:"total_pages"`
    HasNext    bool  `json:"has_next"`
}

func OK(c *gin.Context, data interface{})                    { c.JSON(200, Response{Success: true, Data: data}) }
func Created(c *gin.Context, data interface{})               { c.JSON(201, Response{Success: true, Data: data}) }
func Fail(c *gin.Context, status int, code, msg string)      { c.JSON(status, Response{Success: false, Error: &ErrorInfo{Code: code, Message: msg}}) }
func Paginated(c *gin.Context, data interface{}, meta Meta)  { c.JSON(200, Response{Success: true, Data: data, Meta: &meta}) }
```

### `pkg/client/ai_client.go`
```go
type AIClient struct {
    baseURL    string
    httpClient *http.Client
}

func NewAIClient(baseURL string) *AIClient {
    return &AIClient{
        baseURL:    baseURL,
        httpClient: &http.Client{Timeout: 10 * time.Second},
    }
}

// Request structs khớp với Shared Contracts
type CrawlJobRequest struct {
    JobID       string                 `json:"job_id"`
    UniversityID string               `json:"university_id"`
    CallbackURL  string               `json:"callback_url"`
    Metadata     map[string]interface{} `json:"metadata"`
}

type AnalyzeJobRequest struct {
    JobID       string      `json:"job_id"`
    CaseID      string      `json:"case_id"`
    CallbackURL string      `json:"callback_url"`
    Input       AnalyzeInput `json:"input"`
}

type AnalyzeInput struct {
    FullName           string   `json:"full_name"`
    GpaNormalized      float64  `json:"gpa_normalized"`
    IeltsOverall       *float64 `json:"ielts_overall"`
    SatTotal           *int     `json:"sat_total"`
    IntendedMajor      string   `json:"intended_major"`
    BudgetUsdPerYear   int      `json:"budget_usd_per_year"`
    PreferredCountries []string `json:"preferred_countries"`
    TargetIntake       string   `json:"target_intake"`
    ScholarshipRequired bool    `json:"scholarship_required"`
    Extracurriculars   string   `json:"extracurriculars"`
    Achievements       string   `json:"achievements"`
}

type ReportJobRequest struct {
    JobID           string           `json:"job_id"`
    CaseID          string           `json:"case_id"`
    CallbackURL     string           `json:"callback_url"`
    StudentName     string           `json:"student_name"`
    Recommendations []interface{}   `json:"recommendations"`
}

func (c *AIClient) SubmitCrawlJob(req CrawlJobRequest) error   { return c.post("/jobs/crawl", req) }
func (c *AIClient) SubmitAnalyzeJob(req AnalyzeJobRequest) error { return c.post("/jobs/analyze", req) }
func (c *AIClient) SubmitReportJob(req ReportJobRequest) error  { return c.post("/jobs/report", req) }

func (c *AIClient) post(path string, body interface{}) error {
    b, _ := json.Marshal(body)
    resp, err := c.httpClient.Post(c.baseURL+path, "application/json", bytes.NewReader(b))
    if err != nil { return err }
    defer resp.Body.Close()
    if resp.StatusCode != 200 {
        return fmt.Errorf("ai service error: %d", resp.StatusCode)
    }
    return nil
}
```

---

## 7. Repository Layer

### `internal/repository/interfaces.go`
```go
type CaseRepository interface {
    Create(ctx context.Context, c *model.Case) error
    FindByID(ctx context.Context, id uuid.UUID) (*model.Case, error)
    FindAll(ctx context.Context, status string, page, limit int) ([]model.Case, int64, error)
    Update(ctx context.Context, c *model.Case) error
    Count(ctx context.Context, status string) (int64, error)
}

type UniversityRepository interface {
    Create(ctx context.Context, u *model.University) error
    FindByID(ctx context.Context, id uuid.UUID) (*model.University, error)
    FindAll(ctx context.Context, country, search string, page, limit int) ([]model.University, int64, error)
    FindCrawlable(ctx context.Context) ([]model.University, error)
    CountByCrawlStatus(ctx context.Context, status string) (int64, error)
    UpdateCrawlResult(ctx context.Context, id uuid.UUID, fields map[string]interface{}) error
}

type ActivityRepository interface {
    Create(ctx context.Context, log *model.ActivityLog) error
    FindRecent(ctx context.Context, limit int) ([]model.ActivityLog, error)
}

type DashboardRepository interface {
    GetStats(ctx context.Context) (*dto.DashboardStats, error)
    GetCasesByDay(ctx context.Context) ([]dto.CasesByDay, error)
    GetEscalationTrend(ctx context.Context) ([]dto.EscalationTrend, error)
    GetAnalytics(ctx context.Context) (*dto.Analytics, error)
}
```

### `internal/repository/cases_repository.go` — key patterns
```go
type caseRepository struct { db *gorm.DB }

func NewCaseRepository(db *gorm.DB) CaseRepository {
    return &caseRepository{db: db}
}

func (r *caseRepository) Create(ctx context.Context, c *model.Case) error {
    return r.db.WithContext(ctx).Create(c).Error
}

func (r *caseRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Case, error) {
    var c model.Case
    err := r.db.WithContext(ctx).
        Preload("Student").
        Preload("Recommendations", func(db *gorm.DB) *gorm.DB {
            return db.Order("rank_order ASC")
        }).
        First(&c, "id = ?", id).Error
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, nil
    }
    return &c, err
}

// FindAll với JOIN student để có thông tin display
func (r *caseRepository) FindAll(ctx context.Context, status string, page, limit int) ([]model.Case, int64, error) {
    var cases []model.Case
    var total int64
    
    q := r.db.WithContext(ctx).Model(&model.Case{}).Preload("Student")
    if status != "" && status != "all" {
        q = q.Where("status = ?", status)
    }
    
    q.Count(&total)
    err := q.Order("created_at DESC").
        Offset((page-1)*limit).Limit(limit).
        Find(&cases).Error
    return cases, total, err
}
```

### `internal/repository/universities_repository.go` — key patterns
```go
func (r *universityRepository) FindCrawlable(ctx context.Context) ([]model.University, error) {
    var unis []model.University
    err := r.db.WithContext(ctx).
        Where("crawl_status != ? AND (last_crawled_at IS NULL OR last_crawled_at < NOW() - INTERVAL '1 day')", 
              model.CrawlStatusPending).
        Find(&unis).Error
    return unis, err
}

func (r *universityRepository) UpdateCrawlResult(ctx context.Context, id uuid.UUID, fields map[string]interface{}) error {
    return r.db.WithContext(ctx).Model(&model.University{}).
        Where("id = ?", id).
        Updates(fields).Error
}
```

### Dashboard Repository — raw SQL queries
```go
func (r *dashboardRepository) GetStats(ctx context.Context) (*dto.DashboardStats, error) {
    var stats dto.DashboardStats
    
    r.db.WithContext(ctx).Model(&model.Case{}).
        Where("DATE(created_at) = CURRENT_DATE").Count(&stats.CasesToday)
    
    r.db.WithContext(ctx).Model(&model.Case{}).
        Where("status = ?", model.CaseStatusHumanReview).Count(&stats.AwaitingReview)
    
    r.db.WithContext(ctx).Raw(`
        SELECT COALESCE(AVG(EXTRACT(EPOCH FROM (processing_finished_at - processing_started_at))/60), 0)
        FROM cases
        WHERE processing_finished_at IS NOT NULL
          AND created_at > NOW() - INTERVAL '7 days'
    `).Scan(&stats.AvgProcessingMinutes)
    
    r.db.WithContext(ctx).Raw(`
        SELECT COALESCE(AVG(ai_confidence) * 100, 0)
        FROM cases
        WHERE created_at > NOW() - INTERVAL '7 days'
          AND ai_confidence IS NOT NULL
    `).Scan(&stats.AiConfidenceAvg)
    
    return &stats, nil
}
```

---

## 8. Service Layer

### `internal/service/interfaces.go`
```go
type CaseService interface {
    Create(ctx context.Context, req dto.CreateCaseRequest) (*dto.CaseCreatedResponse, *apperror.AppError)
    GetByID(ctx context.Context, id uuid.UUID) (*model.Case, *apperror.AppError)
    List(ctx context.Context, status string, page, limit int) (*dto.ListCasesResponse, *apperror.AppError)
    Count(ctx context.Context, status string) (int64, *apperror.AppError)
    GetCrawlStatus(ctx context.Context, caseID uuid.UUID) ([]dto.CrawlStatusItem, *apperror.AppError)
    RequestReport(ctx context.Context, caseID uuid.UUID) (*dto.ReportStatusResponse, *apperror.AppError)
    HandleJobDone(ctx context.Context, payload dto.JobDonePayload) *apperror.AppError
}

type UniversityService interface {
    Create(ctx context.Context, req dto.CreateUniversityRequest) (*model.University, *apperror.AppError)
    List(ctx context.Context, country, search string, page, limit int) (*dto.ListUniversitiesResponse, *apperror.AppError)
    CrawlAll(ctx context.Context) (int, *apperror.AppError)
    CountActiveCrawls(ctx context.Context) (int64, *apperror.AppError)
}
```

### `internal/service/cases_service.go` — Critical flows

**Create Case (POST /api/v1/cases):**
```go
func (s *caseService) Create(ctx context.Context, req dto.CreateCaseRequest) (*dto.CaseCreatedResponse, *apperror.AppError) {
    // 1. Build student model
    student := &model.Student{
        FullName:            req.FullName,
        GpaNormalized:       req.GpaNormalized, // FE đã tính
        IeltsOverall:        req.IeltsOverall,
        // ... map các fields còn lại
    }

    // 2. Transaction: insert student + case + activity log
    var caseRecord model.Case
    err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        if err := tx.Create(student).Error; err != nil { return err }
        
        jobID := uuid.New()
        caseRecord = model.Case{
            StudentID: student.ID,
            Status:    model.CaseStatusPending,
            AiJobID:   &jobID,
        }
        if err := tx.Create(&caseRecord).Error; err != nil { return err }
        
        log := &model.ActivityLog{
            CaseID:      &caseRecord.ID,
            EventType:   model.EventCaseCreated,
            Description: fmt.Sprintf("Case created for %s", req.FullName),
        }
        return tx.Create(log).Error
    })
    if err != nil {
        return nil, apperror.Internal(err, "failed to create case")
    }

    // 3. Submit analyze job (fire & forget)
    analyzeReq := client.AnalyzeJobRequest{
        JobID:       caseRecord.AiJobID.String(),
        CaseID:      caseRecord.ID.String(),
        CallbackURL: s.cfg.PublicBaseURL + "/internal/jobs/done",
        Input:       buildAnalyzeInput(student),
    }
    
    now := time.Now()
    if err := s.aiClient.SubmitAnalyzeJob(analyzeReq); err != nil {
        // AI không nhận → failed
        s.db.Model(&caseRecord).Updates(map[string]interface{}{
            "status": model.CaseStatusFailed,
        })
        return nil, apperror.Internal(err, "failed to submit AI job")
    }

    // 4. Update status → processing
    s.db.Model(&caseRecord).Updates(map[string]interface{}{
        "status":                model.CaseStatusProcessing,
        "processing_started_at": &now,
    })
    s.db.Create(&model.ActivityLog{
        CaseID:    &caseRecord.ID,
        EventType: model.EventProcessingStarted,
    })

    return &dto.CaseCreatedResponse{CaseID: caseRecord.ID.String(), Status: model.CaseStatusProcessing}, nil
}
```

**HandleJobDone — callback router:**
```go
func (s *caseService) HandleJobDone(ctx context.Context, payload dto.JobDonePayload) *apperror.AppError {
    switch payload.JobType {
    case "crawl_university":
        return s.handleCrawlDone(ctx, payload)
    case "analyze_profile":
        return s.handleAnalyzeDone(ctx, payload)
    case "generate_report":
        return s.handleReportDone(ctx, payload)
    }
    return nil
}

func (s *caseService) handleAnalyzeDone(ctx context.Context, p dto.JobDonePayload) *apperror.AppError {
    caseID, _ := uuid.Parse(p.CaseID)
    now := time.Now()
    
    if p.Status == "failed" {
        s.db.Model(&model.Case{}).Where("id = ?", caseID).Updates(map[string]interface{}{
            "status":           model.CaseStatusHumanReview,
            "escalation_reason": "AI service failed",
        })
        return nil
    }

    var result dto.AnalyzeResult
    json.Unmarshal(p.Result, &result)

    err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        // Bulk insert recommendations
        recs := make([]model.Recommendation, len(result.Recommendations))
        for i, r := range result.Recommendations {
            recs[i] = model.Recommendation{
                CaseID:                   caseID,
                UniversityID:             uuid.MustParse(r.UniversityID),
                UniversityName:           r.UniversityName,
                Tier:                     r.Tier,
                AdmissionLikelihoodScore: r.AdmissionLikelihoodScore,
                StudentFitScore:          r.StudentFitScore,
                Reason:                   r.Reason,
                RankOrder:                r.RankOrder,
            }
        }
        if len(recs) > 0 {
            if err := tx.Create(&recs).Error; err != nil { return err }
        }

        // Determine final status
        finalStatus := model.CaseStatusDone
        if result.EscalationNeeded {
            finalStatus = model.CaseStatusHumanReview
        }

        profileJSON, _ := json.Marshal(result.ProfileSummary)
        updates := map[string]interface{}{
            "status":                 finalStatus,
            "ai_confidence":          result.ConfidenceScore,
            "profile_summary":        datatypes.JSON(profileJSON),
            "processing_finished_at": &now,
        }
        if result.EscalationReason != "" {
            updates["escalation_reason"] = result.EscalationReason
        }
        if err := tx.Model(&model.Case{}).Where("id = ?", caseID).Updates(updates).Error; err != nil {
            return err
        }

        eventType := model.EventAutoApproved
        if result.EscalationNeeded {
            eventType = model.EventEscalated
        }
        return tx.Create(&model.ActivityLog{CaseID: &caseID, EventType: eventType}).Error
    })
    
    if err != nil { return apperror.Internal(err, "failed to handle analyze done") }
    return nil
}

func (s *caseService) handleCrawlDone(ctx context.Context, p dto.JobDonePayload) *apperror.AppError {
    uniID, _ := uuid.Parse(p.UniversityID)
    
    if p.Status == "failed" {
        s.db.Model(&model.University{}).Where("id = ?", uniID).Updates(map[string]interface{}{
            "crawl_status": model.CrawlStatusFailed,
            "crawl_job_id": nil,
        })
        return nil
    }

    var result dto.CrawlResult
    json.Unmarshal(p.Result, &result)
    now := time.Now()

    // Build update map — chỉ update field nào không nil
    updates := map[string]interface{}{
        "crawl_status":   result.CrawlStatus,
        "last_crawled_at": &now,
        "crawl_job_id":   nil,
    }
    if result.QsRank != nil      { updates["qs_rank"] = result.QsRank }
    if result.IeltsMin != nil    { updates["ielts_min"] = result.IeltsMin }
    if result.TuitionUsdPerYear != nil { updates["tuition_usd_per_year"] = result.TuitionUsdPerYear }
    // ... tương tự cho các field khác

    s.db.Model(&model.University{}).Where("id = ?", uniID).Updates(updates)

    if len(result.ChangesDetected) > 0 {
        meta, _ := json.Marshal(map[string]interface{}{
            "changes":    result.ChangesDetected,
            "source_urls": result.SourceURLs,
        })
        s.db.Create(&model.ActivityLog{
            UniversityID: &uniID,
            EventType:    model.EventCrawlChange,
            Description:  fmt.Sprintf("%d changes detected", len(result.ChangesDetected)),
            Metadata:     datatypes.JSON(meta),
        })
    }
    return nil
}
```

---

## 9. DTO Layer

### `internal/dto/cases_dto.go`
```go
// Request
type CreateCaseRequest struct {
    FullName               string   `json:"full_name" validate:"required"`
    GpaNormalized          float64  `json:"gpa_normalized" validate:"required,min=0,max=4"`
    GpaRaw                 float64  `json:"gpa_raw"`
    GpaScale               float64  `json:"gpa_scale"`
    IeltsOverall           *float64 `json:"ielts_overall"`
    IeltsBreakdown         *IeltsBreakdown `json:"ielts_breakdown"`
    SatTotal               *int     `json:"sat_total"`
    ToeflTotal             *int     `json:"toefl_total"`
    IntendedMajor          string   `json:"intended_major"`
    BudgetUsdPerYear       int      `json:"budget_usd_per_year" validate:"required,min=0"`
    PreferredCountries     []string `json:"preferred_countries"`
    TargetIntake           string   `json:"target_intake"`
    ScholarshipRequired    bool     `json:"scholarship_required"`
    Extracurriculars       string   `json:"extracurriculars"`
    Achievements           string   `json:"achievements"`
    PersonalStatementNotes string   `json:"personal_statement_notes"`
}

// Validate: phải có ít nhất ielts hoặc sat
func (r *CreateCaseRequest) Validate() error {
    if r.IeltsOverall == nil && r.SatTotal == nil {
        return errors.New("ielts_overall or sat_total is required")
    }
    return nil
}

// Response
type CaseCreatedResponse struct {
    CaseID string `json:"case_id"`
    Status string `json:"status"`
}

type ListCasesResponse struct {
    Data []model.Case `json:"data"`
    Meta response.Meta `json:"meta"`
}

// Callback payload từ AI Service
type JobDonePayload struct {
    JobID        string          `json:"job_id"`
    JobType      string          `json:"job_type"`
    Status       string          `json:"status"`
    CaseID       string          `json:"case_id"`
    UniversityID string          `json:"university_id"`
    Error        *string         `json:"error"`
    Result       json.RawMessage `json:"result"`
}

type AnalyzeResult struct {
    ProfileSummary    interface{}            `json:"profile_summary"`
    Recommendations   []RecommendationResult `json:"recommendations"`
    ConfidenceScore   float64                `json:"confidence_score"`
    EscalationNeeded  bool                   `json:"escalation_needed"`
    EscalationReason  string                 `json:"escalation_reason"`
}

type RecommendationResult struct {
    UniversityID             string   `json:"university_id"`
    UniversityName           string   `json:"university_name"`
    Tier                     string   `json:"tier"`
    AdmissionLikelihoodScore int      `json:"admission_likelihood_score"`
    StudentFitScore          int      `json:"student_fit_score"`
    Reason                   string   `json:"reason"`
    Risks                    []string `json:"risks"`
    Improvements             []string `json:"improvements"`
    RankOrder                int      `json:"rank_order"`
}

type CrawlResult struct {
    Name                    string   `json:"name"`
    Country                 string   `json:"country"`
    QsRank                  *int     `json:"qs_rank"`
    IeltsMin                *float64 `json:"ielts_min"`
    SatRequired             *bool    `json:"sat_required"`
    GpaExpectationNormalized *float64 `json:"gpa_expectation_normalized"`
    TuitionUsdPerYear       *int     `json:"tuition_usd_per_year"`
    ScholarshipAvailable    *bool    `json:"scholarship_available"`
    ScholarshipNotes        *string  `json:"scholarship_notes"`
    ApplicationDeadline     *string  `json:"application_deadline"`
    AvailableMajors         []string `json:"available_majors"`
    AcceptanceRate          *float64 `json:"acceptance_rate"`
    CrawlStatus             string   `json:"crawl_status"`
    ChangesDetected         []string `json:"changes_detected"`
    SourceURLs              []string `json:"source_urls"`
    CrawledAt               string   `json:"crawled_at"`
}
```

---

## 10. Handler Layer

### `internal/handler/cases_handler.go`
```go
type CasesHandler struct {
    svc       service.CaseService
    validator *validator.Validate
}

func (h *CasesHandler) Create(c *gin.Context) {
    var req dto.CreateCaseRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Fail(c, 400, "BAD_REQUEST", err.Error())
        return
    }
    if err := req.Validate(); err != nil {
        response.Fail(c, 400, "VALIDATION_FAILED", err.Error())
        return
    }
    result, appErr := h.svc.Create(c.Request.Context(), req)
    if appErr != nil {
        response.Fail(c, appErr.HTTPStatus, appErr.Code, appErr.Message)
        return
    }
    response.Created(c, result)
}

func (h *CasesHandler) List(c *gin.Context) {
    status := c.Query("status")
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

    result, appErr := h.svc.List(c.Request.Context(), status, page, limit)
    if appErr != nil {
        response.Fail(c, appErr.HTTPStatus, appErr.Code, appErr.Message)
        return
    }
    response.Paginated(c, result.Data, result.Meta)
}

func (h *CasesHandler) GetByID(c *gin.Context) {
    id, err := uuid.Parse(c.Param("id"))
    if err != nil {
        response.Fail(c, 400, "BAD_REQUEST", "invalid case id")
        return
    }
    caseRecord, appErr := h.svc.GetByID(c.Request.Context(), id)
    if appErr != nil {
        response.Fail(c, appErr.HTTPStatus, appErr.Code, appErr.Message)
        return
    }
    response.OK(c, caseRecord)
}

func (h *CasesHandler) RequestReport(c *gin.Context) {
    id, _ := uuid.Parse(c.Param("id"))
    result, appErr := h.svc.RequestReport(c.Request.Context(), id)
    if appErr != nil {
        response.Fail(c, appErr.HTTPStatus, appErr.Code, appErr.Message)
        return
    }
    response.OK(c, result)
}
```

### `internal/handler/internal_handler.go`
```go
// Endpoint callback từ AI Service — không expose ra public
func (h *InternalHandler) JobDone(c *gin.Context) {
    var payload dto.JobDonePayload
    if err := c.ShouldBindJSON(&payload); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    if appErr := h.svc.HandleJobDone(c.Request.Context(), payload); appErr != nil {
        c.JSON(500, gin.H{"error": appErr.Message})
        return
    }
    c.JSON(200, gin.H{"received": true})
}
```

---

## 11. Router

### `internal/router/router.go`
```go
func SetupRouter(
    casesH      *handler.CasesHandler,
    uniH        *handler.UniversitiesHandler,
    dashH       *handler.DashboardHandler,
    internalH   *handler.InternalHandler,
) *gin.Engine {
    r := gin.New()
    r.Use(gin.Recovery())
    r.Use(middleware.CORS())
    r.Use(middleware.Logger())

    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })

    // Internal — AI callback (không cần auth nhưng nên giới hạn theo IP ở production)
    r.POST("/internal/jobs/done", internalH.JobDone)

    api := r.Group("/api/v1")
    {
        cases := api.Group("/cases")
        {
            cases.POST("", casesH.Create)
            cases.GET("", casesH.List)
            cases.GET("/count", casesH.Count)
            cases.GET("/:id", casesH.GetByID)
            cases.GET("/:id/crawl-status", casesH.GetCrawlStatus)
            cases.POST("/:id/report", casesH.RequestReport)
            cases.GET("/:id/report/download", casesH.DownloadReport)
        }

        unis := api.Group("/universities")
        {
            unis.GET("", uniH.List)
            unis.POST("", uniH.Create)
            unis.POST("/crawl-all", uniH.CrawlAll)
            unis.GET("/crawl-active", uniH.CrawlActiveCount)
        }

        dash := api.Group("/dashboard")
        {
            dash.GET("/stats", dashH.Stats)
            dash.GET("/cases-by-day", dashH.CasesByDay)
            dash.GET("/escalation-trend", dashH.EscalationTrend)
            dash.GET("/analytics", dashH.Analytics)
        }

        api.GET("/activity-log", dashH.ActivityLog)
    }

    return r
}
```

### CORS Middleware
```go
func CORS() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Header("Access-Control-Allow-Origin", "http://localhost:5173")
        c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
        c.Header("Access-Control-Allow-Headers", "Content-Type,Authorization")
        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }
        c.Next()
    }
}
```

---

## 12. `main.go` — Wire Everything

```go
func main() {
    cfg := config.Load()

    db, err := database.NewPostgres(cfg.DatabaseURL)
    if err != nil {
        log.Fatalf("failed to connect db: %v", err)
    }

    // Seed nếu cần
    // seedUniversities(db)

    aiClient := client.NewAIClient(cfg.AIServiceURL)

    // Repositories
    caseRepo := repository.NewCaseRepository(db)
    uniRepo  := repository.NewUniversityRepository(db)
    actRepo  := repository.NewActivityRepository(db)
    dashRepo := repository.NewDashboardRepository(db)

    // Services
    validate  := validator.New()
    caseSvc   := service.NewCaseService(db, caseRepo, uniRepo, actRepo, aiClient, cfg, validate)
    uniSvc    := service.NewUniversityService(db, uniRepo, actRepo, aiClient, cfg)
    dashSvc   := service.NewDashboardService(dashRepo, actRepo)

    // Handlers
    casesH    := handler.NewCasesHandler(caseSvc)
    uniH      := handler.NewUniversitiesHandler(uniSvc)
    dashH     := handler.NewDashboardHandler(dashSvc)
    internalH := handler.NewInternalHandler(caseSvc, uniSvc)

    r := router.SetupRouter(casesH, uniH, dashH, internalH)
    log.Printf("server running on :%s", cfg.Port)
    r.Run(":" + cfg.Port)
}
```

---

## 13. Build Sequence (Hackathon Order)

| Bước | Task | Thời gian |
|------|------|-----------|
| 1 | Scaffold + `go mod`, folder structure | 15m |
| 2 | `config.go`, `pkg/database`, `AutoMigrate` models | 20m |
| 3 | `pkg/response`, `pkg/apperror`, `pkg/client/ai_client` | 20m |
| 4 | Models đầy đủ (GORM tags, TableName, constants) | 20m |
| 5 | Repository interfaces + implementations | 30m |
| 6 | DTO structs (request + response + callback payloads) | 20m |
| 7 | **Cases service** — `Create` + `HandleJobDone` | 45m |
| 8 | **Universities service** — `CrawlAll`, crawl trigger | 20m |
| 9 | Handlers → thin, chỉ validate + call service | 30m |
| 10 | Router + CORS middleware + `main.go` wire | 15m |
| 11 | Dashboard repository (raw SQL stats queries) | 20m |
| 12 | PDF download endpoint (`go-fpdf`) | 20m |
| 13 | Seed data, smoke test với curl | 15m |
| **Total** | | **~5h** |

---

## 14. Key GORM Patterns (tuân thủ backend SKILL)

```go
// ✅ Luôn dùng WithContext
db.WithContext(ctx).Find(&results)

// ✅ Transaction cho multi-table write
db.Transaction(func(tx *gorm.DB) error { ... })

// ✅ Preload thay vì JOIN thủ công
db.Preload("Student").Preload("Recommendations").First(&c, id)

// ✅ Updates với map để partial update (tránh zero-value override)
db.Model(&model.Case{}).Where("id = ?", id).Updates(map[string]interface{}{
    "status": "done",
    "processing_finished_at": &now,
})

// ✅ GORM Scopes cho reusable query
func WithStatus(status string) func(*gorm.DB) *gorm.DB {
    return func(db *gorm.DB) *gorm.DB {
        if status != "" { return db.Where("status = ?", status) }
        return db
    }
}
db.Scopes(WithStatus(status)).Find(&cases)

// ❌ KHÔNG dùng raw string concat
db.Where("status = " + status) // SQL injection risk
```

---

## 15. Environment & Run

```bash
# Start PostgreSQL
docker run -d --name pg-be -e POSTGRES_PASSWORD=password -e POSTGRES_DB=unimatch_be -p 5432:5432 postgres:15

# Run
go run main.go

# Verify
curl http://localhost:8080/health
# → {"status":"ok"}

curl -X POST http://localhost:8080/api/v1/cases \
  -H "Content-Type: application/json" \
  -d '{"full_name":"Test","gpa_normalized":3.5,"ielts_overall":7.0,"intended_major":"CS","budget_usd_per_year":30000,"preferred_countries":["UK"],"target_intake":"Fall 2026"}'
```

---

## 16. Contract với AI Service (tóm tắt)

| Direction | Endpoint | Payload key |
|-----------|----------|-------------|
| BE → AI | `POST /jobs/crawl` | `job_id, university_id, callback_url, metadata` |
| BE → AI | `POST /jobs/analyze` | `job_id, case_id, callback_url, input` |
| BE → AI | `POST /jobs/report` | `job_id, case_id, callback_url, recommendations` |
| AI → BE | `POST /internal/jobs/done` | `job_id, job_type, status, case_id/university_id, result` |

**AI trả về `{ "accepted": true }` ngay lập tức.** BE không block chờ kết quả.

---

## 17. apperror Package

```go
type AppError struct {
    Code       string
    Message    string
    Details    string
    Err        error
    HTTPStatus int
}

func NotFound(msg string) *AppError    { return &AppError{Code: "NOT_FOUND", Message: msg, HTTPStatus: 404} }
func BadRequest(msg string) *AppError  { return &AppError{Code: "BAD_REQUEST", Message: msg, HTTPStatus: 400} }
func Internal(err error, msg string) *AppError {
    return &AppError{Code: "INTERNAL_ERROR", Message: msg, Err: err, HTTPStatus: 500}
}
```
