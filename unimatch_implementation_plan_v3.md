# UniMatch Copilot — Implementation Plan v3

> Hackathon 24h · 3 người: **Person A (FE)**, **Person B (BE)**, **Person C (AI Service)**
> Stack: Vue3 + Tailwind · Go (Gin) + PostgreSQL + Redis · Python (FastAPI) + Claude Agent SDK + TinyFish MCP + **mcp-neo4j-cypher**

> **v3 changes (AI Service):** Thay `neo4j-graphrag-python` bằng `mcp-neo4j-cypher` MCP server (neo4j-contrib). Crawl worker dùng `claude-agent-sdk` với **2 MCP servers đồng thời** (TinyFish + mcp-neo4j-cypher) — agent tự gọi `run_web_automation` để crawl web rồi tự gọi `write_neo4j_cypher` để ghi graph. Analyze worker dùng `read_neo4j_cypher` qua MCP hoặc raw neo4j driver. Không cần `neo4j-graphrag`, không cần OpenAI embeddings.

---

## Kiến trúc tổng quan

```
[Vue3 FE :5173]
      │ REST poll
      ▼
[Go BE :8080]  ──── fire & forget ──▶  [Python AI Service :9000]
      │                                        │  asyncio.Queue
      │                                        │  ┌───────────────────────────────┐
  [PG BE]                                      │  │  crawl_worker                 │
  students                                     │  │    └─ Claude Agent SDK        │
  cases                                        │  │         └─ TinyFish MCP tools │
  recommendations                              │  │              run_web_automation  │
  activity_log                                 │  │              (TinyFish MCP)      │
                                               │  │              write_neo4j_cypher  │
                                               │  │              (mcp-neo4j-cypher)  │
                                               │  │  analyze_worker               │
                                               │  │    └─ read_neo4j_cypher (MCP) │
                                               │  │    └─ Claude (matching)       │
                                               │  └───────────────────────────────┘
                                               │        │ read/write
                                           [Neo4j :7687]┘        [PG AI :5433]
                                           Knowledge Graph         jobs · crawl_cache
                                           (mcp-neo4j-cypher)
                                           University nodes
                                           Program / AdmissionReq / Scholarship
                                           agent writes Cypher directly
      ▲
      └──── callback POST /internal/jobs/done ────────────────────────────────────┘
```

### Nguyên tắc thiết kế

1. **BE là orchestrator, AI Service là worker.** BE không biết Claude, TinyFish hay Neo4j tồn tại.
2. **3 storage độc lập.** BE DB (PostgreSQL) lưu business data. AI DB (PostgreSQL) lưu job state. Knowledge Graph (Neo4j) lưu entities + relationships do agent tự detect.
3. **Mọi job đều async.** BE gửi request → nhận `accepted: true` ngay → AI xử lý ngầm → callback về BE khi xong.
4. **BE gửi university metadata đầy đủ với null fields.** AI agent tự quyết dùng tool nào, crawl bao nhiêu lần, detect entity gì — fully autonomous. Fixed-schema fields được fill về BE. Knowledge graph được xây song song.
5. **Callback trả đúng schema request, chỉ fill field nào tìm được.** Field nào agent không tìm thấy → giữ nguyên `null`. BE tự xử lý null.
6. **Delete university → AI xóa toàn bộ nodes liên quan trong graph.** `DELETE /jobs/university/:id` xóa university node + tất cả connected nodes không có university khác reference đến.

---

## Shared Contracts (đọc trước khi code — 3 người đều phải nắm)

### Job types

```
JOB_TYPE_CRAWL   = "crawl_university"
JOB_TYPE_ANALYZE = "analyze_profile"
JOB_TYPE_REPORT  = "generate_report"
```

### Job status

```
pending → processing → done
                    → failed
```

### API: BE → AI Service (fire & forget)

#### POST /jobs/crawl
```json
// Request (BE → AI)
// BE gửi toàn bộ university metadata hiện có.
// Các field = null là những field cần agent detect và fill.
// Agent sẽ research, fill những gì tìm được, giữ null những gì không tìm thấy.
{
  "job_id": "uuid-from-be",
  "university_id": "uuid",        // BE's ID, dùng để callback + graph node key
  "callback_url": "http://localhost:8080/internal/jobs/done",
  "metadata": {
    "name": "TU Delft",           // luôn có
    "country": "Netherlands",     // luôn có
    "qs_rank": null,              // → agent detect
    "ielts_min": null,            // → agent detect
    "sat_required": null,         // → agent detect
    "gpa_expectation_normalized": null,  // → agent detect
    "tuition_usd_per_year": null, // → agent detect (convert sang USD)
    "scholarship_available": null,// → agent detect
    "scholarship_notes": null,    // → agent detect
    "application_deadline": null, // → agent detect (YYYY-MM-DD)
    "available_majors": null,     // → agent detect (string array)
    "acceptance_rate": null       // → agent detect (0.0–1.0)
  }
}

// Response ngay lập tức (AI → BE, chỉ acknowledge)
{ "accepted": true, "job_id": "uuid-from-be" }
```

**Callback khi crawl xong** — `POST /internal/jobs/done`:
```json
{
  "job_id": "uuid",
  "job_type": "crawl_university",
  "status": "done",              // hoặc "failed"
  "university_id": "uuid",
  "error": null,
  "result": {
    // Same schema as metadata in request.
    // Chỉ fill field nào agent tìm được. Không tìm được → null (giữ nguyên).
    "name": "TU Delft",
    "country": "Netherlands",
    "qs_rank": 57,
    "ielts_min": 7.0,            // null nếu không tìm thấy
    "sat_required": false,
    "gpa_expectation_normalized": 3.2,
    "tuition_usd_per_year": 19800,
    "scholarship_available": true,
    "scholarship_notes": "Holland Scholarship available",
    "application_deadline": "2026-04-15",
    "available_majors": ["CS", "Aerospace", "Civil Engineering"],
    "acceptance_rate": 0.32,
    // Extra: thay đổi so với lần crawl trước (so sánh với value BE gửi lên)
    "crawl_status": "changed",   // "ok" | "changed" | "partial" | "failed"
    "changes_detected": ["ielts_min: 6.5 → 7.0"],
    "source_urls": ["https://www.tudelft.nl/..."],
    "crawled_at": "2026-03-21T08:12:00Z"
  }
}
```

#### POST /jobs/analyze
```json
// Request (BE → AI)
{
  "job_id": "uuid-from-be",
  "case_id": "uuid",             // BE's case ID, dùng để callback
  "callback_url": "http://localhost:8080/internal/jobs/done",
  "input": {
    "full_name": "Nguyen Linh",
    "gpa_normalized": 3.6,       // đã convert sang 4.0 scale, BE tự tính
    "ielts_overall": 7.0,
    "sat_total": null,
    "intended_major": "Computer Science",
    "budget_usd_per_year": 35000,
    "preferred_countries": ["UK", "NL"],
    "target_intake": "Fall 2026",
    "scholarship_required": false,
    "extracurriculars": "Hackathon winner x2, research assistant",
    "achievements": "Dean's list 2023"
  }
}

// Response ngay lập tức
{ "accepted": true, "job_id": "uuid-from-be" }
```

#### POST /jobs/report
```json
// Request (BE → AI)
{
  "job_id": "uuid-from-be",
  "case_id": "uuid",
  "callback_url": "http://localhost:8080/internal/jobs/done",
  "student_name": "Nguyen Linh",
  "recommendation_ids": ["rec-uuid-1", "rec-uuid-2"]
  // AI tự lấy data từ KB của nó dựa vào university_id trong recommendations
  // BE sẽ cũng gửi kèm recommendations đã có
  "recommendations": [ /* xem Recommendation Object bên dưới */ ]
}
```

### API: AI Service → BE (callback)

Crawl callback đã được định nghĩa inline ở `POST /jobs/crawl` ở trên.

#### Analyze callback — POST /internal/jobs/done
{
  "job_id": "uuid",
  "job_type": "analyze_profile",
  "status": "done",
  "case_id": "uuid",
  "error": null,
  "result": {
    "profile_summary": {
      "academic_strength": "strong",
      "language_readiness": "ready",
      "budget_band": "mid",
      "scholarship_sensitivity": false,
      "strengths": ["Strong GPA trend", "Research experience", "2× hackathon winner"],
      "weaknesses": ["IELTS borderline for top-5 UK", "No SAT score"],
      "risk_tolerance": "moderate"
    },
    "recommendations": [
      {
        "university_id": "uuid",         // BE's university ID
        "university_name": "University of Edinburgh",
        "tier": "safe",
        "admission_likelihood_score": 82,
        "student_fit_score": 91,
        "reason": "GPA và IELTS vượt ngưỡng tối thiểu...",
        "risks": ["Budget sát nếu không có scholarship"],
        "improvements": ["Cân nhắc retake IELTS lên 7.5"],
        "rank_order": 1
      }
    ],
    "confidence_score": 0.91,
    "escalation_needed": false,
    "escalation_reason": null
  }
}

// Report callback
{
  "job_id": "uuid",
  "job_type": "generate_report",
  "status": "done",
  "case_id": "uuid",
  "error": null,
  "result": {
    "executive_summary": "Nguyen Linh có profile tốt...",
    "top_recommendation_rationale": "TU Delft là best fit vì...",
    "action_items": ["Retake IELTS trước tháng 4", "Chuẩn bị personal statement"],
    "report_sections": {
      "profile_analysis": "...",
      "university_list": "...",
      "timeline_advice": "..."
    }
  }
}
```

### University Object (dùng ở cả BE và FE)

```json
{
  "id": "uuid",
  "name": "TU Delft",
  "country": "Netherlands",
  "qs_rank": 57,
  "group_tag": "Technical University",
  "ielts_min": 7.0,
  "sat_required": false,
  "gpa_expectation_normalized": 3.2,
  "tuition_usd_per_year": 19800,
  "scholarship_available": true,
  "scholarship_notes": "Holland Scholarship",
  "available_majors": ["CS", "Aerospace"],
  "application_deadline": "2026-04-15",
  "acceptance_rate": 0.32,
  "crawl_status": "ok|pending|changed|failed",
  "last_crawled_at": "2026-03-21T08:12:00Z",
  "counselor_notes": ""
}
```

### Recommendation Object

```json
{
  "id": "uuid",
  "case_id": "uuid",
  "university_id": "uuid",
  "university_name": "TU Delft",
  "tier": "safe|match|reach",
  "admission_likelihood_score": 82,
  "student_fit_score": 91,
  "reason": "...",
  "risks": ["..."],
  "improvements": ["..."],
  "rank_order": 1,
  "created_at": "..."
}
```

---

## Person A — Frontend (Vue3 + Tailwind)

### A-0: Project scaffold (30m — làm đầu tiên)

- `npm create vue@latest unimatch-fe` — Vue3, Vite, Vue Router, Pinia
- `npm install axios tailwindcss @tailwindcss/forms`
- Cấu trúc thư mục:
  ```
  src/
    views/
    components/
    stores/
    composables/
    services/api.js
  ```
- `src/services/api.js`:
  ```js
  import axios from 'axios'
  export const api = axios.create({
    baseURL: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1',
    headers: { 'Content-Type': 'application/json' }
  })
  ```
- `.env.local`: `VITE_API_BASE_URL=http://localhost:8080/api/v1`

