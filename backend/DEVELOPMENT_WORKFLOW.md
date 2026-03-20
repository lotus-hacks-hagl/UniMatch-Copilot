# Hướng Dẫn Luồng Phát Triển Tính Năng (Development Workflow)

Tài liệu này hướng dẫn Developer cách thêm tính năng hoặc API mới vào backend UniMatch Copilot. Backend tuân thủ kiến trúc **N-Layer Architecture** (còn gọi là kiến trúc Onion/Clean).

---

## 🏗️ Kiến Trúc Thư Mục (Folder Structure)

Project được chia làm các Layer cụ thể, gọi chéo nhau *chỉ theo một chiều* (Từ ngoài vào trong):  
👉 **Handler -> Service -> Repository -> Database**

```text
backend/
├── config/             # Load biến môi trường / constants sys
├── internal/           # Mã nguồn chính của Business Logic
│   ├── dto/            # Data Transfer Object (Struct Requests, Responses)
│   ├── handler/        # Controller, nhận HTTP Request, Validate params, bọc Response (Tầng mỏng)
│   ├── middleware/     # Các interceptors HTTP như Logging, CORS, Auth
│   ├── model/          # Định nghĩa cấu trúc Database cho GORM
│   ├── repository/     # Logic tương tác Database CRUD trực tiếp
│   ├── router/         # Đăng ký các Endpoint HTTP
│   └── service/        # Nơi chứa Business Logic (Tầng Dày) lõi của ứng dụng
├── migrations/         # (Reference) SQL migrations dùng cho reference
├── pkg/                # Các package tiện ích dùng chung (Database Core, Errors, Custom Response)
└── main.go             # Entrypoint & Dependency Injection
```

---

## 🛠️ Step-by-Step Thêm 1 Tính Năng Mới (Add a new Feature)

Ví dụ: Bạn cần thêm chức năng "Quản lý Counselors (Cố vấn viên)". Bạn sẽ thực hiện code theo trình tự sau:

### Step 1: Định nghĩa Model Database
- Tạo file: `internal/model/counselor.go`
- Viết struct mô tả các field, mixin `BaseModel`, và add các tags chuẩn GORM (`gorm:"type:varchar(255)"`).
- **Lưu ý**: Sẽ **không cần** viết file create table SQL thủ công rườm rà. Bạn qua file `pkg/database/postgres.go` và thả `&model.Counselor{}` vào param của hàm `AutoMigrate()`. Lần chạy tiếp theo, DB sẽ tự có table.

### Step 2: Định nghĩa DTO
- Tạo file: `internal/dto/counselor_dto.go`
- Ở đây, bạn định nghĩa các struct quy định dữ liệu Đầu vào (vd: `CreateCounselorRequest`) và Đầu ra (vd: `CounselorResponse`).
- Gắn các thẻ validator (VD: ``validate:"required,email"``) vào field nếu cần validate.

### Step 3: Xây dựng Repository Layer
- Tạo file: `internal/repository/counselor_repository.go`
- Cấu trúc: 
  - Khai báo biến global Interface `CounselorRepository` tại `interfaces.go`.
  - Triển khai struct `counselorRepository` và inject con trỏ `gorm.DB` từ bên ngoài vào.
- **Quy tắc**: File này *Tuyệt đối* KHÔNG được nhận Request của HTTP từ thư viện `gin`, chỉ thao tác lấy, sửa, xóa với `model.Counselor`. Mọi lệnh `db.Create` / `db.Where` đều nằm ở đây.

### Step 4: Xây dựng Service Layer (Business Logic)
- Tạo file: `internal/service/counselor_service.go`
- Tương tự như Repo, khai báo Interface tại `interfaces.go`. Inject con trỏ Repository vào lúc khởi tạo Service.
- **Quy tắc**: Code xử lý logic "Khi Counselor đăng ký mới, phải gửi email, tạo log hoạt động..." nằm ở tầng này.
- Các hàm ở Service phải trả về 2 kết quả: `(KếtQuảThànhCông, LỗiCustomAppError)`. Không throw `err` gốc của Gorm lên handler. Chuyển nó thành `apperror.BadRequest()` hoặc `apperror.Internal()`. 

### Step 5: Xây dựng Handler (Controller Layer)
- Tạo file: `internal/handler/counselor_handler.go`
- Handler cực kỳ mỏng. Quy tắc `AAA` của Layer HTTP: 
  - **A** (Bind & Validate Struct DTO có lỗi không).
  - **A** (Call các func của Service).
  - **A** (Dùng package `response` để wrap object trả về hoặc báo lỗi `response.Fail(...)`).
- **Bắt buộc**: Viết annotation block comment của Swaggo (`// @Summary ...`) ngay phía trên hàm handler để Swagger sinh docs.

### Step 6: Khai báo Route & Main Binding
1. Cấu hình vào hệ thống mạng HTTP (Router):
   - Mở `internal/router/router.go`.
   - Tạo Route group mới `counselors := api.Group("/counselors")`.
   - Add routes: `counselors.POST("", handler.Create)`.
2. Truyền Dependencies ở `main.go`:
   - Inject Database connection vào Init của Repo.
   - Inject Repo vào Init của Service.
   - Inject Service vào Init của Handler.
   - Đẩy Handler vào Router setup hàm main.

### Step 7: Cập nhật Swagger Docs
Sau khi đã thêm logic code thành công và build run được, chạy lệnh sau ở terminal để scan toàn bộ annotations thành UI Document:
```bash
make swagger
```
*(Lệnh này tương đương với `swag init -pd`)*

---

## 📝 Best Practices 
Trong dự án này, Developer cần Follow các Rules sau:
1. **Tuyệt đối Return Custom Error của ứng dụng**: Use `apperror` package — nó chứa sẵn HTTP Status Codes. Đừng return thư viện lỗi của `sql`.
2. **Luôn Inject `context.Context`**: Vì sau này có request cancel hoặc query tracing.
3. **Transaction an toàn**: Đối với các flow create/update nhiều bảng cùng lúc, phải bao bọc trong `db.Transaction(func(tx *gorm.DB) error { ... })`. (Ví dụ hàm `Create` của Case Service).
4. **Không viết SQL cứng trừ Analytics RawQuery**: Phải sử dụng GORM API (`.Where()`, `.Joins()`, `Preload()`). Ngoại trừ lúc join phức tạp cho Chart thống kê hoặc reporting.
