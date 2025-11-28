# TypeScript/JavaScript Code Rules Violations Report
## `/modules/client` Directory Scan

**Generated:** 2024-12-19  
**Rules Source:** `ts-code-rules.md`

---

## Summary

- **Total Files Scanned:** 82 TypeScript/JavaScript files
- **Violations Found:** 8 instances across 3 files
- **Files with Violations:** 3

---

## Violations by Category

### 1. GUARD CLAUSE RULE Violations - `else if` Statement

**Rule:** Use guard clauses; avoid nested conditionals, `else`, and `switch/case`.

**Violations Found:** 1 instance

#### `components/AccountForm/AccountForm.tsx`
- **Line 307:** `else if` statement
  ```typescript
  } else if (!isOrganization && rel.type === "organization") {
  ```
  Context: Inside relationship handling logic for linking individuals to organizations via contacts

**Recommendation:** Refactor to use guard clauses with early returns:
```typescript
// Instead of:
if (isOrganization && rel.type === "individual") {
  // ...
} else if (!isOrganization && rel.type === "organization") {
  // ...
}

// Use:
if (isOrganization && rel.type === "individual") {
  // ...
  return;
}
if (!isOrganization && rel.type === "organization") {
  // ...
}
```

---

### 2. React `useRef` Violations

**Rule:** React code should avoid useRef

**Violations Found:** 7 instances across 2 files

#### `components/Chat/Chat.tsx`
- **Line 310:** `const listRef = useRef<HTMLDivElement | null>(null);`
- **Line 311:** `const inputRef = useRef<HTMLTextAreaElement | null>(null);`
- **Line 313:** `const toolsDropdownRef = useRef<HTMLDivElement | null>(null);`
- **Line 314:** `const demoDropdownRef = useRef<HTMLDivElement | null>(null);`

**Usage Context:**
- `listRef`: Used for scrolling message list (line 333-336)
  ```typescript
  useEffect(() => {
    const el = listRef.current;
    if (!el) return;
    el.scrollTop = el.scrollHeight;
  }, [messages]);
  ```
- `inputRef`: Used for autofocus on textarea (line 377)
  ```typescript
  useEffect(() => {
    inputRef.current?.focus({ preventScroll: true });
  }, []);
  ```
- `toolsDropdownRef`: Used for click-outside detection (line 338+)
- `demoDropdownRef`: Used for click-outside detection (line 370+)

**Recommendation:** Consider alternatives:
- For DOM refs: Use callback refs or state-based approaches
- For scroll management: Use `scrollIntoView` or state-based scroll tracking
- For focus management: Use `autoFocus` prop or state-based focus control
- For click-outside: Use event handlers on document or portal-based solutions

#### `hooks/useWs.tsx`
- **Line 76:** `const wsRef = useRef<WebSocket | null>(null);`

**Usage Context:** Used to store WebSocket instance for connection management

**Recommendation:** Consider using state management or closure-based approach instead of ref

---

### 3. `switch/case` Statements

**Rule:** Avoid `switch/case` statements (use guard clauses instead)

**Violations Found:** 3 files in `reducers/` directory

#### `reducers/userReducer.ts`
- **Line 22:** `switch (action.type)`
- Multiple `case` statements (lines 23, 25, 27, 29, 31, 33)

#### `reducers/navReducer.ts`
- **Line 11:** `switch (action.type)`
- Multiple `case` statements (lines 12, 14)

#### `reducers/pageLoadingReducer.ts`
- **Line 14:** `switch (action.type)`
- Multiple `case` statements (lines 15, 17)

**Note:** Per rule EXCLUSIONS: "Exclude GUARD CLAUSE RULE for files in a reducers/ directory in a Next.js project."

**Status:** ✅ **EXCLUDED** - These violations are acceptable per the rules

---

## Files Without Violations

The following files were scanned and found to be compliant:
- All other TypeScript/TypeScript React files in the codebase
- No `break`/`continue` violations found
- No class-based OOP violations found (good adherence to functional programming)
- No plain `else` statements found (only 1 `else if` remaining)

---

## Priority Recommendations

### High Priority
1. **Replace `useRef` usage** in `Chat.tsx` (4 instances) - Core component functionality
   - These are used for DOM manipulation and may require careful refactoring

### Medium Priority
2. **Refactor `else if` statement** in `AccountForm.tsx` (1 instance) - Relationship handling logic
3. **Replace `useRef` usage** in `useWs.tsx` (1 instance) - WebSocket management

### Low Priority
- Switch/case statements in reducers are excluded per rules, no action needed

---

## Compliance Summary

| Rule | Status | Count |
|------|--------|-------|
| Avoid `else` statements | ✅ Mostly Compliant | 1 (`else if`) |
| Avoid `useRef` in React | ❌ Violations | 7 |
| Avoid `switch/case` | ✅ Excluded (reducers/) | 0 (excluded) |
| Avoid `break`/`continue` | ✅ Compliant | 0 |
| Avoid OOP classes | ✅ Compliant | 0 |
| Guard clauses | ✅ Mostly Compliant | 1 violation |

---

## Notes

- **Significant Improvement:** Most `else` statements have been refactored to use guard clauses. Only 1 `else if` remains.
- The codebase generally follows functional programming principles well
- `useRef` violations are primarily for DOM manipulation and WebSocket management, which may require careful refactoring to maintain functionality
- Reducer files correctly use switch/case per the exclusion rule
- Overall code quality is high with minimal violations

---

## Change Log

**2024-12-19:** Updated report after re-scan
- Reduced `else` violations from 6 to 1 (`else if`)
- Confirmed `useRef` violations remain at 7
- All other findings remain consistent