---

### A-1: Layout Shell + Sidebar (45m)

**File:** `src/components/AppShell.vue`

- Layout 2 cột: sidebar 210px + main flex-1
- Sidebar nav items: Cases → `/cases`, Students → `/students`, Review queue → `/queue`, University KB → `/kb`, Analytics → `/analytics`
- Badge đỏ "3" trên Review Queue: lấy từ `useQueueStore().pendingCount`, gọi `GET /api/v1/cases/count?status=human_review` khi mount
- Topbar:
  - Title + sub-title từ `route.meta.title` và `route.meta.sub`
  - Live pill "TinyFish syncing N universities": `GET /api/v1/universities/crawl-active` poll mỗi 5s, ẩn nếu count = 0, có animation pulse dot
  - Nút "+ New case" → `router.push('/cases/new')`
- User info footer: hardcode "Trang Nguyen / Senior counselor" cho demo

---

### A-2: Cases List Page (90m)

**File:** `src/views/CasesView.vue`
**Route meta:** `{ title: 'Case overview', sub: 'Saturday, 21 Mar 2026 — 47 active cases' }`

**Metric cards** — gọi `GET /api/v1/dashboard/stats`:
- Cases today
- Avg. processing time (format: "4m")
- Awaiting human review (highlight đỏ nếu > 0)
- AI confidence avg. (format: "87%")

**Filter tabs** — All / Done / Processing / Human review:
- Click tab → gọi `GET /api/v1/cases?status=<filter>&page=1&limit=20`
- Lưu filter vào `useCasesStore().activeFilter`

**Table columns:**
- Student: avatar (2 chữ đầu + màu từ `hashColor(student_id)`) + tên + GPA/IELTS
- Profile snapshot: major · country · budget
- Target intake: "Fall 2026"
- AI match tiers: badges Safe/Match/Reach với số lượng
- Confidence: progress bar + %
  - xanh nếu ≥ 80%, vàng nếu 60–79%, đỏ nếu < 60%
- Status badge: Done / Processing / Human review / Pending
- Updated: relative time ("2 min ago")
- Action: "View →" link

Click row → `router.push('/cases/' + case.id)`

**Charts row** — 3 chart cards:
- Bar chart cases/day: `GET /api/v1/dashboard/cases-by-day` (7 ngày)
- Donut match tier: tính % từ stats response (safe_count, match_count, reach_count)
- Sparkline escalation trend: 6 bars từ `GET /api/v1/dashboard/escalation-trend`

**Pinia store** `src/stores/casesStore.js`:
```js
state: () => ({ cases: [], loading: false, filter: 'all', stats: {}, total: 0 })
actions: { async fetchCases(filter), async fetchStats() }
```

---

### A-3: New Case Form (60m)

**File:** `src/views/NewCaseView.vue`
**Route:** `/cases/new`

Multi-step form, local state `step = ref(1)`, progress bar 3 bước.

**Step 1 — Scores:**
- `full_name` (required, text)
- `gpa_raw` (required, number) + `gpa_scale` (dropdown: 4.0 / 10 / 100)
- `ielts_overall` (number 0–9, step 0.5, optional nếu có SAT)
- `ielts_breakdown`: listening, reading, writing, speaking (number, optional, show/hide toggle)
- `sat_total` (number, optional)
- `toefl_total` (number, optional)
- Validate: phải có ít nhất ielts_overall hoặc sat_total

**Step 2 — Preferences:**
- `intended_major` (text + autocomplete: CS, Engineering, Finance, Law, Medicine, Architecture, Business)
- `budget_usd_per_year` (slider $10k–$80k, step $5k, hiển thị "$35,000/year")
- `preferred_countries` (multi-checkbox: UK, US, AUS, NL, DE, SG, CA, FR)
- `target_intake` (dropdown: Fall 2026, Spring 2027, Fall 2027)
- `scholarship_required` (toggle)

**Step 3 — Background:**
- `extracurriculars` (textarea)
- `achievements` (textarea)
- `personal_statement_notes` (textarea)

**Submit — `POST /api/v1/cases`:**
- FE tự tính `gpa_normalized` trước khi gửi:
  ```js
  function normalizeGpa(raw, scale) {
    return Math.round((raw / scale) * 4.0 * 100) / 100
  }
  ```
- Gửi payload, nhận `{ case_id, status: 'pending' }`
- Navigate ngay sang `/cases/:case_id` với `router.push`

---

### A-4: Case Detail Page (90m)

**File:** `src/views/CaseDetailView.vue`
**Route:** `/cases/:id`

**Polling logic** (dùng `usePolling` composable):
- Gọi `GET /api/v1/cases/:id` mỗi 3s khi `case.status === 'pending' || 'processing'`
- Stop polling khi status = `done` hoặc `human_review`

**Breadcrumb:** "← Cases / Nguyen Linh" — click "Cases" → `router.push('/cases')`

**TinyFish crawl panel** (hiển thị khi status đang processing):
- `GET /api/v1/cases/:id/crawl-status` poll mỗi 3s riêng
- List universities đang crawl:
  - `crawl_status = 'pending'`: spinner icon + tên + "Fetching..."
  - `crawl_status = 'ok'`: check xanh + tên + "IELTS min · tuition · deadline" + tag "No change"
  - `crawl_status = 'changed'`: warning icon + tên + diff text (strikethrough old → new xanh) + tag "Req. changed"
  - `crawl_status = 'failed'`: red X + tên + "Could not fetch" + tag "Failed"
- Live pill "Crawling now" ở header crawl panel

**Profile card (trái 280px):**
- Avatar lớn + tên + major + intake
- Score grid 2×2: GPA (hiển thị normalized/4.0), IELTS, Budget ($format), Countries
- Strengths tags (từ `profile_summary.strengths`)
- Risks tags màu đỏ nhạt (từ `profile_summary.weaknesses`)
- Confidence bar:
  - Màu theo giá trị (xanh/vàng/đỏ)
  - Text: "Auto-approved" nếu không escalate, "Escalated to review" nếu có

**Recommendations (phải):**
- Khi `status = 'processing'` hoặc `status = 'pending'`: hiển thị skeleton 3 cards loading
- Khi `status = 'done' || 'human_review'`: render recommendation cards
- Sắp xếp theo `rank_order`
- Card "best fit" = card rank_order = 1 trong tier "match": `border-color: blue, border-width: 1.5px`
- Mỗi card:
  - Header: tên trường + tier badge
  - Meta row: IELTS, budget fit, QS rank (màu amber nếu borderline)
  - Progress bar `admission_likelihood_score`
  - Lý do AI (text `reason`)
- 2 nút dưới cùng:
  - "Generate PDF report" → `POST /api/v1/cases/:id/report` → polling status report → khi done: `window.open(report_url)`
  - "Send to student" (UI only, disabled, tooltip "Coming soon")

---

### A-5: Review Queue Page (45m)

**File:** `src/views/ReviewQueueView.vue`

- `GET /api/v1/cases?status=human_review` khi mount
- 3 metric cards: In queue, Avg. AI confidence, Avg. review time

**Queue list:**
- Mỗi row: avatar + tên + major/country, scores + confidence, escalation_reason (màu amber), SLA pill
- SLA tính: `const slaDeadline = new Date(case.created_at).getTime() + 8 * 3600 * 1000`
  - Hiển thị "Xh Ym left", đỏ nếu < 2h, vàng nếu < 4h
- Nút "Review now" / "Review" → navigate `/cases/:id`

**Activity log:**
- `GET /api/v1/activity-log?limit=10`
- Mỗi item: timestamp, dot màu, text
  - `event_type = 'auto_approved'`: dot xanh
  - `event_type = 'escalated'`: dot vàng
  - `event_type = 'crawl_change'`: dot xanh dương
  - `event_type = 'case_created'`: dot gray

---

### A-6: University KB Page (45m)

**File:** `src/views/UniversityKBView.vue`

- `GET /api/v1/universities?page=1&limit=20`
- 3 metric cards: Total, Data freshness %, Changes this month
- Nút "Run TinyFish crawl": `POST /api/v1/universities/crawl-all` → toast "Crawl started for N universities"
- Nút "+ Add university": mở modal form (name, country, IELTS min, tuition, majors text)

**Table:**
- Tên/rank tag, Country emoji + name, IELTS min, GPA expect, Tuition/yr, Scholarship badge, Last crawled, Status tag
- Row highlight vàng nhạt nếu `crawl_status = 'changed'`
- IELTS cell: nếu changed → strikethrough old + new xanh
- Status tag: "Up to date" (gray) / "Req. changed" (amber) / "Scholarship added" (green) / "Pending" (blue) / "Crawling..." (pulse)

---

### A-7: Analytics Page (45m)

**File:** `src/views/AnalyticsView.vue`

- `GET /api/v1/dashboard/analytics`
- 4 metric cards: Cases/month, Avg. time/case, Auto-approved %, Hours saved
- 2×2 analytics grid:
  - ROI card: big number + 3 sub-metrics (hours freed, avg time, scale capacity)
  - Cases/week bar chart (6 bars, SVG inline)
  - AI confidence distribution bar (5 ranges)
  - TinyFish impact list (4 rows + highlight box xanh)

