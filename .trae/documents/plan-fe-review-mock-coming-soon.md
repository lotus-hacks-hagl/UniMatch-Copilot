# /plan — Review FE: feature chưa hoàn thiện → mock data hoặc “Coming soon”

## 1) Mục tiêu

- Rà soát toàn bộ FE routes/views hiện có, liệt kê **feature chưa hoàn thiện** (thiếu API, thiếu wiring, placeholder, button không action, dữ liệu mock dễ gây hiểu nhầm).
- Với mỗi feature: chọn 1 trong 2 hướng rõ ràng:
  - **A) Mock data** (chỉ dùng khi demo UI/UX, luôn gắn nhãn “Mock” + disabled actions)
  - **B) Coming soon** (hiển thị panel đồng bộ UI, mô tả ngắn, tránh user hiểu nhầm là đã có)
- Ưu tiên “đang có BE support” thì **implement thật** thay vì coming soon (ví dụ activity/note, report generation).

## 2) Kết quả review hiện trạng (frontend/src/views)

### 2.1 CasesView (Dashboard)

- **Search input** chưa wired (không v-model, không lọc).
- **Filter icon button** chưa wired (không action).
- Charts đang hoạt động (đủ demo), stats lấy từ BE.

=> Hướng xử lý đề xuất:
- Implement nhanh **search client-side** theo tên student/major/country (A: real feature).
- Filter icon button: “Coming soon” + toast.

### 2.2 UniversityKBView

- Search input chưa wired.
- CRUD universities đã chạy thật.

=> Hướng xử lý:
- Implement nhanh search client-side (real).

### 2.3 CaseDetailView

- Tab **Documents** hiện đang là mock list “Student_Contract_v1.pdf”, “Transcript_Official.docx” và nút Download không wired → dễ gây hiểu nhầm “đã có upload tài liệu”.
- Tab **Communication** đang hiển thị `caseData.activity_logs` nhưng endpoint chuẩn là `/cases/:id/activity` và note add là `/cases/:id/notes`. Nếu `/cases/:id` không trả `activity_logs`, UI sẽ trống hoặc sai.

=> Hướng xử lý:
- Documents:
  - Nếu chưa làm upload docs: đổi sang **Coming soon** cho “Contract uploads”.
  - Đồng thời hiển thị phần **Report** (Generate/Download) nếu BE đang support.
- Communication:
  - Implement thật: fetch `/cases/:id/activity` và render feed.

### 2.4 StudentsView

- List + edit modal + delete đang chạy thật.
- Không có student detail route (đã remove), nên “click để xem detail” không có.

=> Hướng xử lý:
- Nếu chưa làm student detail: giữ như hiện tại, nhưng Actions nên đủ (edit/delete).
- Nếu muốn “detail”: tạo view + route sau (Coming soon hoặc implement).

### 2.5 AnalyticsView

- Đang hiển thị data thật (placementRate, distribution) và đã bỏ hardcode.
- Một số mục có thể thiếu data → UI nên hiển thị “— / No data yet”.

=> Hướng xử lý:
- Giữ “No data yet” state, không mock.

### 2.6 ReviewQueueView / AuthView / UnverifiedView / TeacherManagementView

- Các view này hoạt động theo flow, không có placeholder lớn.
- TeacherManagementView chỉ cần toast error (đã có).

## 3) Hạng mục implement (đề xuất theo thứ tự)

### Phase 1 — UI component “Coming soon”

- Tạo `ComingSoonPanel.vue` (hoặc `EmptyStatePanel.vue`) dùng chung:
  - Title, description, optional CTA disabled
  - Style theo `card-soft`, icon + border
- Dùng cho:
  - CaseDetail → Documents (contract uploads)
  - CasesView filter button (toast + panel nếu cần)

### Phase 2 — Wiring “real” những chỗ đã có BE

- CaseDetail → Communication:
  - Fetch `GET /cases/:id/activity?page&limit`
  - Render timeline theo `event_type`, `description`, `created_at`
  - Add note `POST /cases/:id/notes` rồi refresh list
- CaseDetail → Documents:
  - Hook report: `POST /cases/:id/report`
  - Nếu có `report_data.pdf_content` thì enable download

### Phase 3 — Search client-side

- CasesView:
  - Thêm `searchTerm` + computed filtered list
  - Lọc theo: student name, intended_major, target_intake
- UniversityKBView:
  - Thêm `searchTerm` + computed filter theo name/location

## 4) Verification

- `npm run build` pass
- Manual smoke:
  - CaseDetail → Documents: không còn mock contract, có coming soon + report real
  - CaseDetail → Communication: add note thấy xuất hiện trong timeline
  - CasesView/UniversityKBView search hoạt động

