# UniMatch Copilot — Backend

Backend REST API cho UniMatch Copilot, được xây dựng với Go (Golang), Gin Framework, GORM và PostgreSQL. Hệ thống này đóng vai trò là Orchestrator, tiếp nhận request từ Client (Frontend/Mobile) và giao tiếp bất đồng bộ với AI Service.

## 🚀 Tính năng chính
* Ghi nhận và quản lý hồ sơ sinh viên (Cases).
* Lưu trữ và tra cứu thông tin trường Đại học (Universities).
* Thống kê Dashboard & Activity Logs.
* Flow bất đồng bộ: Tạo Case -> Gửi Job AI Analyzer -> AI trả kết quả qua Webhook -> Cập nhật Database.

---

## 🛠️ Yêu cầu hệ thống (Prerequisites)
- [Go 1.22+](https://go.dev/dl/)
- [PostgreSQL 15+](https://www.postgresql.org/download/)
- [swag CLI](https://github.com/swaggo/swag) (Dùng để gen API Docs)
- [Make](https://www.gnu.org/software/make/) (Tùy chọn, để dùng các lệnh tiện ích trong Makefile)

---

## 🏃 Hướng dẫn cài đặt & Chạy dự án (Getting Started)

### 1. Cấu hình biến môi trường
Môi trường mặc định được cung cấp trong file `.env.example`. Chạy lệnh copy để tạo file config thật:
```bash
cp .env.example .env
```
Mở `.env` và tùy chỉnh `DATABASE_URL` theo username/password PostgreSQL của bạn:
```ini
DATABASE_URL=postgres://your_user:your_password@localhost:5432/unimatch_be?sslmode=disable
```

Lưu ý callback AI:
```ini
PORT=8080
PUBLIC_BASE_URL=
INTERNAL_BASE_URL=
```
Nếu để trống `PUBLIC_BASE_URL`, backend sẽ tự dùng `http://localhost:<PORT>`. Chỉ set `PUBLIC_BASE_URL` khi AI Service phải callback qua một host khác như Docker bridge, LAN IP, hoặc ngrok. Nếu `PUBLIC_BASE_URL` vẫn trỏ nhầm `8080` trong khi backend chạy cổng khác, callback `POST /internal/jobs/done` sẽ fail.
`INTERNAL_BASE_URL` là địa chỉ backend mà AI Service thật sự gọi tới để callback. Trong local có thể để trống để fallback về `http://localhost:<PORT>`. Trong Docker Compose phải set kiểu `http://backend:8080`, vì `localhost` bên trong container AI không trỏ tới container backend.

### 2. Khởi tạo Database và Migrate cơ bản
Hệ thống sử dụng **GORM AutoMigrate**, nên bạn *chỉ cần tạo sẵn một database rỗng* tên là `unimatch_be`. Code sẽ tự động generate Tables khi chạy.

Nếu bạn dùng lệnh (ở terminal):
```bash
createdb unimatch_be
```

### 3. Tải Dependencies & Generate Swagger Docs
Cài đặt thư viện của Go và gen bộ dữ liệu cho Swagger:
```bash
go mod download
make swagger 
# Hoặc chạy lệnh thẳng: swag init -pd
```

### 4. Khởi động Server
```bash
make run
# Hoặc chạy lệnh thẳng: go run main.go
```
*Server mặc định sẽ chạy ở cổng `8080` (Cấu hình qua biến `PORT` trong env).*

---

## 📖 Hướng Dẫn Sử Dụng (How to Use)

### API Documentation (Swagger UI)
Toàn bộ chi tiết về Input/Output của các Request đều có sẵn trực quan tại Swagger.
Truy cập trang Docs sau khi bật server:
👉 **[http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)**

### Workflow Test Thử Backend (Local)
1. Bật server thành công.
2. Nạp dữ liệu trường Đại Học mẫu bằng lệnh ở Terminal thứ 2:
   ```bash
   make seed
   ```
3. Mở Swagger UI lên.
4. Mở API **`POST /api/v1/cases`**:
   - Gửi payload JSON chứa thông tin sinh viên (GPA, IELTS, v.v..).
   - Backend sẽ trả về HTTP 201 cùng `case_id`. State sẽ là `processing`.
5. Đóng vai AI (AI Mocking via Webhook):
   - Mở API **`POST /internal/jobs/done`** trên Swagger.
   - Gắn payload mock vào (xem document của job type `analyze_profile`).
   - Gửi request. Nếu thành công, check lại `GET /api/v1/cases/{id}` bạn sẽ thấy data trường Đại học rớt về.

---

## 🔌 Tích hợp với AI Service (Integration)

Do quá trình crawl data và analyze AI mất nhiều thời gian, Backend giao tiếp với AI bằng **Webhook Asynchronous**.

| Route | Phương thức | Vai Trò |
| :--- | :--- | :--- |
| `[AI_SERVICE_URL]/jobs/analyze` | Gửi POST | Backend ra lệnh cho AI Service phân tích Student Case |
| `[AI_SERVICE_URL]/jobs/crawl` | Gửi POST | Backend ra lệnh cho AI Service crawl data 1 trường |
| `[AI_SERVICE_URL]/jobs/report` | Gửi POST | Backend ra lệnh cho AI Service render PDF Report |
| **`[INTERNAL_BASE_URL]/internal/jobs/done`** | **Nhận POST** | Webhook để **AI báo kết quả về cho Backend**. Nếu `INTERNAL_BASE_URL` để trống, backend sẽ tự fallback về `http://localhost:<PORT>`. Với Docker Compose, dùng `http://backend:8080`. |

> 📚 **DEVELOPMENT WORKFLOW**: Nếu bạn muốn mở rộng source code (thêm API, Update Database), hãy đọc kỹ tài liệu **[DEVELOPMENT_WORKFLOW.md](./DEVELOPMENT_WORKFLOW.md)**.