---

### A-8: Composables (30m)

**`src/composables/usePolling.js`:**
```js
export function usePolling(fn, ms = 3000) {
  let timer = null
  const start = () => { fn(); timer = setInterval(fn, ms) }
  const stop = () => clearInterval(timer)
  onUnmounted(stop)
  return { start, stop }
}
```

**`src/composables/useToast.js`:**
- Simple toast store: `{ message, type: 'success'|'error'|'info', show }`
- Auto-dismiss sau 4s
- Render ở `AppShell.vue` top-right

**`src/composables/useHashColor.js`:**
```js
const COLORS = ['blue','teal','amber','green','coral']
export function hashColor(id) {
  let hash = 0
  for (const c of id) hash = ((hash << 5) - hash) + c.charCodeAt(0)
  return COLORS[Math.abs(hash) % COLORS.length]
}
```

---

## Person B — Backend (Go + Gin + PostgreSQL + Redis)

### B-0: Project scaffold (30m — làm đầu tiên)

```bash
go mod init unimatch-be
go get github.com/gin-gonic/gin
go get github.com/lib/pq
go get github.com/go-redis/redis/v8
go get github.com/google/uuid
```

Cấu trúc:
```
cmd/server/main.go
internal/
  handler/
    cases.go
    universities.go
    dashboard.go
    internal.go      # callback endpoint từ AI Service
  service/
    cases.go
    universities.go
    dashboard.go
  repository/
    cases.go
    universities.go
    activity.go
  model/
    models.go
  middleware/
    cors.go
    logger.go
  client/
    ai_client.go     # HTTP client gọi AI Service
config/config.go
migrations/
```

CORS: cho phép `http://localhost:5173`.

---

### B-1: Database Schema (45m)

**`migrations/001_init.sql`:**

```sql
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE students (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  full_name VARCHAR(255) NOT NULL,
  gpa_raw DECIMAL(5,2),
  gpa_scale DECIMAL(5,2) DEFAULT 10,
  gpa_normalized DECIMAL(4,2),
  ielts_overall DECIMAL(3,1),
  ielts_breakdown JSONB,
  sat_total INT,
  toefl_total INT,
  intended_major VARCHAR(255),
  budget_usd_per_year INT,
  preferred_countries TEXT[],
  target_intake VARCHAR(50),
  scholarship_required BOOLEAN DEFAULT FALSE,
  extracurriculars TEXT,
  achievements TEXT,
  personal_statement_notes TEXT,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE cases (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  student_id UUID REFERENCES students(id) ON DELETE CASCADE,
  status VARCHAR(50) DEFAULT 'pending',
  -- status values: pending | processing | done | human_review | failed
  ai_job_id UUID,                          -- job_id gửi sang AI Service
  ai_confidence DECIMAL(4,3),
  escalation_reason TEXT,
  profile_summary JSONB,                   -- lưu lại profile_summary từ AI callback
  report_data JSONB,                       -- lưu kết quả generate_report
  report_generated_at TIMESTAMPTZ,
  processing_started_at TIMESTAMPTZ,
  processing_finished_at TIMESTAMPTZ,
  created_by VARCHAR(255) DEFAULT 'counselor',
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE universities (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  name VARCHAR(255) NOT NULL,
  country VARCHAR(100),
  qs_rank INT,
  group_tag VARCHAR(100),
  ielts_min DECIMAL(3,1),
  sat_required BOOLEAN DEFAULT FALSE,
  gpa_expectation_normalized DECIMAL(4,2),
  tuition_usd_per_year INT,
  scholarship_available BOOLEAN DEFAULT FALSE,
  scholarship_notes TEXT,
  available_majors TEXT[],
  application_deadline DATE,
  acceptance_rate DECIMAL(4,3),
  crawl_status VARCHAR(50) DEFAULT 'ok',
  -- crawl_status: ok | pending | changed | failed | never_crawled
  last_crawled_at TIMESTAMPTZ,
  crawl_job_id UUID,                       -- job_id gửi sang AI Service
  counselor_notes TEXT,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE recommendations (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  case_id UUID REFERENCES cases(id) ON DELETE CASCADE,
  university_id UUID REFERENCES universities(id),
  university_name VARCHAR(255),            -- denormalize để tránh JOIN
  tier VARCHAR(20),                        -- safe | match | reach
  admission_likelihood_score INT,
  student_fit_score INT,
  reason TEXT,
  risks JSONB,
  improvements JSONB,
  rank_order INT,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE activity_log (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  case_id UUID REFERENCES cases(id) ON DELETE SET NULL,
  university_id UUID REFERENCES universities(id) ON DELETE SET NULL,
  event_type VARCHAR(100),
  -- event types: case_created | processing_started | auto_approved |
  --              escalated | crawl_started | crawl_change | crawl_done | report_generated
  description TEXT,
  metadata JSONB,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_cases_status ON cases(status);
CREATE INDEX idx_cases_student_id ON cases(student_id);
CREATE INDEX idx_cases_created_at ON cases(created_at DESC);
CREATE INDEX idx_recommendations_case_id ON recommendations(case_id);
CREATE INDEX idx_universities_crawl_status ON universities(crawl_status);
CREATE INDEX idx_activity_log_created_at ON activity_log(created_at DESC);
CREATE INDEX idx_activity_log_case_id ON activity_log(case_id);
```

**`migrations/002_seed.sql`:**

```sql
INSERT INTO universities (name, country, qs_rank, ielts_min, tuition_usd_per_year,
  scholarship_available, available_majors, application_deadline, acceptance_rate)
VALUES
('University of Edinburgh','UK',27,6.5,35000,true,
  ARRAY['CS','Engineering','Law','Medicine'],'2026-03-31',0.18),
('TU Delft','Netherlands',57,6.5,18000,true,
  ARRAY['CS','Aerospace','Civil Engineering'],'2026-04-15',0.32),
('Imperial College London','UK',6,7.0,45000,true,
  ARRAY['CS','Engineering','Medicine'],'2026-01-15',0.14),
('University of Melbourne','Australia',33,6.5,38000,false,
  ARRAY['CS','Business','Law'],'2026-05-01',0.25),
('NUS Singapore','Singapore',8,6.5,26000,true,
  ARRAY['CS','Engineering','Business'],'2026-03-01',0.17);
```

---

### B-2: Go Models (20m)

**`internal/model/models.go`:**

```go
type Student struct {
    ID                   string          `json:"id" db:"id"`
    FullName             string          `json:"full_name" db:"full_name"`
    GpaRaw               float64         `json:"gpa_raw" db:"gpa_raw"`
    GpaScale             float64         `json:"gpa_scale" db:"gpa_scale"`
    GpaNormalized        float64         `json:"gpa_normalized" db:"gpa_normalized"`
    IeltsOverall         *float64        `json:"ielts_overall" db:"ielts_overall"`
    IeltsBreakdown       json.RawMessage `json:"ielts_breakdown" db:"ielts_breakdown"`
    SatTotal             *int            `json:"sat_total" db:"sat_total"`
    ToeflTotal           *int            `json:"toefl_total" db:"toefl_total"`
    IntendedMajor        string          `json:"intended_major" db:"intended_major"`
    BudgetUsdPerYear     int             `json:"budget_usd_per_year" db:"budget_usd_per_year"`
    PreferredCountries   pq.StringArray  `json:"preferred_countries" db:"preferred_countries"`
    TargetIntake         string          `json:"target_intake" db:"target_intake"`
    ScholarshipRequired  bool            `json:"scholarship_required" db:"scholarship_required"`
    Extracurriculars     string          `json:"extracurriculars" db:"extracurriculars"`
    Achievements         string          `json:"achievements" db:"achievements"`
    PersonalStatementNotes string        `json:"personal_statement_notes" db:"personal_statement_notes"`
    CreatedAt            time.Time       `json:"created_at" db:"created_at"`
}

type Case struct {
    ID                     string          `json:"id"`
    StudentID              string          `json:"student_id"`
    Student                *Student        `json:"student,omitempty"`
    Status                 string          `json:"status"`
    AiJobID                *string         `json:"ai_job_id"`
    AiConfidence           *float64        `json:"ai_confidence"`
    EscalationReason       *string         `json:"escalation_reason"`
    ProfileSummary         json.RawMessage `json:"profile_summary"`
    ReportData             json.RawMessage `json:"report_data,omitempty"`
    ProcessingStartedAt    *time.Time      `json:"processing_started_at"`
    ProcessingFinishedAt   *time.Time      `json:"processing_finished_at"`
    CreatedAt              time.Time       `json:"created_at"`
    UpdatedAt              time.Time       `json:"updated_at"`
    Recommendations        []Recommendation `json:"recommendations,omitempty"`
}

type University struct {
    ID                       string         `json:"id"`
    Name                     string         `json:"name"`
    Country                  string         `json:"country"`
    QsRank                   *int           `json:"qs_rank"`
    GroupTag                 string         `json:"group_tag"`
    IeltsMin                 *float64       `json:"ielts_min"`
    SatRequired              bool           `json:"sat_required"`
    GpaExpectationNormalized *float64       `json:"gpa_expectation_normalized"`
    TuitionUsdPerYear        *int           `json:"tuition_usd_per_year"`
    ScholarshipAvailable     bool           `json:"scholarship_available"`
    ScholarshipNotes         string         `json:"scholarship_notes"`
    AvailableMajors          pq.StringArray `json:"available_majors"`
    ApplicationDeadline      *time.Time     `json:"application_deadline"`
    AcceptanceRate           *float64       `json:"acceptance_rate"`
    CrawlStatus              string         `json:"crawl_status"`
    LastCrawledAt            *time.Time     `json:"last_crawled_at"`
    CrawlJobID               *string        `json:"crawl_job_id"`
    CounselorNotes           string         `json:"counselor_notes"`
    CreatedAt                time.Time      `json:"created_at"`
}

type Recommendation struct {
    ID                      string          `json:"id"`
    CaseID                  string          `json:"case_id"`
    UniversityID            string          `json:"university_id"`
    UniversityName          string          `json:"university_name"`
    Tier                    string          `json:"tier"`
    AdmissionLikelihoodScore int            `json:"admission_likelihood_score"`
    StudentFitScore         int             `json:"student_fit_score"`
    Reason                  string          `json:"reason"`
    Risks                   json.RawMessage `json:"risks"`
    Improvements            json.RawMessage `json:"improvements"`
    RankOrder               int             `json:"rank_order"`
    CreatedAt               time.Time       `json:"created_at"`
}
```

