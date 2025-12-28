# Frontend Framework & Language Evaluation

## Status: Draft
## Priority: Medium
## Scope: modules/client
## Decision: SvelteKit + JavaScript (SPA mode)

---

## Context

Evaluating migration away from TypeScript + Next.js for the client module. Primary goals:
- **Developer experience**: Simpler tooling, faster iteration
- **Performance**: Faster builds, smaller bundles
- **Maintainability**: Easier onboarding, reduced complexity

### Current Stack
| Component | Current | Notes |
|-----------|---------|-------|
| Framework | Next.js 15.5.5 | App Router, ~20 routes |
| Language | TypeScript 5 | Moderate complexity (Partial, Omit, Record) |
| Auth | @auth0/nextjs-auth0 v4.6.0 | 5 integration points |
| UI | Mantine 7.17 + Tailwind 4 | Heavy use of Mantine Grid |
| State | React Context + Reducer | 4 contexts, no Redux |

---

## Part 1: TypeScript vs JavaScript

### Assessment

The current TypeScript usage is **simple enough to convert** without significant loss:
- No complex generics or conditional types
- Basic interfaces for domain models
- Utility types limited to `Partial<>`, `Omit<>`, `Record<>`

### Options

| Option | Pros | Cons |
|--------|------|------|
| **Keep TypeScript** | IDE autocomplete, catch bugs at compile time, industry standard | Build overhead, config complexity |
| **Plain JavaScript** | Simpler tooling, faster builds | Lose type safety, harder refactoring |
| **JSDoc + JavaScript** | Types without build step, gradual adoption | Verbose, less IDE support than TS |

### Recommendation

**JSDoc + JavaScript** offers a middle ground: remove TypeScript compilation while retaining type hints via JSDoc comments. Modern editors (VS Code) provide autocomplete and type checking with `// @ts-check`.

---

## Part 2: Framework Alternatives

### Option A: SvelteKit (Non-React)

**Auth0 Integration**: Use `@auth0/auth0-spa-js` directly. No official SvelteKit SDK, but straightforward to implement via hooks and stores.

| Aspect | Assessment |
|--------|------------|
| Performance | Excellent - compiles to vanilla JS, no virtual DOM |
| Bundle size | ~40% smaller than React equivalent |
| Learning curve | Moderate - new paradigm but simpler than React |
| DX | Excellent - less boilerplate, built-in stores, transitions |
| Ecosystem | Smaller than React - fewer libraries, components |
| UI Libraries | Skeleton UI, DaisyUI, or Tailwind-based (no Mantine) |

**Grid Alternatives**:
- Skeleton UI has responsive grid system
- CSS Grid + Tailwind (most flexible)

### Option B: Vite + React (Stay React)

Drop Next.js but keep React. Use Vite for builds, React Router for routing.

| Aspect | Assessment |
|--------|------------|
| Performance | Fast builds (Vite), client-side rendering |
| Bundle size | Similar to Next.js without SSR overhead |
| Learning curve | Low - familiar React patterns |
| DX | Good - Vite HMR is fast, simpler config than Next |
| Ecosystem | Full React ecosystem retained |
| UI Libraries | Can keep Mantine |

**Auth0 Integration**: Use `@auth0/auth0-react` SDK (simpler than Next.js SDK).

**Tradeoff**: Lose SSR/SSG. For a CRM app with authenticated routes, this is likely acceptable.

### Option C: Remix (Stay React)

Modern React framework, focuses on web standards and progressive enhancement.

| Aspect | Assessment |
|--------|------------|
| Performance | Good - nested routing, parallel data loading |
| Bundle size | Similar to Next.js |
| Learning curve | Moderate - loader/action patterns |
| DX | Good - simpler than Next.js App Router |
| Ecosystem | Full React ecosystem |
| UI Libraries | Can keep Mantine |

**Auth0 Integration**: Use `remix-auth` with Auth0 strategy or `@auth0/auth0-react`.

### Option D: Astro + React Islands

Static-first with React components where needed.

| Aspect | Assessment |
|--------|------------|
| Performance | Excellent for content - ships zero JS by default |
| Bundle size | Smallest when hydration is selective |
| Learning curve | Low-moderate |
| DX | Good for content sites |
| Ecosystem | Can use React components as islands |

**Not recommended** for this use case - CRM is highly interactive, not content-focused.

---

## Part 3: Auth0 Migration Complexity

| Framework | SDK | Migration Effort |
|-----------|-----|------------------|
| Next.js (current) | @auth0/nextjs-auth0 | N/A |
| SvelteKit | @auth0/auth0-spa-js | High - rewrite auth layer |
| Vite + React | @auth0/auth0-react | Medium - simpler SDK, some refactoring |
| Remix | remix-auth + Auth0 | Medium - different patterns |

### Current Auth0 Touchpoints
1. `lib/auth0.ts` - Auth0Client initialization
2. `middleware.ts` - Route protection
3. `app/layout.tsx` - Auth0Provider wrapper
4. `components/AuthStateManager.tsx` - Session sync
5. `api/auth/` routes - Token handling

---

## Part 4: UI Library Alternatives (Grid-focused)

If moving away from Mantine:

