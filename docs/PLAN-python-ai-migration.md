# Integration Plan: Python FastAPI AI Service Migration

## 1. Context & Architecture Review

Dựa vào `docs/implementation_plan.md` cũ và Swagger JSON mới của **Python AI Service**, kiến trúc mới hiện tại đã đi đúng định hướng thiết kế:
- **Backend (Go)** đóng vai trò là Orchestrator, lưu trữ trạng thái.
- **AI Service (Python)** đóng vai trò Worker (chạy nền các subagents). Nó đã expose đủ các Endpoint như `/jobs/crawl`, `/jobs/analyze`, `/jobs/{job_id}` và kèm theo endpoint Debug Graph `/graph/university/{university_id}`.

### Đánh giá mức độ tích hợp (Evaluation)
**1. Những phần đã khớp nhau (Sẵn sàng integrate):**
- Flow bất đồng bộ (Fire-and-forget): BE gọi AI Worker kèm theo `callback_url`, nhận lại response HTTP 200/202 ngay lập tức.
- Data Schema cho `CrawlJobRequest` và `AnalyzeJobRequest` khá sát với Models của Go Backend. Backend chỉ cần Serialize đúng chuẩn.

**2. Gaps & Missing Features (Cần bổ sung/sửa đổi):**
- **Sự kiện Delete University**: BE Go khi xóa 1 trường ĐH phải gọi sang AI qua `DELETE /jobs/university/{university_id}` để đồng bộ xóa Node trên Knowledge Graph Neo4j.
- **Status Endpoint (`GET /jobs/{job_id}`)**: Có thể tận dụng cho BE Go để chạy một fallback cron-job kiểm tra (nếu webhook callback thất bại giữa chừng).
- **Callback format**: Cần chốt chính xác Schema của kết quả phân tích gửi từ Python AI về webhook `POST /internal/jobs/done` trên Go Backend.

---

## 2. Step-by-Step Implementation Integration Flow

### Phase 1: Go Backend - API Client Update
- Cập nhật thư mục `internal/client/ai_client.go` trong Backend Go:
  - Thay đổi địa chỉ `BaseURL` trỏ vào container `unimatch-ai-service` (cổng 9000).
  - Khớp lại Struct Payload của hàm `SubmitCrawlJob` theo chuẩn `CrawlJobRequest` ở file Swagger kia. Đặc biệt `metadata` truyền toàn bộ schema của trường.
  - Khớp lại Struct Payload của hàm `SubmitAnalyzeJob` theo chuẩn `AnalyzeJobRequest` và `AnalyzeInput`.

### Phase 2: Go Backend - Sync Graph Event
- Bổ sung ở `internal/handler/universities.go` (Hàm Delete): Khi Admin xóa 1 trường ĐH trên API Backend, BE sau khi xóa ở PostgreSQL phải trigger một HTTP DELETE Request tới `AI_SERVICE/jobs/university/{university_id}`.

### Phase 3: Go Backend - Webhook Handler
- Tái cấu trúc hàm `POST /internal/jobs/done` (Callback webhook).
- Xử lý Payload Parser: Backend phải nhận diện được payload JSON mà Python gọi ngược về để update bảng `cases` (thay đổi trạng thái `done`, cập nhật `profile_summary`, chèn recommendation) hoặc `universities` (cập nhật metadata).

### Phase 4: Frontend (Vue3) - Monitoring Tools (Tùy chọn)
- Frontend vốn dĩ giao tiếp trực tiếp với Backend nên flow chính không hề thay đổi.
- **Feature Mới (Graph Debugger)**: Dành cho Admin/Developer, ta có thể build 1 màn hình nhỏ gọi API qua proxy BE để query `GET /graph/university/{university_id}` và hiển thị trực quan Nodes/Edges của Neo4j.

### Phase 5: End-to-End Orchestration
- Gắn biến môi trường `AI_SERVICE_URL=http://unimatch-ai-service:9000` vào file `docker-compose.yml` của backend Go.
- Boot up Postgres, Neo4j, Redis, Go Backend, Python AI Service.
- Test Run 1: Bấm nút "Crawl University" trên màn trường ĐH (FE) -> Log backend -> Log Python Worker -> Log Neo4j -> Webhook success!
- Test Run 2: Bấm nút "Analyze Case" trên màn Cases (FE) -> Agents phân tích -> Recommend kết quả trả về PostgreSQL màn hình tự Refresh.

---

## 3. Agent Assignments (Dự kiến)
- `backend-specialist`: Nhận thầu Phase 1, Phase 2, Phase 3. 
- `frontend-specialist`: (Tuỳ chọn) Đảm nhận Phase 4 tạo màn hình Graph Debug.
- `orchestrator`: Giám sát test E2E flow.