---

### B-3: AI Service Client (30m)

**`internal/client/ai_client.go`:**

```go
type AIClient struct {
    BaseURL    string
    httpClient *http.Client
}

func NewAIClient(baseURL string) *AIClient {
    return &AIClient{
        BaseURL: baseURL,
        httpClient: &http.Client{Timeout: 10 * time.Second},
        // Timeout ngắn vì chỉ cần nhận "accepted: true"
        // AI xử lý async, BE không chờ
    }
}

// SubmitCrawlJob: gửi job crawl, nhận accepted ngay
func (c *AIClient) SubmitCrawlJob(jobID, universityID, universityName, callbackURL string) error {
    payload := map[string]string{
        "job_id":          jobID,
        "university_id":   universityID,
        "university_name": universityName,
        "callback_url":    callbackURL,
    }
    return c.post("/jobs/crawl", payload)
}

// SubmitAnalyzeJob: gửi job analyze profile
func (c *AIClient) SubmitAnalyzeJob(jobID, caseID, callbackURL string, input AnalyzeInput) error {
    payload := map[string]interface{}{
        "job_id":       jobID,
        "case_id":      caseID,
        "callback_url": callbackURL,
        "input":        input,
    }
    return c.post("/jobs/analyze", payload)
}

// SubmitReportJob: gửi job generate report
func (c *AIClient) SubmitReportJob(jobID, caseID, callbackURL string, recs []Recommendation) error {
    payload := map[string]interface{}{
        "job_id":          jobID,
        "case_id":         caseID,
        "callback_url":    callbackURL,
        "recommendations": recs,
    }
    return c.post("/jobs/report", payload)
}

func (c *AIClient) post(path string, payload interface{}) error {
    body, _ := json.Marshal(payload)
    resp, err := c.httpClient.Post(c.BaseURL+path, "application/json", bytes.NewReader(body))
    if err != nil { return err }
    defer resp.Body.Close()
    if resp.StatusCode != 200 { return fmt.Errorf("AI service returned %d", resp.StatusCode) }
    return nil
}
```

---

### B-4: Cases API (90m)

**`POST /api/v1/cases`**

Handler flow:
1. Parse + validate body: `full_name` required, phải có ielts hoặc sat
2. `gpa_normalized` phải có từ FE (FE tự tính) — validate >= 0 && <= 4.0
3. Begin DB transaction
4. Insert `students` → lấy `student_id`
5. Generate `job_id = uuid.New()`
6. Insert `cases` với `status = 'pending'`, `ai_job_id = job_id`
7. Insert `activity_log`: event_type = "case_created"
8. Commit transaction
9. Build `AnalyzeInput` từ student data
10. Gọi `aiClient.SubmitAnalyzeJob(job_id, case_id, callbackURL, input)`
11. Update case `status = 'processing'`, `processing_started_at = NOW()`
12. Insert `activity_log`: event_type = "processing_started"
13. Nếu AI Service không nhận được (error): update status = 'failed', ghi log
14. Response: `{ "case_id": "...", "status": "processing" }`

`callbackURL = config.PublicBaseURL + "/internal/jobs/done"`

---

**`GET /api/v1/cases`**

Query params: `status`, `page` (default 1), `limit` (default 20)

```sql
SELECT c.*, s.full_name, s.gpa_normalized, s.ielts_overall, s.sat_total,
       s.intended_major, s.budget_usd_per_year, s.preferred_countries,
       COUNT(r.id) FILTER (WHERE r.tier = 'safe') as safe_count,
       COUNT(r.id) FILTER (WHERE r.tier = 'match') as match_count,
       COUNT(r.id) FILTER (WHERE r.tier = 'reach') as reach_count
FROM cases c
JOIN students s ON c.student_id = s.id
LEFT JOIN recommendations r ON r.case_id = c.id
WHERE ($1 = '' OR c.status = $1)
GROUP BY c.id, s.id
ORDER BY c.created_at DESC
LIMIT $2 OFFSET $3
```

---

**`GET /api/v1/cases/:id`**

```sql
-- Case + student
SELECT c.*, s.* FROM cases c JOIN students s ON c.student_id = s.id WHERE c.id = $1

-- Recommendations
SELECT r.* FROM recommendations r WHERE r.case_id = $1 ORDER BY r.rank_order ASC
```

Response: case object với `student` embedded, `recommendations` array, `profile_summary` parsed.

---

**`GET /api/v1/cases/:id/crawl-status`**

Logic:
1. Lấy list university_id từ recommendations của case này (nếu đã có)
2. Nếu chưa có recommendations (đang processing): lấy university_id từ universities table filter theo `preferred_countries` của student
3. Query universities với crawl_status + last_crawled_at
4. Response: `[{ university_id, name, crawl_status, changes_detected_text, last_crawled_at }]`

`changes_detected_text` lấy từ activity_log gần nhất của university đó.

---

**`GET /api/v1/cases/count`**

Query: `?status=human_review`
Response: `{ "count": 3 }`

---

**`POST /api/v1/cases/:id/report`**

1. Lấy case + student + recommendations
2. Nếu `report_data` đã có → trả ngay `{ report_url: "/api/v1/cases/:id/report/download" }`
3. Generate `job_id`, gọi `aiClient.SubmitReportJob(...)`
4. Response: `{ "status": "generating" }` — FE poll `GET /api/v1/cases/:id` và check `report_generated_at != null`

---

**`GET /api/v1/cases/:id/report/download`**

1. Lấy `report_data` từ DB (JSONB)
2. Render thành PDF đơn giản dùng `go-fpdf`:
   - Trang 1: tên học sinh + profile summary
   - Trang 2: top 10 recommendations với lý do
   - Trang 3: action items + timeline
3. Response: `Content-Type: application/pdf`, `Content-Disposition: attachment; filename="report_<name>.pdf"`

---

### B-5: Callback Endpoint (45m)

**`POST /internal/jobs/done`**

Đây là endpoint quan trọng nhất. AI Service gọi vào đây sau khi xong job.

```go
func HandleJobDone(c *gin.Context) {
    var payload JobDonePayload
    if err := c.ShouldBindJSON(&payload); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    switch payload.JobType {
    case "crawl_university":
        handleCrawlDone(payload)
    case "analyze_profile":
        handleAnalyzeDone(payload)
    case "generate_report":
        handleReportDone(payload)
    }

    c.JSON(200, gin.H{"received": true})
}

func handleCrawlDone(p JobDonePayload) {
    if p.Status == "failed" {
        // Update university crawl_status = "failed"
        // Insert activity_log
        return
    }
    r := p.Result // CrawlResult
    // Update university fields từ result
    // UPDATE universities SET
    //   ielts_min = r.IeltsMin,
    //   tuition_usd_per_year = r.TuitionUsdPerYear,
    //   scholarship_available = r.ScholarshipAvailable,
    //   scholarship_notes = r.ScholarshipNotes,
    //   application_deadline = r.ApplicationDeadline,
    //   available_majors = r.AvailableMajors,
    //   qs_rank = r.QsRank,
    //   acceptance_rate = r.AcceptanceRate,
    //   crawl_status = r.CrawlStatus,
    //   last_crawled_at = NOW(),
    //   crawl_job_id = NULL
    // WHERE id = p.UniversityID

    if len(r.ChangesDetected) > 0 {
        // Insert activity_log với event_type = "crawl_change"
        // metadata = { changes: r.ChangesDetected, source_url: r.SourceURL }
    }
}

func handleAnalyzeDone(p JobDonePayload) {
    if p.Status == "failed" {
        // Update case status = "human_review", escalation_reason = "AI service failed"
        return
    }
    r := p.Result // AnalyzeResult

    // Begin transaction
    // 1. Insert recommendations (bulk insert)
    // 2. Update case:
    //    status = "human_review" nếu r.EscalationNeeded, else "done"
    //    ai_confidence = r.ConfidenceScore
    //    escalation_reason = r.EscalationReason
    //    profile_summary = r.ProfileSummary (JSONB)
    //    processing_finished_at = NOW()
    // 3. Insert activity_log:
    //    nếu escalated: event_type = "escalated", description = reason
    //    nếu done: event_type = "auto_approved"
    // Commit
}

func handleReportDone(p JobDonePayload) {
    if p.Status == "failed" { return }
    r := p.Result // ReportResult
    // UPDATE cases SET report_data = r (JSONB), report_generated_at = NOW()
    //   WHERE id = p.CaseID
}
```

---

### B-6: Universities API (45m)

**`GET /api/v1/universities`**
- Params: `page`, `limit`, `country`, `search` (ILIKE name)
- Response: `{ data: [...], total: N }`

**`POST /api/v1/universities`**
- Insert university, crawl_status = 'never_crawled'
- Auto-trigger crawl: gọi `aiClient.SubmitCrawlJob(...)` ngay
- Response: university object với status = 'pending'

**`POST /api/v1/universities/crawl-all`**
- Lấy tất cả universities có `crawl_status != 'pending'`
- Filter thêm: `last_crawled_at IS NULL OR last_crawled_at < NOW() - INTERVAL '1 day'`
- Với mỗi university: generate job_id, update `crawl_status = 'pending'`, `crawl_job_id = job_id`, gọi AI Service
- Response: `{ "message": "Crawl started", "count": N }`

**`GET /api/v1/universities/crawl-active`**
- `SELECT COUNT(*) FROM universities WHERE crawl_status = 'pending'`
- Response: `{ "count": N }`

---

### B-7: Dashboard APIs (45m)

**`GET /api/v1/dashboard/stats`**
```sql
-- cases_today
SELECT COUNT(*) FROM cases WHERE created_at::date = CURRENT_DATE

-- avg_processing_minutes
SELECT AVG(EXTRACT(EPOCH FROM (processing_finished_at - processing_started_at))/60)
FROM cases WHERE processing_finished_at IS NOT NULL AND created_at > NOW() - INTERVAL '7 days'

-- awaiting_review
SELECT COUNT(*) FROM cases WHERE status = 'human_review'

-- ai_confidence_avg
SELECT AVG(ai_confidence) FROM cases WHERE created_at > NOW() - INTERVAL '7 days' AND ai_confidence IS NOT NULL
```

