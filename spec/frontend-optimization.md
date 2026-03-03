# Frontend Performance Optimization

## Context

The app feels slow on Render's cheapest tier. While the server is constrained, the frontend has significant self-inflicted performance issues that compound the problem. An audit of `modules/client` found cascading re-renders, zero memoization, client-side business logic, and no data caching — all fixable without a framework switch.

## Should we switch to SolidJS?

**No — not yet.** The current slowness is from React anti-patterns, not React itself. Fixing these would be far less effort than rewriting 52 pages + all components + replacing Mantine (no SolidJS equivalent). Revisit only if performance is still inadequate after these optimizations.

---

## Phase 1: High-impact, low-effort wins

### 1. Flatten context provider cascade
**File**: `app/layout.jsx` (lines 37-59)

Currently 11 nested providers — any context update re-renders the entire tree. Consolidate into 2-3 grouped providers (auth, app-state, ui) using composition pattern.

### 2. Add `useMemo` / `React.memo` to hot paths
- `automation/page.jsx:552-555` — `documentOptions` array recreated every render
- `components/DataTable/DataTable.jsx` — no `React.memo` on reusable table component
- `components/ProposalQueue/ProposalQueue.jsx:26-30` — `operationColors` already at module scope (confirmed no change needed)

### 3. Debounce search inputs
- `components/SearchBox/SearchBox.jsx` already implements debounced preview and Enter-to-search — **no change needed**.

### 4. Remove `pg` from client dependencies
**File**: `package.json` — PostgreSQL driver has no business in the client bundle.

---

## Phase 2: Data fetching improvements

### 5. Fix fetch waterfalls
- `usage/page.jsx:195-212` — `getProviderCosts()` fetched in separate `useEffect` instead of joining the existing `Promise.all`

---

## Phase 3: Code cleanup (reduces bundle + cognitive load)

### 6. Deduplicate automation form logic
`automation/page.jsx` — `handleOpenEdit()` (lines 251-325) and `handleDuplicate()` (lines 327-399) share ~75 lines of identical form transformation. Extract shared helper.

### 7. Move formatting utils out of components
`usage/page.jsx:30-73` — `formatTokens()`, `formatCost()`, `formatAvg()` are pure functions defined inline. Move to a shared util.

### 8. Replace `axios` with native `fetch`
Axios adds ~85KB gzipped. The app already has an `api/index.js` wrapper — swap the underlying transport.

---

## Out of scope (noted for future)
- Enable `reactStrictMode` in `next.config.js`
- Split 600-900 line components (`LeadForm.jsx`, `Chat.jsx`, `AccountForm.jsx`)
- Explore server components for read-heavy pages (usage, automation list)
- Remove duplicate image assets in `/public/`
- Replace raw `useEffect` fetching with SWR or React Query (larger effort, separate initiative)

---

## Verification
1. Run `next build` — check bundle size report before/after
2. Use React DevTools Profiler on automation and usage pages — confirm reduced re-render count
3. Lighthouse performance score on key pages
4. Manual test: navigation between pages should feel snappier
