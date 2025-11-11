# Environment Variable Pattern Comparison

## Source Pattern: `discogs-player-2024`

### Key Features

1. **Runtime Environment Detection** (`lib/get-base-url.ts`)
   ```typescript
   export const getBaseUrl = () => {
       const { host } = window.location;
       const parts = host.split('.');

       // Production: Transform domain to API subdomain
       if (parts.length === 2) {
           parts.unshift('api');
           return `https://${parts.join('.')}`;
       }

       // Development: Use localhost
       if (host === 'localhost:3001') return 'http://localhost:5000';
   };
   ```

2. **Environment Variables** (`.env.local`)
   - All server-side only (no `NEXT_PUBLIC_` prefix)
   - Contains Auth0 credentials
   - Gitignored
   - Used in API routes and server components only

3. **Usage Pattern**
   ```typescript
   // Client-side: Runtime hostname detection
   const baseUrl = getBaseUrl();
   axios.get(`${baseUrl}/api/endpoint`);

   // Server-side: Process environment variables
   const auth0 = new Auth0Client({
       domain: process.env.AUTH0_DOMAIN,
       clientId: process.env.AUTH0_CLIENT_ID,
       // ...
   });
   ```

---

## Implemented Pattern: `pina-colada-co`

### Applied to Authentication

1. **Runtime Environment Detection** (`lib/get-auth-config.ts`)
   ```typescript
   export const getAuthConfig = () => {
       if (typeof window === 'undefined') {
           return { isDev: false, password: DEV_PASSWORD }
       }

       const { host } = window.location

       // Development environment
       if (host === 'localhost:3001' || host === '127.0.0.1:3001') {
           return { isDev: true, password: DEV_PASSWORD }
       }

       // Production environment
       return { isDev: false, password: null }
   }
   ```

2. **Environment Variables** (`.env.local`)
   ```bash
   # Server-side only (no NEXT_PUBLIC_)
   JOBS_PASSWORD="secure-password"
   SUPABASE_URL="..."
   SUPABASE_SERVICE_KEY="..."
   ```

3. **Usage Pattern**
   ```typescript
   // Client-side: Runtime detection
   const { isDev, password: devPassword } = getAuthConfig();

   if (isDev) {
       // Development: Client-side validation
       if (password === devPassword) { /* login */ }
   } else {
       // Production: Server-side API validation
       fetch('/api/auth/login', { ... })
   }

   // Server-side: Process environment variables
   const AUTH_PASSWORD = process.env.JOBS_PASSWORD
   ```

---

## Pattern Alignment

| Aspect | discogs-player-2024 | pina-colada-co | Match |
|--------|---------------------|----------------|-------|
| Runtime detection | ✅ `window.location.host` | ✅ `window.location.host` | ✅ |
| Utility in `lib/` | ✅ `get-base-url.ts` | ✅ `get-auth-config.ts` | ✅ |
| No `NEXT_PUBLIC_` vars | ✅ | ✅ | ✅ |
| `.env.local` for secrets | ✅ | ✅ | ✅ |
| Guard clauses | ✅ | ✅ | ✅ |
| Dev checks `localhost:3001` | ✅ | ✅ | ✅ |
| Server API routes | ✅ `/api/*` | ✅ `/api/auth/login` | ✅ |

---

## Benefits of This Pattern

### 1. **No Build-Time Secrets**
- Secrets never bundled into JavaScript
- Safe to commit `.env.local.example` (with placeholders)
- Real `.env.local` is gitignored

### 2. **Runtime Flexibility**
- Same build works in dev and prod
- Environment detected at runtime
- No need for different builds

### 3. **Developer Experience**
- Simple local development (hardcoded localhost)
- No environment variable configuration needed for dev
- Just `npm run dev` and it works

### 4. **Security**
- Production passwords only on server
- Client never has access to sensitive credentials
- API routes validate server-side

### 5. **Maintainability**
- Clear separation of concerns
- Utility functions are reusable
- Follows existing project patterns

---

## Extensibility

This pattern can be extended for other environment-dependent features:

### API Base URLs
```typescript
// lib/get-api-url.ts
export const getApiUrl = () => {
    if (typeof window === 'undefined') return 'http://localhost:8000'

    const { host } = window.location

    if (host === 'localhost:3001') return 'http://localhost:8000'

    // Transform domain: pinacolada.co → api.pinacolada.co
    const parts = host.split('.')
    if (parts.length === 2) {
        parts.unshift('api')
        return `https://${parts.join('.')}`
    }

    return `https://api.${host}`
}
```

### WebSocket URLs
```typescript
// lib/get-ws-url.ts
export const getWsUrl = () => {
    if (typeof window === 'undefined') return 'ws://localhost:8000/ws'

    const { host } = window.location

    if (host === 'localhost:3001') return 'ws://localhost:8000/ws'

    return `wss://api.${host}/ws`
}
```

### Feature Flags
```typescript
// lib/get-features.ts
export const getFeatures = () => {
    if (typeof window === 'undefined') {
        return { debugMode: false, analytics: false }
    }

    const { host } = window.location

    if (host === 'localhost:3001') {
        return { debugMode: true, analytics: false }
    }

    return { debugMode: false, analytics: true }
}
```

---

## Migration from Previous Pattern

### Before (Insecure)
```typescript
// ❌ Built into bundle
const password = process.env.NEXT_PUBLIC_JOBS_PASSWORD
```

### After (Secure)
```typescript
// ✅ Runtime detection + server validation
const { isDev } = getAuthConfig()
if (isDev) { /* client check */ }
else { /* API call */ }
```

---

## Related Patterns in Codebase

This pattern is consistent with:

1. **discogs-player-2024** - Base URL detection
2. **Existing Next.js best practices** - API routes for sensitive operations
3. **TypeScript patterns** - Type-safe configuration objects
4. **Functional programming** - Pure utility functions with guard clauses

---

**Summary**: The authentication pattern now perfectly matches the established pattern from `discogs-player-2024`, providing runtime environment detection, secure server-side validation, and excellent developer experience.