**`GET /api/v1/dashboard/cases-by-day`**
```sql
SELECT DATE(created_at) as day, COUNT(*) as count
FROM cases
WHERE created_at > NOW() - INTERVAL '7 days'
GROUP BY DATE(created_at)
ORDER BY day ASC
```

**`GET /api/v1/dashboard/escalation-trend`**
```sql
SELECT DATE_TRUNC('week', created_at) as week,
       COUNT(*) FILTER (WHERE status = 'human_review') as escalated,
       COUNT(*) as total
FROM cases
WHERE created_at > NOW() - INTERVAL '6 weeks'
GROUP BY week ORDER BY week ASC
```

**`GET /api/v1/dashboard/analytics`**
```sql
-- Tổng hợp nhiều query, return JSON object lớn
-- cases_this_month, avg_time, auto_approved_pct, hours_saved (estimated)
-- cases_per_week (6 tuần), tier_distribution, confidence_distribution
-- tinyfish_changes: COUNT từ activity_log WHERE event_type = 'crawl_change'
```

**`GET /api/v1/activity-log`**
- Params: `limit` (default 10)
- JOIN với cases + students để lấy student_name
- ORDER BY created_at DESC

---

### B-8: Server Bootstrap (20m)

**`cmd/server/main.go`:**
1. Load config từ env
2. Connect PostgreSQL, run migrations tự động
3. Init AIClient với `AI_SERVICE_URL`
4. Setup Gin router với CORS middleware
5. Register tất cả routes
6. `router.Run(":" + config.Port)`

**`config/config.go`:**
```go
type Config struct {
    Port          string // default "8080"
    DatabaseURL   string // postgres://...
    AIServiceURL  string // http://localhost:9000
    PublicBaseURL string // http://localhost:8080 (dùng cho callback URL)
}
```

---

## Person C — AI Service (Python + FastAPI + Claude Agent SDK + TinyFish MCP + Neo4j)

### C-0: Project scaffold (30m — làm đầu tiên)

```bash
pip install fastapi uvicorn httpx python-dotenv \
            psycopg2-binary sqlalchemy \
            neo4j \
            claude-agent-sdk \
            anyio \
            anthropic
# claude-agent-sdk: query() + ClaudeAgentOptions + MCP support (crawl worker)
# neo4j: raw async driver cho analyze worker + delete orphans
# anthropic: trực tiếp cho analyze + report worker (1 lần call, không cần agent loop)
# KHÔNG cần neo4j-graphrag, KHÔNG cần openai — graph ghi qua MCP tool write_neo4j_cypher
```

> **mcp-neo4j-cypher** chạy như một process riêng (stdio hoặc HTTP), không phải Python package import. Cài bằng `uvx` hoặc `pip install mcp-neo4j-cypher`. Claude Agent SDK connect tới nó qua `mcp_servers` config.

> **Crawl worker** dùng `claude-agent-sdk` với **2 MCP servers**: TinyFish (web automation) + mcp-neo4j-cypher (write graph). Agent tự quyết crawl bao nhiêu trang, tự viết Cypher để MERGE entities vào graph.

> **Analyze worker** dùng `anthropic` SDK trực tiếp + raw `neo4j` driver để query — không cần agent loop, 1 lần call.

Cấu trúc:
```
main.py
config.py
models.py            # Pydantic models
graph.py             # raw neo4j async driver — chỉ dùng cho analyze query + delete orphans
                     # (crawl worker ghi graph qua mcp-neo4j-cypher MCP, không qua file này)
job_db.py            # SQLAlchemy — chỉ lưu job state
workers/
  crawl_worker.py    # claude-agent-sdk: TinyFish MCP (crawl) + mcp-neo4j-cypher MCP (write graph)
  analyze_worker.py  # raw neo4j driver query → hard filter → anthropic Claude → recommendations
  report_worker.py   # Claude report generation
  callback.py        # httpx callback helper
queue/
  job_queue.py       # asyncio.Queue
prompts/
  crawl_system.txt   # system prompt cho crawl agent
  analyze.txt        # matching prompt
  report.txt
```

**.env:**
```
ANTHROPIC_API_KEY=sk-ant-...    # claude-agent-sdk (crawl) + anthropic SDK (analyze/report)
TINYFISH_API_KEY=...            # TinyFish REST API fallback (nếu OAuth không work)
NEO4J_URI=bolt://localhost:7687
NEO4J_USER=neo4j
NEO4J_PASSWORD=password
JOB_DATABASE_URL=postgres://postgres:password@localhost:5433/unimatch_jobs
PORT=9000
# Không cần OPENAI_API_KEY — mcp-neo4j-cypher không dùng embeddings
```

---

### C-1: Neo4j Schema + `graph.py` — raw driver (30m)

**File:** `graph.py`

**Thay đổi v3:** `graph.py` giờ chỉ là thin wrapper dùng raw `neo4j.AsyncGraphDatabase.driver` cho:
1. **Init constraints** (chạy 1 lần khi startup)
2. **`get_all_universities_flat()`** — analyze worker query Cypher để lấy candidate list
3. **`delete_university_and_orphans()`** — DELETE endpoint
4. **`get_university_flat()`** — report worker enrich

**Ghi graph (crawl)** không còn qua `graph.py` nữa — crawl agent tự gọi `write_neo4j_cypher` MCP tool.

**Node labels và relationships** (agent tự viết Cypher MERGE vào Neo4j):

```
(:University {be_id, name, country, qs_rank, ielts_min, sat_required,
              gpa_expectation_normalized, tuition_usd_per_year,
              acceptance_rate, application_deadline})
(:Program    {name, degree})
(:AdmissionReq {type, min_score, notes})   -- type: IELTS | GPA | SAT | TOEFL
(:Scholarship  {name, amount, currency, notes})
(:City         {name, country})
(:Deadline     {intake, date})

(University)-[:HAS_PROGRAM]  ->(Program)
(University)-[:REQUIRES]     ->(AdmissionReq)
(University)-[:OFFERS]       ->(Scholarship)
(University)-[:LOCATED_IN]   ->(City)
(University)-[:HAS_DEADLINE] ->(Deadline)
(Program)   -[:REQUIRES]     ->(AdmissionReq)
```

**`graph.py`:**

```python
from neo4j import AsyncGraphDatabase
from config import config
import logging

neo4j_driver = AsyncGraphDatabase.driver(
    config.NEO4J_URI,
    auth=(config.NEO4J_USER, config.NEO4J_PASSWORD),
)


async def init_graph_constraints():
    """Chạy 1 lần khi startup."""
    async with neo4j_driver.session() as s:
        await s.run("CREATE CONSTRAINT IF NOT EXISTS FOR (u:University) REQUIRE u.be_id IS UNIQUE")
        await s.run("CREATE CONSTRAINT IF NOT EXISTS FOR (p:Program) REQUIRE p.name IS UNIQUE")
        await s.run("CREATE CONSTRAINT IF NOT EXISTS FOR (c:City) REQUIRE c.name IS UNIQUE")
    logging.info("Neo4j constraints ready")


async def get_all_universities_flat() -> list[dict]:
    """Analyze worker: lấy toàn bộ universities kèm requirements."""
    async with neo4j_driver.session() as s:
        result = await s.run("""
            MATCH (u:University) WHERE u.be_id IS NOT NULL
            OPTIONAL MATCH (u)-[:REQUIRES]->(r:AdmissionReq)
            OPTIONAL MATCH (u)-[:HAS_PROGRAM]->(p:Program)
            OPTIONAL MATCH (u)-[:OFFERS]->(sc:Scholarship)
            RETURN u,
                   collect(DISTINCT r) AS reqs,
                   collect(DISTINCT p.name) AS majors,
                   collect(DISTINCT sc) AS scholarships
        """)
        records = await result.data()
        return [_flatten(r) for r in records]


async def get_university_flat(university_id: str) -> dict | None:
    """Report worker: lấy 1 university theo be_id."""
    async with neo4j_driver.session() as s:
        result = await s.run("""
            MATCH (u:University {be_id: $id})
            OPTIONAL MATCH (u)-[:REQUIRES]->(r:AdmissionReq)
            OPTIONAL MATCH (u)-[:HAS_PROGRAM]->(p:Program)
            OPTIONAL MATCH (u)-[:OFFERS]->(sc:Scholarship)
            OPTIONAL MATCH (u)-[:HAS_DEADLINE]->(d:Deadline)
            RETURN u,
                   collect(DISTINCT r) AS reqs,
                   collect(DISTINCT p.name) AS majors,
                   collect(DISTINCT sc) AS scholarships,
                   collect(DISTINCT d) AS deadlines
        """, id=university_id)
        record = await result.single()
        return _flatten(record) if record else None


async def delete_university_and_orphans(university_id: str):
    """DELETE endpoint: xóa University node + orphaned connected nodes."""
    async with neo4j_driver.session() as s:
        await s.run("""
            MATCH (u:University {be_id: $id})-[r]->(n)
            WHERE size([(n)<--(:University) | n]) <= 1
            DETACH DELETE n
        """, id=university_id)
        await s.run(
            "MATCH (u:University {be_id: $id}) DETACH DELETE u",
            id=university_id
        )


def _flatten(record: dict) -> dict:
    u = dict(record["u"])
    reqs = [dict(r) for r in record.get("reqs", [])]
    scholarships = record.get("scholarships", [])
    ielts = next((r for r in reqs if r.get("type") == "IELTS"), None)
    gpa   = next((r for r in reqs if r.get("type") == "GPA"), None)
    return {
        "university_id":              u.get("be_id"),
        "name":                       u.get("name"),
        "country":                    u.get("country"),
        "qs_rank":                    u.get("qs_rank"),
        "ielts_min":                  ielts.get("min_score") if ielts else u.get("ielts_min"),
        "sat_required":               u.get("sat_required"),
        "gpa_expectation_normalized": gpa.get("min_score") if gpa else u.get("gpa_expectation_normalized"),
        "tuition_usd_per_year":       u.get("tuition_usd_per_year"),
        "scholarship_available":      len(scholarships) > 0,
        "scholarship_notes":          dict(scholarships[0]).get("notes") if scholarships else None,
        "application_deadline":       u.get("application_deadline"),
        "available_majors":           record.get("majors", []),
        "acceptance_rate":            u.get("acceptance_rate"),
    }
```

