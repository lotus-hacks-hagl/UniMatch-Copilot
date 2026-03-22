# 🔌 Frontend API Integration Guide

Tài liệu này dành cho **Frontend / Mobile Developers** để tích hợp với Backend API của UniMatch Copilot. Backend tuân thủ nghiêm ngặt chuẩn RESTful, JSON Payload và có tài liệu Swagger hỗ trợ.

---

## 🔗 Thông tin Cơ bản (Basics)
- **Base URL**: `http://localhost:8894/api/v1` (Mặc định cho local)
- **Swagger UI**: `http://localhost:8894/swagger/index.html` (Xem chi tiết Models, tham số truyền vào).
- **Content-Type**: Backend chỉ giao tiếp qua `application/json`.
- **CORS**: Đã được config sẵn ở Backend chặn/bật tuỳ env, hỗ trợ `*` trên môi trường dev.

---

## 📦 Kiến Trúc Response Trả Về (The Response Envelope)

Tất cả các API của UniMatch Copilot đều trả về JSON với cấu trúc thống nhất (bảo đảm bạn viết 1 axios interceptor là bắt được hết lỗi).

**✅ 1. Khi Gọi Thành Công (HTTP 200 / 201)**
```json
{
  "success": true,
  "data": { 
     // Object hoặc Array kết quả nằm ở đây
  },
  "meta": null
}
```

**📄 2. Khi Phân Trang (Pagination)**
Tại các endpoint có List (như `GET /cases` hoặc `GET /universities`):
```json
{
  "success": true,
  "data": [
    { ...item1 },
    { ...item2 }
  ],
  "meta": {
    "page": 1,
    "limit": 20,
    "total": 45,
    "total_pages": 3,
    "has_next": true,
    "has_prev": false
  }
}
```

