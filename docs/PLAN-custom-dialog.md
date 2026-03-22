# PLAN: Custom Confirmation Dialog

This plan covers the implementation of a custom confirmation dialog in the frontend to replace the native browser `confirm()`. This will provide a more premium and integrated user experience.

## User Review Required
> [!IMPORTANT]
> Please confirm the desired design approach:
> 1. Should it be a **Global Helper** (e.g., `await confirm(...)`) or a **Component-based** approach (adding `<ConfirmDialog />` to each page)?
> 2. Do you want different visual styles for "Danger" (e.g., Delete University) vs "Warning" (e.g., Re-analyze)?

## Proposed Changes

### [Component] [NEW] [ConfirmDialog.vue](file:///d:/CODE/UniMatch-Copilot/frontend/src/components/ConfirmDialog.vue)
- A reusable modal component using current project aesthetics (card-soft, backdrop blur).
- Props: `title`, `message`, `confirmLabel`, `cancelLabel`, `type` (danger/warning).

### [Composable] [NEW] [useConfirm.js](file:///d:/CODE/UniMatch-Copilot/frontend/src/composables/useConfirm.js)
- A global service/composable to trigger the dialog programmatically.
- Uses a Promise-based API for `await confirm(...)` usage.

### [View] [MODIFY] UniversityKBView.vue, CasesView.vue, CaseDetailView.vue
- Replace calls like `if (!confirm(...)) return` with the new custom dialog.

## Verification Plan
### Automated Tests
- Check if dialog appears on click.
- Verify "Cancel" and "Confirm" return correct values.
### Manual Verification
- Test deleting a university in Knowledge Base.
- Test "Re-analyze" in Case Details.