---

### C-2: Job DB (PostgreSQL — chỉ lưu job state) (20m)

**File:** `job_db.py`

```python
from sqlalchemy import create_engine, Column, String, JSON, DateTime
from sqlalchemy.orm import declarative_base, sessionmaker
from datetime import datetime

Base = declarative_base()

class JobRecord(Base):
    __tablename__ = "jobs"
    id = Column(String, primary_key=True)      # job_id từ BE
    job_type = Column(String)
    status = Column(String, default="pending") # pending|processing|done|failed
    callback_url = Column(String)
    payload = Column(JSON)
    result = Column(JSON, nullable=True)
    error = Column(String, nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

engine = create_engine(config.JOB_DATABASE_URL)
Base.metadata.create_all(engine)
Session = sessionmaker(bind=engine)

def create_job(job_id, job_type, callback_url, payload):
    with Session() as s:
        s.add(JobRecord(id=job_id, job_type=job_type,
                        callback_url=callback_url, payload=payload))
        s.commit()

def update_job(job_id, status, result=None, error=None):
    with Session() as s:
        job = s.get(JobRecord, job_id)
        if job:
            job.status = status
            job.result = result
            job.error = error
            job.updated_at = datetime.utcnow()
            s.commit()
```

---

### C-3: Job Queue (15m)

**File:** `queue/job_queue.py`

```python
import asyncio
from collections import defaultdict

_queues: dict[str, asyncio.Queue] = defaultdict(asyncio.Queue)

async def enqueue(job_type: str, data: dict):
    await _queues[job_type].put(data)

async def start_all_workers():
    from workers.crawl_worker import crawl_worker_loop
    from workers.analyze_worker import analyze_worker_loop
    from workers.report_worker import report_worker_loop
    await asyncio.gather(
        crawl_worker_loop(),
        analyze_worker_loop(),
        report_worker_loop(),
    )

def make_worker_loop(job_type: str, process_fn):
    """Generic worker loop factory."""
    async def loop():
        while True:
            job = await _queues[job_type].get()
            try:
                await process_fn(job)
            except Exception as e:
                logging.error(f"[{job_type}] job {job.get('job_id')} failed: {e}")
                from workers.callback import callback_be
                await callback_be(job["callback_url"], job["job_id"],
                                  job_type, "failed", error=str(e),
                                  university_id=job.get("university_id"),
                                  case_id=job.get("case_id"))
    return loop
```

---

### C-4: Crawl Worker — claude-agent-sdk + TinyFish + mcp-neo4j-cypher (90m)

**File:** `workers/crawl_worker.py`

**Kiến trúc:** Agent dùng **2 MCP servers đồng thời**:
- **TinyFish** (`run_web_automation`) — browse web, extract data từ university websites
- **mcp-neo4j-cypher** (`write_neo4j_cypher`, `get_neo4j_schema`) — ghi entities/relationships trực tiếp vào Neo4j

Agent tự quyết crawl bao nhiêu trang, tự viết Cypher MERGE — không cần `graph.py` cho bước ghi.

**mcp-neo4j-cypher setup (chạy trước khi start AI Service):**
```bash
# Cài
pip install mcp-neo4j-cypher

# Chạy ở HTTP mode (để FastAPI process connect được)
NEO4J_URI=bolt://localhost:7687 \
NEO4J_USERNAME=neo4j \
NEO4J_PASSWORD=password \
NEO4J_DATABASE=neo4j \
mcp-neo4j-cypher --transport http \
  --server-host 127.0.0.1 \
  --server-port 8081 \
  --server-path /api/mcp/
```

Hoặc thêm vào `docker-compose.yml` bên cạnh Neo4j container.

**System prompt** (`prompts/crawl_system.txt`):

```
You are a university research agent with access to two MCP servers:
1. TinyFish Web Agent — use run_web_automation to browse real websites
2. Neo4j Cypher — use get_neo4j_schema to check schema, write_neo4j_cypher to store data

RESEARCH STRATEGY:
Step 1. Use run_web_automation on google.com: goal = "find official admissions
        page URL for {university_name} {country}"
Step 2. Use run_web_automation on that URL: goal = "extract IELTS requirements,
        tuition fees, available programs, scholarships, deadlines, acceptance rate"
Step 3. Make additional run_web_automation calls for specific pages if needed.
        Use browser_profile="stealth" if a site blocks normal access.

FIELDS TO EXTRACT (null if not findable):
- ielts_min (float), sat_required (bool)
- gpa_expectation_normalized (float, 0.0-4.0 scale)
- tuition_usd_per_year (int, convert: GBP×1.27 EUR×1.08 AUD×0.65 SGD×0.74)
- scholarship_available (bool), scholarship_notes (str)
- application_deadline (YYYY-MM-DD, nearest upcoming)
- available_majors (array of strings), acceptance_rate (float 0-1), qs_rank (int)

KNOWLEDGE GRAPH WRITING:
After extracting data, use write_neo4j_cypher to store entities in Neo4j.
The University node already exists with be_id={university_id}. Update it and
create connected nodes using MERGE patterns. Example Cypher:

  // Update University node
  MERGE (u:University {be_id: "{university_id}"})
  SET u.qs_rank = 57, u.tuition_usd_per_year = 19800,
      u.acceptance_rate = 0.32, u.application_deadline = "2026-04-15"

  // Add Program nodes
  MERGE (p:Program {name: "Computer Science"})
  SET p.degree = "MSc"
  MERGE (u)-[:HAS_PROGRAM]->(p)

  // Add AdmissionReq nodes
  MERGE (r:AdmissionReq {type: "IELTS", university_be_id: "{university_id}"})
  SET r.min_score = 7.0, r.notes = "All bands min 6.0"
  MERGE (u)-[:REQUIRES]->(r)

  // Add Scholarship
  MERGE (sc:Scholarship {name: "Holland Scholarship"})
  SET sc.amount = 5000, sc.currency = "EUR", sc.notes = "Merit-based"
  MERGE (u)-[:OFFERS]->(sc)

  // Add City
  MERGE (c:City {name: "Delft"})
  SET c.country = "Netherlands"
  MERGE (u)-[:LOCATED_IN]->(c)

You may split into multiple write_neo4j_cypher calls (1 per entity type is fine).
Use get_neo4j_schema first if unsure about existing node structure.

FINAL RESPONSE — when all done, respond with JSON:
{
  "fixed_fields": {
    "ielts_min": 7.0,
    "sat_required": false,
    "gpa_expectation_normalized": 3.2,
    "tuition_usd_per_year": 19800,
    "scholarship_available": true,
    "scholarship_notes": "Holland Scholarship EUR 5000",
    "application_deadline": "2026-04-15",
    "available_majors": ["CS", "Aerospace"],
    "acceptance_rate": 0.32,
    "qs_rank": 57
  },
  "source_urls": ["https://www.tudelft.nl/..."]
}
Null for any field not found. Graph has already been written — no need to
include graph_entities in the response.
```

**Worker implementation:**

```python
import json
from datetime import datetime
from claude_agent_sdk import query, ClaudeAgentOptions, ResultMessage, AssistantMessage
from job_db import update_job
from workers.callback import callback_be
from config import config
from queue.job_queue import make_worker_loop

TINYFISH_MCP_URL   = "https://agent.tinyfish.ai/mcp"
NEO4J_MCP_URL      = "http://127.0.0.1:8081/api/mcp/"  # mcp-neo4j-cypher local HTTP


async def process_crawl_job(job: dict):
    university_id = job["university_id"]
    metadata      = job["metadata"]
    job_id        = job["job_id"]
    callback_url  = job["callback_url"]

    update_job(job_id, "processing")

    known_fields = {k: v for k, v in metadata.items() if v is not None}
    null_fields  = [k for k, v in metadata.items() if v is None]

    with open("prompts/crawl_system.txt") as f:
        system_prompt = (f.read()
                         .replace("{university_name}", metadata["name"])
                         .replace("{university_id}",   university_id)
                         .replace("{country}",         metadata["country"]))

    prompt = f"""Research this university and update the knowledge graph.

University: {metadata['name']}
Country: {metadata['country']}
BE ID: {university_id}

Already known: {json.dumps(known_fields, indent=2)}
Fields still null (need to find): {null_fields}

Start by finding the official admission page, then extract all missing fields
and write them to Neo4j using write_neo4j_cypher.
"""

    options = ClaudeAgentOptions(
        system_prompt=system_prompt,
        mcp_servers={
            "tinyfish": {
                "type": "http",
                "url": TINYFISH_MCP_URL,
                # Auth via OAuth 2.1 (setup: claude mcp add --transport http tinyfish ...)
            },
            "neo4j": {
                "type": "http",
                "url": NEO4J_MCP_URL,
                # mcp-neo4j-cypher local, no auth needed
            },
        },
        max_turns=25,  # crawl + multiple write_neo4j_cypher calls
    )

    final_text  = None
    source_urls = []

    async for message in query(prompt=prompt, options=options):
        if isinstance(message, ResultMessage):
            final_text = message.result
        elif isinstance(message, AssistantMessage):
            for block in message.content:
                inp = getattr(block, "input", {}) or {}
                if isinstance(inp, dict) and "url" in inp:
                    u = inp["url"]
                    if u and u not in source_urls:
                        source_urls.append(u)

    if not final_text:
        raise ValueError("Agent returned no final text")

    # Parse JSON final response
    clean        = final_text.strip().lstrip("```json").lstrip("```").rstrip("```").strip()
    agent_output = json.loads(clean)
    fixed_fields = agent_output.get("fixed_fields", {})
    # source_urls may also come from agent JSON
    source_urls  = agent_output.get("source_urls", source_urls)

    # Detect changes vs previously known values
    changes = [
        f"{f}: {metadata[f]} → {v}"
        for f, v in fixed_fields.items()
        if metadata.get(f) is not None and v is not None and metadata[f] != v
    ]

    result = {
        **metadata,
        **{k: v for k, v in fixed_fields.items() if v is not None},
        "crawl_status":     "changed" if changes else "ok",
        "changes_detected": changes,
        "source_urls":      source_urls,
        "crawled_at":       datetime.utcnow().isoformat() + "Z",
    }
    await callback_be(callback_url, job_id, "crawl_university", "done",
                      result=result, university_id=university_id)
    update_job(job_id, "done", result=result)


