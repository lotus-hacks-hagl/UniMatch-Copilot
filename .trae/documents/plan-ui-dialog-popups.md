# /plan — Chuẩn hoá dialog/popup UI (không dùng alert/confirm mặc định)

## Mục tiêu

- Tất cả dialog/popup (đặc biệt **delete confirm**, re-analyze confirm, report confirm nếu có) phải dùng **component modal đẹp**, style đồng bộ theo UI hiện tại.
- Loại bỏ hoàn toàn `window.alert()` và `window.confirm()` trong frontend.
- Thông báo kết quả (success/error) dùng **ToastStack** (đã có), không dùng alert.

## Hiện trạng (điểm dùng popup mặc định)

Các file đang dùng `alert()/confirm()`:
- `CaseDetailView.vue` (claim/update summary/generate report/add note/reanalyze confirm)
- `CasesView.vue` (claim)
- `NewCaseView.vue` (submit fail)
- `StudentsView.vue` (update/delete confirm)
- `TeacherManagementView.vue` (verify fail)
- `UniversityKBView.vue` (delete confirm)

Ngoài ra đã có pattern modal inline theo style website tại:
- `UniversityKBView.vue` (modal confirm/delete style đẹp)
- `StudentsView.vue` (modal style đẹp)

## Định hướng UI/UX thống nhất

### 1) Dialog types (MVP)

1. **ConfirmDialog** (Yes/Cancel)
   - Variants: `default` / `danger`
   - Dùng cho: delete, re-analyze, các action có side effects.
2. **AlertDialog** (OK only) (tuỳ chọn)
   - Nếu nội dung cần user acknowledge. Tuy nhiên ưu tiên Toast cho success/error.

### 2) Quy chuẩn hành vi

- Mở/đóng bằng state controlled (không dùng `confirm()`).
- `Esc` để đóng (trừ khi đang loading submit).
- Click backdrop để đóng (trừ khi `danger` + cần bắt confirm rõ ràng, tuỳ UX).
- Khoá scroll nền khi modal mở.
- Focus management:
  - focus vào nút “Cancel” hoặc “Confirm” theo variant
  - `role="dialog"`, `aria-modal="true"`, `aria-labelledby`, `aria-describedby`

### 3) Quy chuẩn style (dùng class có sẵn)

- Backdrop: `bg-black/40 backdrop-blur-sm`
- Card: dùng `card-soft` + `rounded-[24px] p-8 shadow-2xl border border-black/5`
- Buttons:
  - Primary: `.btn-primary`
  - Outline: `.btn-outline`
  - Danger: bổ sung class `.btn-danger` theo style hệ thống (đỏ, outline nhẹ)

## Thiết kế component & API

### 1) Component mới

Tạo các component tại `frontend/src/components/dialogs/`:

1. `BaseModal.vue`
   - Props: `open`, `titleId`, `closeOnBackdrop`, `closeOnEsc`, `zIndex`
   - Emits: `close`
   - Nội dung slot: `header`, `default`, `footer`

2. `ConfirmDialog.vue`
   - Props:
     - `open: boolean`
     - `title: string`
     - `message: string`
     - `variant: 'default' | 'danger'`
     - `confirmText`, `cancelText`
     - `loading: boolean`
   - Emits: `confirm`, `cancel`, `close`
   - Dùng `BaseModal` bên dưới

3. (Optional) `AlertDialog.vue`
   - Props: `open`, `title`, `message`, `okText`
   - Emits: `ok`, `close`

### 2) Dialog controller (composable + host)

Tạo `frontend/src/composables/useDialog.js`:
- `confirm({ title, message, variant, confirmText, cancelText }) -> Promise<boolean>`
- (optional) `alert({ title, message, okText }) -> Promise<void>`

Tạo `frontend/src/components/dialogs/DialogHost.vue`:
- Mount 1 lần ở `App.vue` (giống ToastStack)
- Render `ConfirmDialog` dựa trên state của composable (reactive singleton)

Lợi ích:
- Các view chỉ gọi `await dialog.confirm(...)` thay cho `confirm()`
- Style và hành vi dialog thống nhất toàn app

## Kế hoạch refactor theo file

### 1) Thay `confirm()` bằng ConfirmDialog

- `UniversityKBView.vue`
  - thay `confirm(t('universityKb.confirmDelete'))` bằng `await dialog.confirm({ variant:'danger', ... })`
- `StudentsView.vue`
  - thay confirm delete bằng dialog confirm; toast cho success/fail
- `CaseDetailView.vue`
  - thay confirm re-analyze bằng dialog confirm (variant danger hoặc default)

### 2) Thay `alert()` bằng Toast (chuẩn hoá thông báo)

Áp dụng ToastStack (đã có) cho:
- `CasesView.vue` (claim success/fail)
- `CaseDetailView.vue` (claim/update summary/report/add note/reanalyze)
- `NewCaseView.vue` (submit fail)
- `TeacherManagementView.vue` (verify fail)

Chuẩn hoá message:
- Bổ sung i18n keys (en/vi) cho các message đang hardcode
- Toast types: success / error / info

### 3) Chuẩn hoá “avatar/name không null”

Trong các chỗ hiển thị name/avatar:
- Luôn dùng helper `displayName(name)` fallback `"Unnamed student"`
- `getAvatar(name)` fallback `'??'`
(đảm bảo không render “UNDEFINED”)

## Testing/Verification

### 1) Manual checklist

- Delete flow (Universities, Students): dialog mở đúng style, confirm/cancel đúng, toast hiển thị.
- Re-analyze flow trong case detail: dialog confirm, không dùng browser confirm.
- Report generation + note add: không dùng alert, dùng toast.
- Keyboard: Esc đóng, Tab focus trong dialog.

### 2) Build & lint

- `npm run build` (frontend) không lỗi.

### 3) E2E (nếu Playwright đã có)

- Update/extend test để:
  - Click delete → dialog xuất hiện → cancel/confirm
  - Assert dialog DOM thay vì browser dialog

## Deliverables

- Bộ dialog components + host + composable confirm.
- Toàn bộ view chuyển sang dialog/toast, xoá sạch `alert()/confirm()`.
- i18n strings cho popup/toast.
- Checklist test pass.

