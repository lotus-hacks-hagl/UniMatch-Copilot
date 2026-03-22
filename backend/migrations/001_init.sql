-- Manual Migration Reference
-- GORM AutoMigrate sẽ lo schema generation, file này thuần tuý cho document schema mục đích review

CREATE TABLE students (
    id UUID PRIMARY KEY,
    full_name VARCHAR(255) NOT NULL,
    gpa_normalized FLOAT,
    gpa_raw FLOAT,
    gpa_scale FLOAT,
    ielts_overall FLOAT,
    ielts_breakdown JSONB,
    sat_total INT,
    toefl_total INT,
    intended_major VARCHAR(255),
    budget_usd_per_year INT,
    preferred_countries TEXT[],
    target_intake VARCHAR(50),
    scholarship_required BOOLEAN,
    extracurriculars TEXT,
    achievements TEXT,
    personal_statement_notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE cases (
    id UUID PRIMARY KEY,
    student_id UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    ai_job_id UUID,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    ai_confidence FLOAT,
    escalation_reason TEXT,
    profile_summary JSONB,
    report_data JSONB,
    processing_started_at TIMESTAMP WITH TIME ZONE,
    processing_finished_at TIMESTAMP WITH TIME ZONE,
    report_generated_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE universities (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    country VARCHAR(100),
    qs_rank INT,
    group_tag VARCHAR(100),
    ielts_min FLOAT,
    sat_required BOOLEAN DEFAULT FALSE,
    gpa_expectation_normalized FLOAT,
    tuition_usd_per_year INT,
    scholarship_available BOOLEAN DEFAULT FALSE,
    scholarship_notes TEXT,
    application_deadline TIMESTAMP WITH TIME ZONE,
    available_majors TEXT[],
    acceptance_rate FLOAT,
    counselor_notes TEXT,
    crawl_status VARCHAR(50) DEFAULT 'never_crawled',
    crawl_job_id UUID,
    last_crawled_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE recommendations (
    id UUID PRIMARY KEY,
    case_id UUID NOT NULL REFERENCES cases(id) ON DELETE CASCADE,
    university_id UUID NOT NULL REFERENCES universities(id) ON DELETE CASCADE,
    university_name VARCHAR(255) NOT NULL,
    tier VARCHAR(50) NOT NULL,
    admission_likelihood_score INT NOT NULL,
    student_fit_score INT NOT NULL,
    reason TEXT,
    risks JSONB,
    improvements JSONB,
    rank_order INT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE activity_logs (
    id UUID PRIMARY KEY,
    case_id UUID REFERENCES cases(id) ON DELETE CASCADE,
    university_id UUID REFERENCES universities(id) ON DELETE CASCADE,
    event_type VARCHAR(100) NOT NULL,
    description TEXT,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Indexes (Created by GORM but documented here)
CREATE INDEX idx_students_countries ON students USING GIN (preferred_countries);
CREATE INDEX idx_cases_status ON cases(status);
CREATE INDEX idx_cases_student_id ON cases(student_id);
CREATE INDEX idx_universities_country ON universities(country);
CREATE INDEX idx_universities_crawl_status ON universities(crawl_status);
CREATE INDEX idx_recommendations_case_id ON recommendations(case_id);
CREATE INDEX idx_activity_logs_case_id ON activity_logs(case_id);
CREATE INDEX idx_activity_logs_event_type ON activity_logs(event_type);
