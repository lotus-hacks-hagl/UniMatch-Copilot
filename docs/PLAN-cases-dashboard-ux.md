# PLAN: Cases Dashboard UX Revamp & Metrics

## 1. Context & Objectives
The current "CasesView" dashboard has several UX shortcomings:
- **Empty State Dominance:** The empty state (when no cases match) is overly large and visually interrupts the dashboard flow.
- **"Fake" Charts (Metrics):** The bottom analytics charts (Cases per day, Match tier distribution, Escalation trend) are purely decorative SVGs. They lack numeric parameters (thông số), axes, legends, and tooltips, causing them to look unprofessional and "mờ nhòa" (faded/meaningless).
- **Missing Top Level KPIs:** Top-level metrics need to be aggressively anchored at the top to provide instant value to the user.

## 2. Agent Assignments
- **UX/Frontend Specialist (`前端/frontend-specialist`)**: To restructure the Vue template, integrate actual data binding into the charts, and refine the empty state css.

## 3. Task Breakdown (Implementation Steps)

### Phase 1: Layout & Empty State Refinement
- [ ] Reduce the vertical footprint of the `Empty State` card.
- [ ] Move the empty state inside the Table's bounding box so the Page Layout doesn't radically collapse when data is missing.
- [ ] Ensure the 4 Top KPI cards (Cases Today, Avg Processing, Awaiting Review, AI Confidence) are styled distinctly and remain visible.

### Phase 2: Professional Analytics Integration
- [ ] **Chart Enhancement Strategy:** Replace the static SVG decorators with data-bound visualizations. 
  - *Implementation approach to be determined by user (e.g., Vue-ApexCharts vs Custom CSS Bars).*
- [ ] Add numeric overlays (e.g., central text in the Donut chart showing "Total Cases").
- [ ] Add tooltips or hover states to show Exact Data Values (thông số) on hover.
- [ ] Add simple X/Y axis labels or legends to clarify what the user is looking at.

### Phase 3: Data Binding & Polish
- [ ] Connect the `stats` object from `casesStore` directly to the new visual charts so they update reactively.
- [ ] Verify animations and transitions during data loading.

## 4. Verification Check
- User can hover over the "Escalation rate trend" and see actual data points.
- Empty state looks integrated and not like a massive placeholder.