| Library | Grid System | React | Svelte | Notes |
|---------|-------------|-------|--------|-------|
| Tailwind CSS | Flex/Grid utilities | Yes | Yes | Most flexible, no components |
| Skeleton UI | 12-column grid | No | Yes | Svelte-native, similar to Mantine |
| DaisyUI | Tailwind-based | Yes | Yes | Component classes, no JS |
| Radix + Tailwind | DIY with primitives | Yes | No | Accessible primitives |
| shadcn/ui | Tailwind-based | Yes | No | Copy-paste components |

---

## Recommendation Matrix

| Priority | SvelteKit | Vite+React | Remix |
|----------|-----------|------------|-------|
| DX | +++| ++ | ++ |
| Performance | +++ | ++ | ++ |
| Maintainability | ++ | +++ | ++ |
| Migration effort | High | Medium | Medium |
| Auth0 effort | High | Low | Medium |
| Keep Mantine | No | Yes | Yes |

---

## Suggested Path

### If willing to fully migrate (best long-term DX/perf):
**SvelteKit + Skeleton UI + JavaScript**
- Rewrite effort: ~2-3 weeks for current scope
- Auth0: Implement custom auth hooks with auth0-spa-js
- Result: Fastest builds, smallest bundles, simplest mental model

### If prefer incremental change (lower risk):
**Vite + React + Mantine + JSDoc JavaScript**
- Rewrite effort: ~1 week
- Auth0: Swap to @auth0/auth0-react (simpler API)
- Result: Keep React knowledge, faster builds than Next.js, retain Mantine

---

## Implementation Plan: SvelteKit Migration

### Phase 1: Scaffold & Auth (Week 1)
1. Initialize SvelteKit project with adapter-static (SPA mode)
2. Configure Vite + Tailwind
3. Implement Auth0 integration:
   - Create `auth.js` store using `@auth0/auth0-spa-js`
   - Build `+layout.svelte` with auth guards
   - Port `AuthStateManager` logic to Svelte store subscription
4. Set up API client (axios or native fetch)

### Phase 2: Core Layout & Navigation (Week 1-2)
1. Port `Sidebar` and `Header` components
2. Implement routing structure matching current app routes
3. Choose and integrate grid system:
   - **Recommended**: Skeleton UI (closest to Mantine patterns)
   - Alternative: Pure CSS Grid + Tailwind utilities
4. Port `AuthenticatedPageLayout` wrapper

### Phase 3: Components & Pages (Week 2-3)
Priority order based on complexity:
1. **Simple pages first**: Settings, About, static content
2. **Data tables**: Port `DataTable` component → Skeleton Table or custom
3. **Forms**: LeadForm, ContactForm → Svelte forms (simpler than React)
4. **Complex features**: LeadTracker (~880 lines), Documents, Reports

### Phase 4: State & Context Migration
| React Context | Svelte Equivalent |
|--------------|-------------------|
| UserContext | `userStore` (writable store) |
| NavContext | `navStore` (writable store) |
| ProjectContext | `projectStore` (writable store) |
| PageLoadingContext | `loadingStore` (writable store) |

Svelte stores are simpler - no Provider wrappers, direct imports.

### Phase 5: Testing & Cleanup
1. Manual QA of all routes
2. Remove old Next.js client module
3. Update Docker/deployment configs
4. Update CI/CD pipelines

---

## SvelteKit Auth0 Implementation

```javascript
// src/lib/auth.js
import { writable, derived } from 'svelte/store';
import { createAuth0Client } from '@auth0/auth0-spa-js';

export const auth0Client = writable(null);
export const isAuthenticated = writable(false);
export const user = writable(null);
export const loading = writable(true);

export async function initAuth0() {
  const client = await createAuth0Client({
    domain: import.meta.env.VITE_AUTH0_DOMAIN,
    clientId: import.meta.env.VITE_AUTH0_CLIENT_ID,
    authorizationParams: {
      redirect_uri: window.location.origin,
      audience: import.meta.env.VITE_AUTH0_AUDIENCE
    }
  });

  auth0Client.set(client);

  if (window.location.search.includes('code=')) {
    await client.handleRedirectCallback();
    window.history.replaceState({}, '', window.location.pathname);
  }

  isAuthenticated.set(await client.isAuthenticated());
  user.set(await client.getUser());
  loading.set(false);
}

export async function login() {
  const client = get(auth0Client);
  await client.loginWithRedirect();
}

export async function logout() {
  const client = get(auth0Client);
  await client.logout({ returnTo: window.location.origin });
}

export async function getToken() {
  const client = get(auth0Client);
  return client.getTokenSilently();
}
```

---

## File Mapping (Key Components)

| Next.js | SvelteKit |
|---------|-----------|
| `app/layout.tsx` | `src/routes/+layout.svelte` |
| `app/page.tsx` | `src/routes/+page.svelte` |
| `app/[route]/page.tsx` | `src/routes/[route]/+page.svelte` |
| `components/*.tsx` | `src/lib/components/*.svelte` |
| `context/*.tsx` | `src/lib/stores/*.js` |
| `lib/auth0.ts` | `src/lib/auth.js` |
| `api/index.ts` | `src/lib/api.js` |
| `middleware.ts` | `src/hooks.client.js` |

---

## Open Questions

- [ ] Any Next.js-specific features in use that would be hard to replace? (Image optimization, ISR)
- [ ] Timeline constraints for migration?
- [ ] Parallel development (keep Next.js running while building Svelte) or hard cutover?
