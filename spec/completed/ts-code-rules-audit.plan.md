# TypeScript Code Rules Audit - modules/client

**Date:** 2025-01-XX  
**Scope:** `modules/client/`  
**Rules:** KISS, YAGNI, SOLID (functional), ECMAScript 2023, Guard Clauses, No break/continue, Conciseness, Avoid useRef

## Summary

Found violations across multiple categories:
- **useRef usage:** 4 instances (all appear necessary for DOM/WebSocket refs)
- **Else statements:** 2 instances that could use guard clauses
- **ECMAScript 2023 conformance:** ✅ Compliant (tsconfig.json target: ES2023)

## Violations by Category

### 1. useRef Usage (4 instances)

**Rule:** React code should avoid useRef.

**Note:** All instances appear to be necessary for:
- DOM element references (click-outside detection, scroll management)
- WebSocket lifecycle management

#### `modules/client/components/NotificationBell/NotificationBell.tsx`

**Line 22** - DOM ref for click-outside detection:
```typescript
const bellRef = useRef<HTMLDivElement>(null);
```

**Usage:** Lines 43, 128 - Used to detect clicks outside dropdown.

**Assessment:** This is a legitimate use case for DOM refs. However, could potentially use a callback ref pattern or a custom hook that doesn't expose useRef directly.

#### `modules/client/hooks/useWs.tsx`

**Line 134** - WebSocket reference for lifecycle management:
```typescript
const wsRef = useRef<WebSocket | null>(null);
```

**Usage:** Lines 140, 144, 152, 183, 191, 204 - Used to track WebSocket instance across re-renders and cleanup.

**Assessment:** This is necessary for WebSocket lifecycle management. The ref ensures the correct socket instance is used even if the component re-renders. This is a legitimate use case.

#### `modules/client/components/Chat/Chat.tsx`

**Lines 1, 332, 341, 359** - Multiple refs for DOM elements:
```typescript
import React, { useEffect, useRef, useState } from "react";
// ...
const listRef = useRef<HTMLDivElement>(null);
const toolsDropdownRef = useRef<HTMLDivElement>(null);
const demoDropdownRef = useRef<HTMLDivElement>(null);
```

**Usage:** 
- `listRef` (line 332) - Used for scrolling chat messages
- `toolsDropdownRef` (line 341) - Click-outside detection for tools dropdown
- `demoDropdownRef` (line 359) - Click-outside detection for demo dropdown

**Assessment:** These are legitimate DOM refs for scroll management and click-outside detection. However, they violate the "avoid useRef" rule. Consider using callback refs or a custom hook pattern.

---

### 2. Guard Clause Rule (2 violations)

**Rule:** Use guard clauses; avoid nested conditionals, `else`, and `switch/case`.

**Note:** Switch statements in `reducers/` directory are excluded per rules.

#### `modules/client/components/DocumentsSection/DocumentsSection.tsx`

**Lines 237-242** - Else statement in useEffect:
```typescript
useEffect(() => {
  if (isCreateMode && pendingDocumentIds && pendingDocumentIds.length > 0) {
    getPendingDocuments().then(setPendingDocuments);
  } else {
    setPendingDocuments([]);
  }
}, [isCreateMode, pendingDocumentIds]);
```

**Fix:** Use guard clause:
```typescript
useEffect(() => {
  if (!isCreateMode || !pendingDocumentIds || pendingDocumentIds.length === 0) {
    setPendingDocuments([]);
    return;
  }
  getPendingDocuments().then(setPendingDocuments);
}, [isCreateMode, pendingDocumentIds]);
```

#### `modules/client/components/Documents/DocumentUpload.tsx`

**Lines 127-139** - If/else for version vs new document:
```typescript
if (asVersion && existingDocument) {
  // Create as new version
  document = await createDocumentVersion(existingDocument.id, file);
} else {
  // Create as new document
  document = await uploadDocument(
    file,
    tags.length > 0 ? tags : undefined,
    entityType,
    entityId,
    description || undefined
  );
}
```

**Fix:** Use guard clause:
```typescript
if (asVersion && existingDocument) {
  document = await createDocumentVersion(existingDocument.id, file);
  clearInterval(progressInterval);
  setProgress(100);
  // ... reset form ...
  onUploadComplete?.(document);
  return;
}

// Create as new document
document = await uploadDocument(
  file,
  tags.length > 0 ? tags : undefined,
  entityType,
  entityId,
  description || undefined
);
```

---

### 3. ECMAScript 2023 Conformance

**Status:** ✅ Compliant

**Evidence:** `tsconfig.json` shows:
```json
{
  "compilerOptions": {
    "target": "ES2023",
    "lib": ["ES2023", "DOM", "DOM.Iterable"]
  }
}
```

---

### 4. Break/Continue Usage

**Status:** ✅ No violations found

---

### 5. Switch/Case Usage

**Status:** ✅ No violations (excluded for `reducers/` directory)

**Found in:**
- `modules/client/reducers/navReducer.ts`
- `modules/client/reducers/userReducer.ts`
- `modules/client/reducers/projectReducer.ts`
- `modules/client/reducers/pageLoadingReducer.ts`

**Note:** These are excluded from guard clause rule per code rules.

---

## Recommendations

### High Priority

1. **Refactor else statements to guard clauses:**
   - `DocumentsSection.tsx` lines 237-242
   - `DocumentUpload.tsx` lines 127-139

### Medium Priority

2. **Review useRef usage:**
   - `NotificationBell.tsx` - Consider using callback ref or custom hook
   - `useWs.tsx` - Keep as-is (necessary for WebSocket lifecycle)
   - `Chat.tsx` - Multiple refs for scroll and click-outside detection; consider callback refs or custom hook

### Low Priority

3. **Code quality improvements:**
   - Most code already follows functional patterns well
   - Guard clause violations are minor and easy to fix

---

## Files Reviewed

- `modules/client/components/DocumentsSection/DocumentsSection.tsx`
- `modules/client/components/Documents/DocumentUpload.tsx`
- `modules/client/components/CommentsSection/CommentsSection.tsx`
- `modules/client/components/NotificationBell/NotificationBell.tsx`
- `modules/client/components/Chat/Chat.tsx`
- `modules/client/hooks/useWs.tsx`
- `modules/client/app/accounts/organizations/page.tsx`
- `modules/client/app/accounts/individuals/page.tsx`
- `modules/client/app/accounts/contacts/page.tsx`
- `modules/client/reducers/*.ts` (excluded from guard clause rule)
- `modules/client/tsconfig.json`

## Notes

- **useRef exceptions:** The useRef usage found appears to be necessary for:
  - DOM element references (click-outside detection, scroll management)
  - WebSocket lifecycle management (preventing stale closures)
  
  These are legitimate use cases, though the rule suggests avoiding useRef. Consider documenting these exceptions or using alternative patterns (callback refs, custom hooks) if possible.

- **Ternary operators:** Some `else` statements found are actually ternary operators in JSX (e.g., `org.website ? <a>...</a> : "-"`), which are acceptable patterns.

- **Overall code quality:** The codebase generally follows functional programming patterns well. Violations are minor and easy to address.

