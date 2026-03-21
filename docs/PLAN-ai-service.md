# AI Service Scaffold - Implementation Plan (v3)

**Mục tiêu**: Xây dựng backend AI Service (Python/FastAPI) độc lập, chuyên xử lý background jobs từ hệ thống Backend chính (Go). Đảm nhận việc thu thập thông tin các trường đại học thông qua TinyFish MCP & Neo4j MCP và thực hiện luồng phân tích/gợi ý hồ sơ với Claude. **(Bỏ qua BE/FE)**.

## 1. Môi trường & Dependencies (Phase 1)
- **Công cụ**: Python 3.10+, FastAPI, Uvicorn.
- **Libs**: `fastapi`, `uvicorn`, `httpx`, `python-dotenv`, `psycopg2-binary`, `sqlalchemy`, `neo4j`, `claude-agent-sdk`, `anyio`, `anthropic`.
- **Cấu trúc thư mục**: 
  ```text
  ai-service/
  ├── main.py
  ├── config.py
  ├── models.py
  ├── graph.py
  ├── job_db.py
  ├── workers/
  │   ├── crawl_worker.py
  │   ├── analyze_worker.py
  │   ├── report_worker.py
  │   └── callback.py
  ├── queue/
  │   └── job_queue.py
  └── prompts/
      ├── crawl_system.txt
      ├── analyze.txt
      └── report.txt
  ```

### Verification (Bước 1):
Chạy lệnh `uvicorn main:app --port 9000`. Kiểm tra `http://localhost:9000/docs` mở thành công UI của FastAPI Swagger.

---

## 2. Config & Database (Phase 2)

### 2.1 Cấu hình Environment Variables (`config.py`):
Parse các env map sang class `Config`: `ANTHROPIC_API_KEY`, `TINYFISH_API_KEY`, `NEO4J_URI`, `NEO4J_USER`, `NEO4J_PASSWORD`, `JOB_DATABASE_URL`, `PORT`. (Linh hoạt fallback REST API nếu OAuth của TinyFish bị chặn).

### 2.2 PostgreSQL Job State Table (`job_db.py`):
Dùng `SQLAlchemy` khai báo bảng `jobs` lưu trạng thái của Background Task.
- Fields: `id` (PK, job_id từ BE), `job_type`, `status` (pending/processing/done/failed), `callback_url`, `payload` (JSON), `result` (JSON), `error` (String).
- Hàm thực thi: `create_job()` & `update_job()`.

### 2.3 Neo4j Driver Connection (`graph.py`):
Viết driver kết nối với Neo4j thuần (Khởi tạo `AsyncGraphDatabase.driver`).
- Hàm `init_graph_constraints()` để tạo index `be_id` cho University.
- Hàm `get_all_universities_flat()`: query raw cypher đọc data.
- Hàm `delete_university_and_orphans()`: cypher xóa node & orphan.

### Verification (Bước 2):
Viết file test kết nối: 
- Gọi SQL insert 1 row vào table `jobs`. Kết nối DB xem ID có push thành công.
- Gọi Neo4j constraints => Báo success trên terminal.

---

## 3. Queue & Data Models (Phase 3)

### 3.1 Pydantic Models (`models.py`)
Định nghĩa: `UniversityMetadata`, `CrawlJobRequest`, `AnalyzeInput`, `AnalyzeJobRequest`, `ReportJobRequest`.

### 3.2 Asyncio Job Queue (`queue/job_queue.py`)
- Dùng `asyncio.Queue` để bóc tách luồng gọi API sang Async processing (vì AI xử lý rất chậm).
- Factory function: `make_worker_loop("type", process_fn)` để scale dễ dàng cho crawl/analyze.

### Verification (Bước 3):
Mock thử đẩy payload qua Pydantic. Quăng JSON thiếu field để xác nhận Pydantic ném Validation Error.

---

## 4. System Prompts (Phase 4)

Viết cứng các prompt để `claude-agent-sdk` (Crawl) và `anthropic` (Analyze/Report) sử dụng.
- **`crawl_system.txt`**: Prompt chi tiết hướng dẫn Claude Agent tra cứu qua TinyFish và sử dụng công cụ `write_neo4j_cypher` ghi data vảo Knowledge Graph.
- **`analyze.txt`**: Cấu trúc JSON chứa rules so khớp hồ sơ student và university arrays từ Cypher filter.
- **`report.txt`**: Kịch bản sinh Report định dạng JSON từ thông tin Recommend.

### Verification (Bước 4):
Chạy thử template render biến chuỗi không bị lỗi KeyError do vướng ký tự ngoặc nhọn `{}`.

---

## 5. Workers Layer (Phân tích, Scrape - Phase 5)

### 5.1 Callback Helper (`workers/callback.py`)
- Viết hàm `callback_be(callback_url, job_id, status...)` 
- Có cơ chế retry 1 lân nếu BE timeout (do network/lock).

### 5.2 Crawl Worker (`workers/crawl_worker.py`)
- Khởi tạo Agent với `ClaudeAgentOptions`. 
- Gán **2 MCP Server Configs**: Tinyfish (`https://agent.tinyfish.ai/mcp`) & Neo4j MCP local (`http://127.0.0.1:8081/api/mcp/`).
- Thực hiện vòng lặp `query()`. Khi crawl xong agent parse string Json, lưu changes vào `Job DB` -> Gọi `callback_be`. 

### 5.3 Analyze Worker (`workers/analyze_worker.py`)
- Thực hiện raw sql qua hàm `get_all_universities_flat()`.
- Filter cứng `_hard_filter()` với điều kiện IELTS thiếu hụt < 1.5 band, Budget không vượt 1.4 lần...
- Gọi **1 lệnh duy nhất** sang Claude-Sonnet-3.5 qua thư viện `anthropic`.

