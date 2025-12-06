# Remove TypeScript Entirely

## Problem

TypeScript adds significant boilerplate (interfaces, type declarations, generics syntax) that bloats code and degrades developer experience. We want to eliminate it as a dependency entirely.

## Proposed Solution: Plain JavaScript + Runtime Validation

Remove TypeScript completely. Use Zod for runtime validation at API boundaries where type safety matters most.

### Type Checking Options (No TypeScript)

| Option | Enforcement | Runtime Cost | Tooling |
|--------|-------------|--------------|---------|
| **Zod** | Runtime validation | Yes (at boundaries) | Good IDE inference |
| **JSDoc only** | IDE hints only | None | VS Code support |
| **None** | None | None | Simplest |

### Recommended: Zod at API Boundaries

```javascript
import { z } from 'zod';

// Define schema once, get validation + IDE autocomplete
const UserSchema = z.object({
  id: z.number(),
  name: z.string(),
});

// Validate incoming data
function createUser(data) {
  const user = UserSchema.parse(data); // throws if invalid
  return user;
}

// Or safe parse
const result = UserSchema.safeParse(data);
if (!result.success) {
  console.error(result.error);
}
```

### Benefits

- No build step, no compilation
- Actual runtime validation (catches real bugs TS can't)
- IDE autocomplete via Zod's type inference
- Zero TypeScript dependency
- Schemas double as documentation

### Where to Use Zod

- API request/response validation
- Form data validation
- External data (localStorage, API responses)
- Configuration objects

### Where Types Don't Matter

- Internal function calls (trust your own code)
- Simple utilities
- UI components (props are self-documenting)

## Migration Steps

1. `npm uninstall typescript @types/*` (all @types packages)
2. Delete `tsconfig.json`
3. Rename all `.ts` → `.js`, `.tsx` → `.jsx`
4. Remove type annotations, interfaces, type imports
5. Convert TypeScript enums to plain objects
6. Add `zod` for API boundary validation
7. Update build scripts (vite/esbuild handle JS natively)
8. Configure `jsconfig.json` for IDE path resolution (optional)

## jsconfig.json (Optional, for IDE)

```json
{
  "compilerOptions": {
    "baseUrl": ".",
    "paths": {
      "@/*": ["src/*"]
    }
  }
}
```

## Files to Migrate

- [ ] Audit `modules/web/src` for all `.ts`/`.tsx` files
- [ ] Identify files with complex types (prioritize for Zod)
- [ ] Simple files: just rename and strip types
- [ ] Remove all `@types/*` from package.json

## Code Rules (Still Apply)

The following rules from `/ts-code-rules` remain in effect for JavaScript:

- **KISS** - Keep It Simple
- **YAGNI** - Don't add speculative features
- **SOLID (functional)** - Single responsibility, composable functions
- **ECMAScript 2023** - Use modern JS features (optional chaining, nullish coalescing, array methods)
- **Guard clauses** - Avoid `else`, nested conditionals; early return instead
- **No break/continue** - Use functional patterns
- **Avoid useRef** - Except for legitimate DOM refs and WebSocket lifecycle
- **Conciseness** - No boilerplate

Only the TypeScript-specific syntax is removed (interfaces, type annotations, generics syntax).

## Decision

**Use Zod** for runtime validation at API boundaries (request/response parsing, form data, external data sources).
