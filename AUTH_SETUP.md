# Auth0 Integration Setup Complete

## Implementation Summary

Auth0 multi-tenant authentication has been implemented following the specification in `auth_spec.md`.

## What Was Implemented

### Database (modules/agent/migrations/)

- ✅ `003_auth0_integration.sql` - Adds `auth0_sub` to User table, creates TenantInvitation table

### Backend (modules/agent/src/)

- ✅ `lib/auth.py` - JWT verification, `@require_auth` decorator
- ✅ `services/auth_service.py` - User/tenant management logic
- ✅ `repositories/user_repository.py` - User data access with Auth0 sub lookup
- ✅ `api/routes/auth.py` - Auth endpoints (/api/auth/me, /tenant/create, /tenant/switch)
- ✅ `models/User.py` - Updated with `auth0_sub` field
- ✅ `server.py` - Registered auth routes
- ✅ `pyproject.toml` - Added python-jose dependencies

### Frontend (modules/client/)

- ✅ `lib/auth0.ts` - Auth0 client configuration
- ✅ `lib/fetch-bearer-token.ts` - Token fetching utility
- ✅ `middleware.ts` - Route protection
- ✅ `app/api/auth/[auth0]/route.ts` - Auth0 callback handler
- ✅ `app/api/auth/token/route.ts` - Access token endpoint
- ✅ `app/layout.tsx` - Wrapped with UserProvider
- ✅ `app/login/page.tsx` - Login redirect page
- ✅ `app/tenant/select/page.tsx` - Tenant selection UI
- ✅ `components/auth/TenantSelector.tsx` - Tenant selector component
- ✅ `package.json` - Added @auth0/nextjs-auth0 and axios

## Next Steps

### 1. Run Database Migration

```bash
cd modules/agent
python -m agent.scripts.run_migrations
```

### 2. Install Dependencies

**Backend:**

```bash
cd modules/agent
pip install -r requirements.txt
# or
pip install python-jose[cryptography] python-multipart
```

**Frontend:**

```bash
cd modules/client
npm install
```

### 3. Configure Auth0

#### Create Auth0 Application

1. Go to Auth0 Dashboard → Applications → Create Application
2. Select "Regular Web Application"
3. Configure Application URIs:
   - **Allowed Callback URLs**: `http://localhost:3001/api/auth/callback`
   - **Allowed Logout URLs**: `http://localhost:3001`
   - **Allowed Web Origins**: `http://localhost:3001`
4. Enable Refresh Token Rotation
5. Note: Domain, Client ID, Client Secret

#### Create Auth0 API

1. Go to Auth0 Dashboard → APIs → Create API
2. Set **Identifier**: `https://api.pinacolada.co`
3. Enable RBAC
4. Enable "Add Permissions in Access Token"

Problem:
By default, Auth0 access tokens only contain standard JWT claims (sub,
aud, iat, etc.). Email is in the ID token, but your backend validates the
access token.

Solution: Add Auth0 Action to include email in access token

Steps:

1. Go to Auth0 Dashboard → Actions → Flows
2. Select "Login" flow
3. Click "+" (Custom) → Build from scratch
4. Name it: "Add email to access token"
5. Add this code:

exports.onExecutePostLogin = async (event, api) => {
if (event.authorization) {
// Add email to access token
api.accessToken.setCustomClaim('email', event.user.email);

      // Optional: Add other useful claims
      api.accessToken.setCustomClaim('name', event.user.name);
      api.accessToken.setCustomClaim('picture', event.user.picture);
    }

};

6. Click "Deploy"
7. Drag the action into your Login flow (between "Start" and "Complete")
8. Click "Apply"

### 4. Environment Variables

**Backend (.env):**

```env
AUTH0_DOMAIN=your-tenant.auth0.com
AUTH0_AUDIENCE=https://api.pinacolada.co
```

**Frontend (.env.local):**

```env
AUTH0_DOMAIN=your-tenant.auth0.com
AUTH0_CLIENT_ID=<client-id>
AUTH0_CLIENT_SECRET=<client-secret>
AUTH0_SECRET=<random-32-char-string>
AUTH0_AUDIENCE=https://api.pinacolada.co
AUTH0_SCOPE=openid profile email
APP_BASE_URL=http://localhost:3001
NEXT_PUBLIC_API_URL=http://localhost:8000
```

Generate AUTH0_SECRET:

```bash
openssl rand -hex 32
```

### 5. Test Authentication Flow

1. Start backend: `cd modules/agent && uvicorn src.server:app --reload --port 8000`
2. Start frontend: `cd modules/client && npm run dev`
3. Visit `http://localhost:3001`
4. Should redirect to login
5. After Auth0 login, should land on tenant selection
6. Create or select tenant
7. Should be redirected to home page

### 6. Protect Routes

To protect backend routes, add the `@require_auth` decorator:

```python
from lib.auth import require_auth

@router.get("/api/jobs")
@require_auth
async def list_jobs(request: Request):
    tenant_id = request.state.tenant_id
    # ... your logic
```

Example already applied in `api/routes/auth.py`.

### 7. Frontend API Calls

Use `fetchBearerToken` for authenticated requests:

```typescript
import { fetchBearerToken } from "../lib/fetch-bearer-token";
import axios from "axios";

const headers = await fetchBearerToken();
const response = await axios.get(
  `${process.env.NEXT_PUBLIC_API_URL}/api/jobs`,
  headers
);
```

## Outstanding Tasks

- [ ] Apply `@require_auth` to existing job/lead routes
- [ ] Add tenant context to all repository queries
- [ ] Implement invitation system (TenantInvitation CRUD)
- [ ] Add tenant switcher dropdown to Header component
- [ ] Write tests for auth flow
- [ ] Update Azure Pipeline with Auth0 environment variables
- [ ] Configure production Auth0 callback URLs

## Architecture Notes

### Multi-Tenancy Strategy

- All queries should filter by `tenant_id` from `request.state.tenant_id`
- User belongs to one primary tenant (`User.tenant_id`)
- User can be member of multiple tenants via `UserRole` relationships
- Active tenant stored in cookie and `X-Tenant-Id` header

### Security

- JWT verified on every request
- Tokens short-lived (24h access, refresh token rotation)
- Middleware protects all routes except /login, /about, /api/auth
- User/tenant context automatically injected by `@require_auth`

## Troubleshooting

**"Invalid token" errors:**

- Check AUTH0_DOMAIN and AUTH0_AUDIENCE match in backend/frontend
- Verify Auth0 API identifier matches AUTH0_AUDIENCE

**"Missing Authorization header":**

- Ensure `fetchBearerToken()` is used for API calls
- Check token endpoint returns accessToken

**Tenant not set:**

- User must complete tenant selection flow after first login
- Check `/tenant/select` page loads correctly

## Reference Files

- Spec: `auth_spec.md`
- Migration: `modules/agent/migrations/003_auth0_integration.sql`
- Backend Auth: `modules/agent/src/lib/auth.py`
- Frontend Config: `modules/client/lib/auth0.ts`