### 5.4 Report Worker (`workers/report_worker.py`)
Nhận recommendations từ FE/BE, lấy text gen từ Claude -> callback gửi PDF report summary.

### Verification (Bước 5):
Mock URL của BE (webhook test api), gõ file `.py` chạy bằng terminal truyền cứng dữ liệu test xem có log successfully received request ở đích mcp-neo4j-cypher và webhook.

---

## 6. FastAPI Routes Integration (Phase 6)

Ghép Controller với Queue (`main.py`)
- Middleware / Context Lifespan: Khởi chạy Async Loop queue (gọi 3 worker loop song song), init constraints base neo4j.
- Route đăng ký:
  - `POST /jobs/crawl`: Lưu SQL state -> đẩy vào Queue. Ghi Log ngay lập tức.
  - `POST /jobs/analyze`: Lưu State vào SQL -> đẩy array vào Async Queue. 
  - `POST /jobs/report`
  - `DELETE /jobs/university/{university_id}`
  - `GET /health` : Lấy Status, query Neo4j nodes (check Connection status DB).
  - `GET /jobs/{job_id}`: Lấy state.

### Verification (Bước 6):
Dùng file `.http` hoặc Postman gọi tới 3 API `POST`. Trả về `{ accepted: True, job_id: id }` dưới 10ms. Console in log processing.

---

## Swagger API Specification 

Khi `uvicorn` khởi chạy, tự động docs sinh ra ở `/docs` như sau:

#### `POST /jobs/crawl`
**Payload:**
```json
{
  "job_id": "uuid",
  "university_id": "string",
  "callback_url": "http://be-endpoint.internal",
  "metadata": {
    "name": "TU Delft",
    "country": "Netherlands",
    "qs_rank": null,
    "ielts_min": null,
    "sat_required": null,
    "gpa_expectation_normalized": null,
    "tuition_usd_per_year": null,
    "scholarship_available": null,
    "scholarship_notes": null,
    "application_deadline": null,
    "available_majors": null,
    "acceptance_rate": null
  }
}
```
**Response (200 OK):**
```json
{ "accepted": true, "job_id": "uuid-from-BE" }
```

#### `POST /jobs/analyze`
**Payload:**
```json
{
  "job_id": "uuid",
  "case_id": "string",
  "callback_url": "http://be-endpoint.internal",
  "input": {
    "full_name": "Nguyen Linh",
    "gpa_normalized": 3.6,
    "ielts_overall": 7.0,
    "sat_total": null,
    "intended_major": "Computer Science",
    "budget_usd_per_year": 35000,
    "preferred_countries": ["UK", "NL"],
    "target_intake": "Fall 2026",
    "scholarship_required": false,
    "extracurriculars": "Hackathon winner x2...",
    "achievements": "Dean's list 2023"
  }
}
```
**Response (200 OK):**
```json
{ "accepted": true, "job_id": "uuid" }
```

#### `POST /jobs/report`
**Payload:**
```json
{
  "job_id": "uuid",
  "case_id": "string",
  "callback_url": "http://be-endpoint.internal",
  "student_name": "Nguyen Linh",
  "recommendations": [ { "id": "rec..", "university_id": "uni..." } ]
}
```
**Response (200 OK):**
```json
{ "accepted": true, "job_id": "uuid" }
```

#### `DELETE /jobs/university/{university_id}`
Dùng cho web sync của Backend -> remove hoàn toàn Node trên Neo4j.
**Response (200 OK):**
```json
{ "deleted": true, "university_id": "uuid" }
```

---

## Integration Document (Dành cho Backend Team)

**1. Neo4j & MCP Dependency (QUAN TRỌNG):**
- AI Service **yêu cầu tiên quyết** `mcp-neo4j-cypher` server cần phải chạy trước, ở mạng ảo (docker) nội bộ hoặc Local bằng lệnh:
  ```bash
  mcp-neo4j-cypher --transport http --server-host 127.0.0.1 --server-port 8081 --server-path /api/mcp/
  ```
- **Không có server này**: Crawl Worker sẽ Exception và liên tục trả Callback fail về BE.

**2. Giao thức Fire & Forget:**
- BE không mở HTTP connection giữ nguyên chờ AI (Timeout 15s).
- AI luôn response cứng mã HTTP 200 kèm `{"accepted": true}` ngay lập tức sau khoảng ~5ms.
- Toàn bộ outcome sẽ được AI **gọi ngược lại Backend** thông qua URL string truyền trực tiếp ở field `callback_url` nằm trong payload lúc BE gửi request đi. 

**3. Format Callback Trả Về (BE Cần handle POST router):**
```json
{
  "job_id": "uuid",
  "job_type": "crawl_university", /* hoặc "analyze_profile", "generate_report" */
  "status": "done",               /* hoặc "failed" */
  "university_id": "uuid",      
  "case_id": "uuid", /* Trả về đối với case Analyze Profile và Report */
  "error": null,                  
  "result": { ... }               
}
```
- Nếu `status = "failed"`, BE cần check field `"error"` để tracking lỗi TinyFish Agent hay network gián đoạn.

**4. Khởi tạo Node (Crawl Requirement):**
- Trong `metadata` được gửi từ Backend lúc crawl, những fields mang giá trị **`null`** sẽ được Agent đánh giá là Target Goal cần tìm kiếm.
- Trường đã có Value từ Database Backend (Ví dụ: `name`, `country`), AI sẽ tham chiếu và merge thông tin lại chứ không dò trên Web. Hãy điền data càng đủ từ phía Backend để tiết kiệm Token Cost tối đa.