crawl_worker_loop = make_worker_loop("crawl_university", process_crawl_job)
```

> **University node pre-created:** Trước khi enqueue crawl job, BE gọi `POST /jobs/crawl` — AI Service cần MERGE University node với `be_id` ngay khi nhận job (trong `main.py` route handler hoặc đầu `process_crawl_job`), trước khi agent chạy. Agent chỉ cần `SET` properties, không cần CREATE.

> **mcp-neo4j-cypher port 8081:** Tránh conflict với Neo4j Browser (7474) và bolt (7687). Thêm vào docker-compose hoặc supervisor để auto-restart.

> **OAuth TinyFish:** Setup 1 lần: `claude mcp add --transport http tinyfish https://agent.tinyfish.ai/mcp`. Nếu Docker/server không có browser, dùng TinyFish REST API fallback (`X-API-Key` header).

### C-5: Analyze Worker (60m)

**File:** `workers/analyze_worker.py`

Analyze worker **không dùng agent loop** — query Neo4j bằng raw driver qua `graph.py`, hard filter, rồi gọi Claude 1 lần.

**Thay đổi v3:** Thay `VectorCypherRetriever` bằng `get_all_universities_flat()` (Cypher scan thuần) — đơn giản hơn, không cần embeddings, đủ dùng cho hackathon với ~5-20 universities.

```python
import asyncio
import json
import anthropic   # anthropic SDK trực tiếp — 1 lần call, không cần agent loop
from graph import get_all_universities_flat
from job_db import update_job
from workers.callback import callback_be
from config import config
from queue.job_queue import make_worker_loop

claude = anthropic.Anthropic(api_key=config.ANTHROPIC_API_KEY)
SYSTEM = "You are an expert admissions counselor. Respond with valid JSON only."


def _hard_filter(unis: list[dict], input_data: dict) -> list[dict]:
    """Pure Python hard filter — loại candidates không pass threshold cứng."""
    filtered = []
    for u in unis:
        # Country filter
        if input_data.get("preferred_countries"):
            if u.get("country") not in input_data["preferred_countries"]:
                continue
        # Budget (40% buffer cho scholarship cases)
        if not input_data.get("scholarship_required") and u.get("tuition_usd_per_year"):
            if u["tuition_usd_per_year"] > input_data["budget_usd_per_year"] * 1.4:
                continue
        # Major (loose 4-char prefix match)
        if input_data.get("intended_major") and u.get("available_majors"):
            m = input_data["intended_major"].lower()
            if not any(m[:4] in major.lower() for major in u["available_majors"]):
                continue
        # IELTS hard cutoff (1.5 band tolerance)
        if input_data.get("ielts_overall") and u.get("ielts_min"):
            if input_data["ielts_overall"] < u["ielts_min"] - 1.5:
                continue
        filtered.append(u)

    filtered.sort(key=lambda x: x.get("qs_rank") or 999)
    return filtered[:12]


async def process_analyze_job(job: dict):
    case_id    = job["case_id"]
    input_data = job["input"]
    job_id     = job["job_id"]

    update_job(job_id, "processing")

    # 1. Query graph
    all_unis = await get_all_universities_flat()

    # 2. Hard filter
    filtered = _hard_filter(all_unis, input_data)

    if not filtered:
        await callback_be(job["callback_url"], job_id, "analyze_profile", "done",
                          case_id=case_id, result={
                              "profile_summary": {"weaknesses": ["No matching universities in KB"]},
                              "recommendations": [],
                              "confidence_score": 0.3,
                              "escalation_needed": True,
                              "escalation_reason": "No universities in knowledge graph match the given criteria",
                          })
        update_job(job_id, "done")
        return

    # 3. Claude matching (1 call)
    with open("prompts/analyze.txt") as f:
        prompt_template = f.read()

    prompt = prompt_template.format(
        student_json=json.dumps(input_data, ensure_ascii=False, indent=2),
        universities_json=json.dumps(filtered, ensure_ascii=False, indent=2),
    )

    msg = claude.messages.create(
        model="claude-sonnet-4-20250514",
        max_tokens=4000,
        system=SYSTEM,
        messages=[{"role": "user", "content": prompt}],
    )

    result = json.loads(msg.content[0].text)
    await callback_be(job["callback_url"], job_id, "analyze_profile", "done",
                      case_id=case_id, result=result)
    update_job(job_id, "done", result=result)


analyze_worker_loop = make_worker_loop("analyze_profile", process_analyze_job)
```

### C-6: Report Worker (40m)

**File:** `workers/report_worker.py`

```python
async def process_report_job(job: dict):
    case_id = job["case_id"]
    recommendations = job["recommendations"]
    job_id = job["job_id"]

    update_job(job_id, "processing")

    # Enrich recommendations từ Neo4j
    enriched = []
    for rec in recommendations:
        uni_data = await get_university_flat(rec["university_id"])
        enriched.append({**rec, **(uni_data or {})})

    with open("prompts/report.txt") as f:
        prompt_template = f.read()

    msg = claude.messages.create(
        model="claude-sonnet-4-20250514",
        max_tokens=2000,
        system="You are an admissions counselor. Respond JSON only.",
        messages=[{"role": "user", "content": prompt_template.format(
            student_name=job.get("student_name", ""),
            recs_json=json.dumps(enriched, ensure_ascii=False, indent=2)
        )}]
    )

    result = json.loads(msg.content[0].text)
    await callback_be(job["callback_url"], job_id, "generate_report", "done",
                      case_id=case_id, result=result)
    update_job(job_id, "done", result=result)


report_worker_loop = make_worker_loop("generate_report", process_report_job)
```

---

### C-7: Callback Helper (15m)

**File:** `workers/callback.py`

```python
import httpx, asyncio, logging

async def callback_be(callback_url: str, job_id: str, job_type: str,
                      status: str, result: dict = None, error: str = None,
                      case_id: str = None, university_id: str = None):
    payload = {"job_id": job_id, "job_type": job_type,
               "status": status, "error": error, "result": result}
    if case_id: payload["case_id"] = case_id
    if university_id: payload["university_id"] = university_id

    for attempt in range(2):
        try:
            async with httpx.AsyncClient() as c:
                resp = await c.post(callback_url, json=payload, timeout=10)
                resp.raise_for_status()
                return
        except Exception as e:
            logging.error(f"Callback attempt {attempt+1} failed: {e}")
            if attempt == 0: await asyncio.sleep(5)
```

---

### C-8: Pydantic Models (20m)

**File:** `models.py`

```python
from pydantic import BaseModel
from typing import Optional

class UniversityMetadata(BaseModel):
    name: str
    country: str
    qs_rank: Optional[int] = None
    ielts_min: Optional[float] = None
    sat_required: Optional[bool] = None
    gpa_expectation_normalized: Optional[float] = None
    tuition_usd_per_year: Optional[int] = None
    scholarship_available: Optional[bool] = None
    scholarship_notes: Optional[str] = None
    application_deadline: Optional[str] = None   # YYYY-MM-DD
    available_majors: Optional[list[str]] = None
    acceptance_rate: Optional[float] = None

class CrawlJobRequest(BaseModel):
    job_id: str
    university_id: str                           # BE's UUID
    callback_url: str
    metadata: UniversityMetadata                 # full metadata, nulls = fields to detect

class AnalyzeInput(BaseModel):
    full_name: str
    gpa_normalized: float
    ielts_overall: Optional[float] = None
    sat_total: Optional[int] = None
    intended_major: str
    budget_usd_per_year: int
    preferred_countries: list[str] = []
    target_intake: str
    scholarship_required: bool = False
    extracurriculars: str = ""
    achievements: str = ""

class AnalyzeJobRequest(BaseModel):
    job_id: str
    case_id: str
    callback_url: str
    input: AnalyzeInput

class ReportJobRequest(BaseModel):
    job_id: str
    case_id: str
    callback_url: str
    student_name: str
    recommendations: list[dict]
```

---

### C-9: FastAPI Routes (25m)

**File:** `main.py`

```python
from fastapi import FastAPI, HTTPException
from contextlib import asynccontextmanager
import asyncio
from graph import init_graph_constraints, neo4j_driver
from job_db import create_job, update_job, get_job
from queue.job_queue import enqueue, start_all_workers
from models import CrawlJobRequest, AnalyzeJobRequest, ReportJobRequest

@asynccontextmanager
async def lifespan(app: FastAPI):
    await init_graph_constraints()
    asyncio.create_task(start_all_workers())
    yield

app = FastAPI(title="UniMatch AI Service", lifespan=lifespan)

@app.post("/jobs/crawl")
async def submit_crawl(body: CrawlJobRequest):
    # Pre-create University node với be_id TRƯỚC khi agent chạy
    # Agent sẽ MERGE properties vào node này — không cần tạo mới
    async with neo4j_driver.session() as s:
        await s.run("""
            MERGE (u:University {be_id: $id})
            SET u.name = $name, u.country = $country, u.updated_at = datetime()
        """, id=body.university_id,
             name=body.metadata.name,
             country=body.metadata.country)

    create_job(body.job_id, "crawl_university", body.callback_url, body.model_dump())
    await enqueue("crawl_university", body.model_dump())
    return {"accepted": True, "job_id": body.job_id}

@app.post("/jobs/analyze")
async def submit_analyze(body: AnalyzeJobRequest):
    create_job(body.job_id, "analyze_profile", body.callback_url, body.model_dump())
    await enqueue("analyze_profile", body.model_dump())
    return {"accepted": True, "job_id": body.job_id}

@app.post("/jobs/report")
async def submit_report(body: ReportJobRequest):
    create_job(body.job_id, "generate_report", body.callback_url, body.model_dump())
    await enqueue("generate_report", body.model_dump())
    return {"accepted": True, "job_id": body.job_id}

@app.delete("/jobs/university/{university_id}")
async def delete_university_graph(university_id: str):
    """BE gọi khi xóa university — xóa toàn bộ nodes liên quan trong graph."""
    from graph import delete_university_and_orphans
    await delete_university_and_orphans(university_id)
    return {"deleted": True, "university_id": university_id}

@app.get("/jobs/{job_id}")
async def get_job_status(job_id: str):
    job = get_job(job_id)
    if not job: raise HTTPException(404, "Job not found")
    return {"job_id": job_id, "status": job.status, "error": job.error}

@app.get("/graph/university/{university_id}")
async def get_graph_node(university_id: str):
    """Debug: xem graph data của 1 university."""
    from graph import get_university_flat
    data = await get_university_flat(university_id)
    if not data: raise HTTPException(404)
    return data

@app.get("/health")
async def health():
    from graph import neo4j_driver  # async raw driver
    try:
        async with neo4j_driver.session() as s:
            result = await s.run("MATCH (u:University) RETURN count(u) as n")
            record = await result.single()
            uni_count = record["n"]
    except Exception:
        uni_count = -1
    return {"status": "ok", "graph_university_count": uni_count}
```