**❌ 3. Khi Có Lỗi (HTTP 4xx / 5xx)**
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_FAILED", 
    "message": "validation failed",
    "details": "Key: 'CreateCaseRequest.FullName' Error:Field validation for 'FullName' failed on the 'required' tag"
  }
}
```

---

## 🚦 Luồng Tích Hợp Chính: Nộp Hồ Sơ Tư Vấn (The Core Flow)

Flow quan trọng nhất của hệ thống là: Học sinh nộp thông tin cá nhân → Chờ AI phân tích → Ra báo cáo gợi ý trường. Vì AI chạy rất mất thời gian (10-30s), Backend áp dụng kiến trúc **Bất Đồng Bộ (Async)**. 

Frontend bắt buộc phải dùng **Polling** (gọi API liên tục mỗi 3-5 giây) để cập nhật trạng thái Case.

### 👣 Bước 1: Submit Form (Nộp Hồ Sơ)
- **Endpoint**: `POST /cases`
- **Body**: Truyền thông tin học phí, GPA, IELTS, Ngành học... (Xem Swagger để lấy JSON Schema).
- **Response Trả Về**:
  ```json
  {
    "success": true,
    "data": {
      "case_id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
      "status": "processing"
    }
  }
  ```
- **Xử lý trên UI**: Chuyển User sang màn hình Loading/Chờ kết quả, và giữ lấy cái `case_id`.

### 👣 Bước 2: Polling Trạng Thái Case (Lắng nghe AI)
- **Hành động FE**: Viết 1 vòng lặp (hoặc dùng `react-query` refetchInterval = 5000), gọi API sau mỗi 5 giây với cái `case_id`.
- **Endpoint**: `GET /cases/{case_id}`
- **Kiểm tra trường Status**:
  - ⏳ `status === "pending"` hoặc `"processing"`: Giao diện hiện Spinner báo "Đang xử lý hồ sơ...".
  - ✅ `status === "done"`: Lập tức dừng polling. Dữ liệu Report và Recommendation đã có sẵn trong response (field `recommendations` & `profile_summary`). Giao diện chuyển sang màn hình Hiển thị kết quả Đại Học!
  - ⚠️ `status === "human_review"`: AI từ chối tư vấn (điểm quá dị biệt, AI ko chắc chắn). Giao diện hiện báo cáo `"Hồ sơ của bạn đã được chuyển cho Cố vấn viên con người"`.
  - ❌ `status === "failed"`: Lỗi hệ thống. 

### 👣 Bước 3: Xuất Báo Cáo PDF
- Ở màn hình `status === "done"`, FE hiển thị 1 nút **"Sinh file báo cáo PDF"**.
- **Endpoint**: `POST /cases/{case_id}/report`
- **Logic Backend**: API này cũng chạy bất đồng bộ y hệt Bước 1. Nó sẽ trả về 200 `{ "status": "generating" }`.
- **Hành động FE**: Tiếp tục lặp polling `GET /cases/{case_id}` (khoảng 3-5s/lần), và chờ cho đến khi trường `report_data` trong API GET hết bị `null` (Nó sẽ chứa link PDF nội dung `pdf_content`).

---

## 🏫 Luồng Tích Hợp: Quản  Universities Data

### 1. Hiển thị danh sách trường
- **Endpoint**: `GET /universities?search=Havard&country=USA&page=1&limit=20`
- Bắn param search/country tuỳ thích. FE sẽ được hứng object phân trang đầy đủ thông tin: rank QS, Học phí, Yêu cầu Ielts...

### 2. Cập nhật data tất cả các trường bằng AI (Chỉ dùng cho Admin/Cronjob)
- **Endpoint**: `POST /universities/crawl-all`
- Server sẽ tìm ra tất cả các trường có `crawl_status !== "pending"` và `last_crawled_at > 24h` để quăng job cho AI đi cào dữ liệu học phí các trường mới nhất trên Web. 
- API này trigger xong là nhả `200 OK` luôn (Trả về số lượng trường được cho đi crawl bao nhiêu). Admin Panel muốn coi tiến độ độ crawl thì dùng API Dashboard.

---

## 📈 Tích Hợp Màn Dashboard Thống Kê (Dành cho Admin/Counselor)

### 1. Stats tổng quan
- **Endpoint**: `GET /dashboard/stats`
- **Output**: Lấy các con số thẻ Card đầu trang (Số case hôm nay, số case chờ human review, thời gian trung bình AI rep, độ tự tin trung bình AI, số Crawl đang chạy ngầm).

### 2. Biểu đồ Đường (Cases By Day)
- **Endpoint**: `GET /dashboard/cases-by-day`
- Phù hợp nạp thẳng vào `recharts` / `chart.js` (Trả về mảng `[ { "date": "2026-03-20T...", "count": 15 } ]`).

### 3. Biểu đồ Tròn Phân bố quốc gia & Analytics
- **Endpoint**: `GET /dashboard/analytics`
- Lấy top 10 trường được rec nhiều nhất và tỷ lệ chọn Country của sinh viên.

### 4. Bảng Activity Log (Real-time Timeline)
- **Endpoint**: `GET /activity-log?limit=20`
- API này return lịch sử thao tác hệ thống mới nhất (VD: AI phát hiện Havard tăng học phí 2000$, Case #123 được tạo, Case #456 báo lỗi). Rất phù hợp đổ vào UI Timeline.

---

## 🚨 Bảng Mã Lỗi Quan Trọng (Error Codes)
Khi bạn nhận được HTTP Error (VD: 400 Bad Request), hãy check field `error.code` để translate/hiển thị message phù hợp cho người dùng:

| HTTP Status | Error Code | Mô tả |
| :--- | :--- | :--- |
| `400` | `VALIDATION_FAILED` | User quên nhập field bắp buộc, hoặc format sai (vd IELTS < 0). Check `error.details`. |
| `400` | `BAD_REQUEST` | Business validation (Sai type UUID, state không cho phép gọi API...) |
| `404` | `NOT_FOUND` | Không tìm thấy ID Case / Uni trong Database. Thiết kế màn 404 cho User. |
| `500` | `INTERNAL_ERROR` | Server Crash hoặc lỗi Database. Thông báo "Hệ thống bảo trì". |
| `503` | `SERVICE_UNAVAILABLE` | Hệ thống AI Service (Python) bị sập hoặc quá tải. |

---

## 💡 Lời khuyên cho Frontend Developer
1. Đọc file `internal/dto` hoặc truy cập Swagger `localhost:8894/swagger` để copy/paste y nguyên format Input vào Typescript Interfaces của bạn. 
2. Nên build 1 file `api-client.ts` bằng Axios Interceptor để global bắt `{ success: false }` và toast lỗi góc màn hình dựa trên `error.message`.
3. Sử dụng thư viện `SWR` hoặc `React Query` để quản lý logic Polling (Fetching interval) của API `GET /cases/{id}` cho thật sạch sẽ, tránh bị rò rỉ Memory Leaks vì `setInterval`.