---

## Integration Checklist (cả 3 người verify khi merge)

### Thứ tự start services

```bash
# 1. Start databases
docker run -d -p 5432:5432 -e POSTGRES_PASSWORD=password -e POSTGRES_DB=unimatch_be postgres:15
docker run -d -p 5433:5433 -e POSTGRES_PASSWORD=password -e POSTGRES_DB=unimatch_jobs postgres:15
docker run -d -p 7474:7474 -p 7687:7687 \
  -e NEO4J_AUTH=neo4j/password \
  neo4j:5

# 2. Start mcp-neo4j-cypher (phải chạy TRƯỚC AI Service)
pip install mcp-neo4j-cypher
NEO4J_URI=bolt://localhost:7687 \
NEO4J_USERNAME=neo4j \
NEO4J_PASSWORD=password \
NEO4J_DATABASE=neo4j \
mcp-neo4j-cypher --transport http \
  --server-host 127.0.0.1 \
  --server-port 8081 \
  --server-path /api/mcp/ &
# Verify: curl http://127.0.0.1:8081/api/mcp/
# Expected: MCP endpoint responds

# 3. Start AI Service
cd ai-service && python main.py
# Verify: curl http://localhost:9000/health
# Expected: { "status": "ok", "graph_university_count": 0 }

# 4. Start BE (tự run migrations + seed)
cd be && go run cmd/server/main.go
# Verify: curl http://localhost:8080/api/v1/universities
# Expected: 5 universities, auto-trigger crawl cho tất cả
# Verify graph sau ~60s: curl http://localhost:9000/health
# Expected: { "status": "ok", "graph_university_count": 5 }

# 5. Start FE
cd fe && npm run dev
# Verify: http://localhost:5173
```

### Test end-to-end

**Flow 1 — Crawl:**
1. BE start → tự gọi `POST /jobs/crawl` cho 5 universities seed
2. AI enqueue → crawl workers xử lý → callback về `POST /internal/jobs/done`
3. BE cập nhật universities với data từ TinyFish + Claude
4. FE → University KB: thấy status "ok" hoặc "changed"

**Flow 2 — Analyze:**
1. FE → Cases → New Case → nhập Nguyen Linh, GPA 8.4/10, IELTS 7.0, CS, UK/NL, $35k
2. FE tự tính `gpa_normalized = 3.6` → `POST /api/v1/cases`
3. BE insert student + case, gọi `POST /jobs/analyze` sang AI
4. AI enqueue → analyze worker lấy KB → hard filter → Claude → callback
5. BE nhận callback → insert recommendations → update case status = "done"
6. FE poll `GET /api/v1/cases/:id` mỗi 3s → thấy status "done" → render recommendations
7. TinyFish crawl panel: `GET /api/v1/cases/:id/crawl-status` hiển thị trạng thái từng trường

**Flow 4 — Delete university:**
1. FE → University KB → xóa "TU Delft"
2. BE `DELETE /api/v1/universities/:id` → gọi `DELETE /jobs/university/:id` sang AI Service
3. AI Service xóa University node + orphaned Program/AdmissionReq/Scholarship/Deadline nodes
4. City node "Delft" được giữ lại nếu có university khác reference (orphan check)
5. `curl http://localhost:9000/health` → `graph_university_count` giảm 1

### Failure scenarios cần test

| Scenario | Expected behavior |
|----------|-------------------|
| AI Service down khi BE gửi job | BE update case/university status = "failed", log error |
| TinyFish MCP không trả kết quả | Agent tự retry tool, cuối cùng callback "partial" với fields tìm được |
| mcp-neo4j-cypher down khi crawl | Agent không ghi được graph → callback "done" với fixed_fields nhưng graph rỗng |
| Claude agent trả invalid JSON | Worker retry 1 lần, sau đó callback "failed" |
| Callback về BE thất bại | AI retry 1 lần sau 5s, log nếu vẫn fail |
| Neo4j down | init_graph_constraints fail → AI Service không start; crawl/analyze callback "failed" |
| mcp-neo4j-cypher không kết nối được Neo4j | write_neo4j_cypher trả error → agent thấy lỗi, báo trong response |
| Graph rỗng khi analyze | Escalate với reason "No universities in knowledge graph" |
| BE xóa university | `DELETE /jobs/university/:id` → Neo4j xóa node + orphans |

---

## Environment Variables

### FE `.env.local`
```
VITE_API_BASE_URL=http://localhost:8080/api/v1
```

### BE `.env`
```
PORT=8080
DATABASE_URL=postgres://postgres:password@localhost:5432/unimatch_be
AI_SERVICE_URL=http://localhost:9000
PUBLIC_BASE_URL=http://localhost:8080
```

### AI Service `.env`
```
PORT=9000
JOB_DATABASE_URL=postgres://postgres:password@localhost:5433/unimatch_jobs
ANTHROPIC_API_KEY=sk-ant-...       # claude-agent-sdk (crawl) + anthropic SDK (analyze/report)
TINYFISH_API_KEY=...               # TinyFish REST API fallback (nếu OAuth không work)
NEO4J_URI=bolt://localhost:7687
NEO4J_USER=neo4j
NEO4J_PASSWORD=password
NEO4J_MCP_URL=http://127.0.0.1:8081/api/mcp/  # mcp-neo4j-cypher local HTTP
# KHÔNG cần OPENAI_API_KEY
```

---

## Task & Time Estimates

| ID | Task | Person | Est. | Priority |
|----|------|--------|------|----------|
| A-0 | FE scaffold + routing | A | 30m | P0 |
| B-0 | BE scaffold + config | B | 30m | P0 |
| C-0 | AI Service scaffold | C | 30m | P0 |
| B-1 | DB schema + migrations + seed | B | 45m | P0 |
| C-1 | Neo4j schema + graph.py (raw driver) | C | 30m | P0 |
| C-2 | Job DB (PostgreSQL) | C | 20m | P0 |
| C-3 | Job queue + worker loop factory | C | 15m | P0 |
| B-2 | Go models | B | 20m | P0 |
| B-3 | AI Client (fire & forget) | B | 30m | P0 |
| C-8 | Pydantic models | C | 20m | P0 |
| A-1 | Layout shell + sidebar | A | 45m | P0 |
| C-4 | Crawl worker: claude-agent-sdk + TinyFish MCP + mcp-neo4j-cypher | C | 90m | P0 |
| C-5 | Analyze worker: raw Cypher query + hard filter + Claude | C | 60m | P0 |
| B-4 | Cases API (POST + GET list + GET :id) | B | 90m | P0 |
| B-5 | Callback endpoint `/internal/jobs/done` | B | 45m | P0 |
| A-2 | Cases list page | A | 90m | P0 |
| A-4 | Case detail page | A | 90m | P0 |
| C-9 | FastAPI routes (incl. DELETE university) | C | 25m | P0 |
| C-7 | Callback helper | C | 15m | P0 |
| B-6 | Universities API (incl. DELETE → call AI) | B | 45m | P1 |
| A-3 | New case form | A | 60m | P1 |
| A-5 | Review queue page | A | 45m | P1 |
| B-7 | Dashboard stats APIs | B | 45m | P1 |
| C-6 | Report worker | C | 40m | P1 |
| A-6 | University KB page | A | 45m | P2 |
| A-7 | Analytics page | A | 45m | P2 |
| A-8 | Composables + toast + polish | A | 30m | P2 |
| ALL | Integration test + bug fix | All | 60m | P0 |

> **P0** — demo không chạy được nếu thiếu
> **P1** — demo đầy đủ tính năng
> **P2** — nice-to-have nếu còn thời gian

### Lưu ý quan trọng khi integrate

- **Start order quan trọng:** Neo4j → mcp-neo4j-cypher (port 8081) → AI Service → BE → FE. `init_graph_constraints()` kết nối Neo4j khi startup, nếu Neo4j chưa ready sẽ fail. mcp-neo4j-cypher phải chạy trước AI Service vì crawl agent gọi nó ngay khi nhận job đầu tiên.
- **mcp-neo4j-cypher port 8081:** Tránh conflict với Neo4j Browser (7474) và bolt (7687). Chạy background với `&` hoặc trong tmux, hoặc thêm vào docker-compose.
- **Crawl phải xong trước analyze.** Graph rỗng → analyze sẽ escalate 100%. BE cần trigger crawl-all ngay sau seed và chờ ít nhất 1–2 universities done trước khi test analyze.
- **TinyFish MCP setup (bắt buộc trước khi chạy):** `claude mcp add --transport http tinyfish https://agent.tinyfish.ai/mcp` — chạy 1 lần, browser mở OAuth, token lưu local. TinyFish dùng HTTP/Streamable HTTP, **không phải SSE**. Nếu chạy trong Docker không có browser, dùng REST API fallback (`X-API-Key` header).
- **Agent loop timeout:** Crawl agent mất 30–120s/university (crawl + nhiều write_neo4j_cypher calls). `B-3 AIClient` set timeout 15s chỉ chờ `accepted`, không phải chờ kết quả.
- **Không cần OPENAI_API_KEY.** mcp-neo4j-cypher không dùng embeddings. Graph được ghi bởi agent viết Cypher trực tiếp.
