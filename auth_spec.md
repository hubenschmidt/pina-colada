# Authentication Specification

Multi-tenant authentication system for PinaColada.co using Auth0 as the identity provider.

## Architecture Overview

### Components
- **Identity Provider**: Auth0 (familiar pattern from discogs-player-2024)
- **Frontend**: Next.js 15 with `@auth0/nextjs-auth0` v4.x SDK
- **Backend**: Python FastAPI with JWT verification using `python-jose`
- **Database**: PostgreSQL with multi-tenant schema (Tenant, User, Role, UserRole)

### Authentication Flow
1. User signs up/logs in via Auth0 hosted UI
2. Auth0 redirects back to Next.js with authorization code
3. Next.js exchanges code for tokens (ID token, access token, refresh token)
4. First-time users: Create User record in database, prompt for tenant selection
5. Returning users: Load tenant association(s)
6. Access token sent to FastAPI backend in Authorization header
7. Backend verifies JWT signature and extracts claims (sub, email)
8. Backend maps Auth0 `sub` to User record and loads tenant_id
9. All API operations scoped to tenant_id

## Frontend Implementation (Next.js)

### Dependencies
```json
{
  "@auth0/nextjs-auth0": "^4.6.0"
}
```

### File Structure
```
modules/client/
├── lib/
│   └── auth0.ts              # Auth0 client configuration
├── middleware.ts             # Route protection
├── app/
│   ├── layout.tsx            # Wrap with UserProvider
│   ├── api/
│   │   └── auth/
│   │       └── [auth0]/
│   │           └── route.ts  # Auth0 callback handler
│   ├── login/
│   │   └── page.tsx          # Login page (redirects to Auth0)
│   └── tenant/
│       └── select/
│           └── page.tsx      # Tenant selection UI
└── components/
    ├── TenantSelector.tsx    # Component for tenant selection
    └── ProtectedRoute.tsx    # Client-side route guard
```

### Configuration (`lib/auth0.ts`)
```typescript
import { Auth0Client } from '@auth0/nextjs-auth0/server';

export const auth0 = new Auth0Client({
  domain: process.env.AUTH0_DOMAIN,
  clientId: process.env.AUTH0_CLIENT_ID,
  clientSecret: process.env.AUTH0_CLIENT_SECRET,
  appBaseUrl: process.env.APP_BASE_URL,
  secret: process.env.AUTH0_SECRET,
  authorizationParameters: {
    scope: process.env.AUTH0_SCOPE,      // 'openid profile email'
    audience: process.env.AUTH0_AUDIENCE, // Backend API identifier
  },
});
```

### Middleware (`middleware.ts`)
Protects all routes except public pages. Pattern from discogs-player-2024:
```typescript
import { auth0 } from './lib/auth0';
import type { NextRequest } from 'next/server';

export async function middleware(request: NextRequest) {
  return await auth0.middleware(request);
}

export const config = {
  matcher: [
    '/((?!_next/static|_next/image|favicon.ico|login|api/auth).*)',
  ],
};
```

### Access Token Retrieval
Pattern from discogs-player for backend API calls:
```typescript
// lib/fetch-bearer-token.ts
import axios from 'axios';

export const fetchBearerToken = (): Promise<{ headers: { Authorization: string } }> =>
  axios.get('/api/auth/token')
    .then(res => ({
      headers: { Authorization: `Bearer ${res.data.accessToken}` }
    }));
```

### Root Layout Updates (`app/layout.tsx`)
```typescript
import { UserProvider } from '@auth0/nextjs-auth0/client';

export default function RootLayout({ children }) {
  return (
    <html>
      <body>
        <UserProvider>
          <MantineProvider>
            <TenantProvider>  {/* Custom provider for tenant context */}
              <Header />
              {children}
            </TenantProvider>
          </MantineProvider>
        </UserProvider>
      </body>
    </html>
  );
}
```

### Tenant Selection Flow
After successful Auth0 authentication:

1. **First Login**:
   - Frontend calls `/api/auth/me` to check if User record exists
   - If not: Create User record with Auth0 `sub` mapping
   - Redirect to `/tenant/select` page
   - Options:
     - Create new tenant (user becomes owner)
     - Enter invitation code to join existing tenant

2. **Subsequent Logins**:
   - Load user's tenant associations from `/api/auth/me`
   - If multiple tenants: Show tenant selector
   - If single tenant: Auto-select
   - Store selected tenant_id in encrypted cookie

3. **Tenant Switching**:
   - Dropdown in header to switch active tenant
   - Updates cookie, refreshes UI

### Environment Variables
```env
AUTH0_DOMAIN=your-tenant.auth0.com
AUTH0_CLIENT_ID=<client-id>
AUTH0_CLIENT_SECRET=<client-secret>
AUTH0_SECRET=<random-32-char-string>
AUTH0_AUDIENCE=https://api.pinacolada.co
AUTH0_SCOPE=openid profile email
APP_BASE_URL=http://localhost:3001
```

## Backend Implementation (Python FastAPI)

### Dependencies
Add to `pyproject.toml`:
```toml
dependencies = [
  # ... existing deps
  "python-jose[cryptography]>=3.3.0",
  "python-multipart>=0.0.6",
]
```

### File Structure
```
modules/agent/src/
├── lib/
│   └── auth.py               # JWT verification, decorators
├── api/
│   └── routes/
│       ├── auth.py           # Auth endpoints (/api/auth/me, /api/auth/token)
│       └── jobs.py           # Protected job routes (updated)
├── repositories/
│   └── user_repository.py    # User CRUD with Auth0 sub mapping
├── services/
│   └── auth_service.py       # Business logic for user/tenant management
└── middleware/
    └── tenant.py             # Tenant context injection
```

### JWT Verification (`lib/auth.py`)
Pattern similar to `express-oauth2-jwt-bearer` from discogs-player Express backend:

```python
from functools import wraps
from typing import Optional
from fastapi import HTTPException, Request
from fastapi.security import HTTPBearer, HTTPAuthorizationCredentials
from jose import jwt, JWTError
import requests

security = HTTPBearer()

# Cache JWKS (JSON Web Key Set) from Auth0
def get_jwks():
    """Fetch Auth0 public keys for JWT verification."""
    jwks_url = f"https://{os.getenv('AUTH0_DOMAIN')}/.well-known/jwks.json"
    response = requests.get(jwks_url)
    return response.json()

def verify_token(token: str) -> dict:
    """Verify JWT token and return claims."""
    try:
        # Get signing key from JWKS
        unverified_header = jwt.get_unverified_header(token)
        jwks = get_jwks()
        rsa_key = {}
        for key in jwks["keys"]:
            if key["kid"] == unverified_header["kid"]:
                rsa_key = {
                    "kty": key["kty"],
                    "kid": key["kid"],
                    "use": key["use"],
                    "n": key["n"],
                    "e": key["e"]
                }

        # Verify token
        payload = jwt.decode(
            token,
            rsa_key,
            algorithms=["RS256"],
            audience=os.getenv("AUTH0_AUDIENCE"),
            issuer=f"https://{os.getenv('AUTH0_DOMAIN')}/"
        )
        return payload
    except JWTError as e:
        raise HTTPException(status_code=401, detail=f"Invalid token: {str(e)}")

def require_auth(func):
    """Decorator to protect routes with JWT authentication."""
    @wraps(func)
    async def wrapper(request: Request, *args, **kwargs):
        auth_header = request.headers.get("Authorization")
        if not auth_header or not auth_header.startswith("Bearer "):
            raise HTTPException(status_code=401, detail="Missing or invalid Authorization header")

        token = auth_header.split(" ")[1]
        claims = verify_token(token)

        # Attach user info to request state
        request.state.auth0_sub = claims.get("sub")
        request.state.email = claims.get("email")

        # Load user from database
        from services.auth_service import get_or_create_user
        user = await get_or_create_user(claims.get("sub"), claims.get("email"))
        request.state.user = user
        request.state.user_id = user.id

        # Get tenant from header or session
        tenant_id = request.headers.get("X-Tenant-Id") or request.cookies.get("tenant_id")
        if tenant_id:
            request.state.tenant_id = int(tenant_id)
        else:
            # Default to user's first tenant if not specified
            if user.tenant_id:
                request.state.tenant_id = user.tenant_id

        return await func(request, *args, **kwargs)
    return wrapper
```

### Auth Service (`services/auth_service.py`)
```python
from repositories.user_repository import find_user_by_auth0_sub, create_user
from repositories.tenant_repository import get_user_tenants
from models.User import UserCreateData

async def get_or_create_user(auth0_sub: str, email: str):
    """Get or create user from Auth0 sub."""
    user = find_user_by_auth0_sub(auth0_sub)
    if not user:
        # First login - create user without tenant
        data: UserCreateData = {
            "auth0_sub": auth0_sub,
            "email": email,
            "status": "active"
        }
        user = create_user(data)
    return user

def get_user_tenants(user_id: int):
    """Get all tenants user belongs to."""
    # Query User → UserRole → Role relationships
    # Return list of tenants with role info
    pass

def create_tenant_for_user(user_id: int, tenant_name: str):
    """Create new tenant and assign user as owner."""
    # Create Tenant
    # Create owner Role assignment
    # Update User.tenant_id
    pass

def add_user_to_tenant(user_id: int, tenant_id: int, role_name: str):
    """Add user to existing tenant with role."""
    # Create UserRole record
    pass
```

### User Repository Updates (`repositories/user_repository.py`)
Add `auth0_sub` mapping:
```python
def find_user_by_auth0_sub(auth0_sub: str) -> Optional[User]:
    """Find user by Auth0 subject ID."""
    session = get_session()
    try:
        stmt = select(User).where(User.auth0_sub == auth0_sub)
        return session.execute(stmt).scalar_one_or_none()
    finally:
        session.close()
```

### Auth Routes (`api/routes/auth.py`)
```python
from fastapi import APIRouter, Request, HTTPException
from lib.auth import require_auth

router = APIRouter(prefix="/api/auth", tags=["auth"])

@router.get("/me")
@require_auth
async def get_current_user(request: Request):
    """Get current user profile with tenant associations."""
    user = request.state.user
    tenants = get_user_tenants(user.id)
    return {
        "user": user_to_dict(user),
        "tenants": tenants,
        "current_tenant_id": request.state.tenant_id
    }

@router.post("/tenant/create")
@require_auth
async def create_tenant(request: Request, tenant_name: str):
    """Create new tenant and assign current user as owner."""
    user_id = request.state.user_id
    tenant = create_tenant_for_user(user_id, tenant_name)
    return {"tenant": tenant_to_dict(tenant)}

@router.post("/tenant/switch")
@require_auth
async def switch_tenant(request: Request, tenant_id: int):
    """Switch active tenant for current user."""
    # Verify user has access to tenant
    # Set cookie with new tenant_id
    pass
```

### Protected Route Example (`api/routes/jobs.py`)
```python
from lib.auth import require_auth

@router.get("")
@require_auth
async def list_jobs(request: Request):
    """List jobs for current tenant."""
    tenant_id = request.state.tenant_id
    if not tenant_id:
        raise HTTPException(status_code=400, detail="No tenant selected")

    jobs = find_all_jobs(tenant_id=tenant_id)
    return {"jobs": [job_to_dict(j) for j in jobs]}
```

### Environment Variables
```env
AUTH0_DOMAIN=your-tenant.auth0.com
AUTH0_AUDIENCE=https://api.pinacolada.co
```

## Database Schema Updates

### User Table
Add `auth0_sub` column to link Auth0 identity:
```sql
ALTER TABLE "User" ADD COLUMN IF NOT EXISTS auth0_sub TEXT UNIQUE;
CREATE INDEX IF NOT EXISTS idx_user_auth0_sub ON "User"(auth0_sub);
```

### Migration
Create `003_auth0_integration.sql`:
```sql
-- Add Auth0 subject mapping to User table
ALTER TABLE "User" ADD COLUMN IF NOT EXISTS auth0_sub TEXT UNIQUE;
CREATE INDEX IF NOT EXISTS idx_user_auth0_sub ON "User"(auth0_sub);

-- Add invitation system for tenant onboarding
CREATE TABLE IF NOT EXISTS "TenantInvitation" (
  id              BIGSERIAL PRIMARY KEY,
  tenant_id       BIGINT NOT NULL REFERENCES "Tenant"(id) ON DELETE CASCADE,
  email           TEXT NOT NULL,
  role_id         BIGINT NOT NULL REFERENCES "Role"(id),
  token           TEXT NOT NULL UNIQUE,
  expires_at      TIMESTAMPTZ NOT NULL,
  accepted_at     TIMESTAMPTZ,
  invited_by      BIGINT REFERENCES "User"(id),
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_invitation_token ON "TenantInvitation"(token);
CREATE INDEX IF NOT EXISTS idx_invitation_email ON "TenantInvitation"(email);
```

## Auth0 Configuration

### Application Setup
1. Create **Regular Web Application** in Auth0 dashboard
2. Configure Application URIs:
   - **Allowed Callback URLs**: `http://localhost:3001/api/auth/callback`
   - **Allowed Logout URLs**: `http://localhost:3001`
   - **Allowed Web Origins**: `http://localhost:3001`
3. Enable **Refresh Token Rotation**
4. Set token expiration: Access Token = 24h, ID Token = 10h

### API Setup (for Backend)
1. Create **API** in Auth0 dashboard
2. Set **Identifier**: `https://api.pinacolada.co` (matches AUTH0_AUDIENCE)
3. Enable **RBAC** (Role-Based Access Control)
4. Enable **Add Permissions in Access Token**

### Custom Claims with Actions
Create Auth0 Action to add tenant metadata to token:

**Action Name**: `add-tenant-claims`
**Trigger**: Login / Post Login
**Code**:
```javascript
exports.onExecutePostLogin = async (event, api) => {
  const namespace = 'https://pinacolada.co';

  // Get user's tenant from app_metadata (set after tenant selection)
  const tenantId = event.user.app_metadata?.tenant_id;
  const tenants = event.user.app_metadata?.tenants || [];

  if (tenantId) {
    api.accessToken.setCustomClaim(`${namespace}/tenant_id`, tenantId);
  }

  api.accessToken.setCustomClaim(`${namespace}/tenants`, tenants);
  api.idToken.setCustomClaim(`${namespace}/tenant_id`, tenantId);
};
```

### Roles
Create standard roles in Auth0 (optional - can also manage in app):
- `owner` - Full access, billing
- `admin` - Manage users, settings
- `member` - Standard access
- `viewer` - Read-only

## Multi-Tenancy Patterns

### Tenant Isolation Strategies

**Option 1: Application-Level (Current)**
- All queries include `WHERE tenant_id = ?`
- Enforced in repository layer
- Fast, flexible
- Risk: Developer error can leak data

**Option 2: Row Level Security (Future)**
```sql
ALTER TABLE "Deal" ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_policy ON "Deal"
  USING (tenant_id = current_setting('app.current_tenant_id')::BIGINT);
```
Set tenant context per request:
```python
@app.middleware("http")
async def set_tenant_context(request: Request, call_next):
    if hasattr(request.state, "tenant_id"):
        # Set Postgres session variable
        session = get_session()
        session.execute(text(f"SET app.current_tenant_id = {request.state.tenant_id}"))
    return await call_next(request)
```

### Tenant Onboarding Flows

**Individual User Sign-up**:
1. User signs up via Auth0 (creates account)
2. First login → lands on `/tenant/select`
3. Options:
   - "Create New Organization" → Creates Tenant, user becomes owner
   - "Join Existing Organization" → Enter invitation code
4. Selected tenant stored in session cookie
5. Update Auth0 `app_metadata` with tenant_id

**Invitation Flow**:
1. Existing user invites new user by email
2. System creates `TenantInvitation` record with unique token
3. Email sent with link: `https://pinacolada.co/invite/{token}`
4. New user clicks link → prompted to sign up/log in via Auth0
5. After Auth0 callback, token validated and user added to tenant
6. User assigned specified role

**Multiple Tenants**:
- Users can belong to multiple tenants
- Header dropdown to switch active tenant
- Updates `X-Tenant-Id` header and cookie
- All queries scoped to active tenant

## Security Considerations

### Token Security
- Access tokens short-lived (24h)
- Refresh tokens rotated on use
- Tokens stored in httpOnly cookies (not localStorage)
- CSRF protection via SameSite cookies

### API Security
- All routes protected with `@require_auth` decorator
- JWT signature verified on every request
- Claims validated (audience, issuer, expiration)
- Rate limiting per user/tenant

### Tenant Isolation
- All DB queries filtered by tenant_id
- Repository layer enforces tenant scope
- Consider RLS for additional safety layer
- Audit logs for cross-tenant access attempts

### Environment Security
- Secrets in environment variables
- Never commit `.env.local` files
- Use Azure Key Vault in production
- Rotate AUTH0_SECRET regularly

## Testing Strategy

### Unit Tests
- JWT verification with valid/invalid tokens
- Tenant scoping in repositories
- User creation/lookup by Auth0 sub

### Integration Tests
- Full auth flow with Auth0 test tenant
- Tenant switching
- Invitation acceptance
- Multi-tenant data isolation

### Manual Testing Checklist
- [ ] Sign up new user → creates User record
- [ ] First login → tenant selection shown
- [ ] Create tenant → user becomes owner
- [ ] Invite user → receives email, can join
- [ ] Switch tenant → queries scoped correctly
- [ ] Logout → tokens cleared
- [ ] Expired token → 401 response
- [ ] Cross-tenant access → blocked

## Comparison: Auth0 vs Alternatives

### Auth0 (Recommended)
**Pros**:
- Familiar pattern from discogs-player-2024
- Hosted UI (no login form maintenance)
- Enterprise features (SSO, MFA) available
- Excellent documentation
- Custom claims via Actions
- Built-in user management dashboard

**Cons**:
- Cost scales with MAU (Monthly Active Users)
- Vendor lock-in
- Requires Auth0 Actions for custom logic

### Supabase Auth (Alternative)
**Pros**:
- Already using Supabase for database
- Free tier generous
- Built-in RLS integration
- Open source

**Cons**:
- Less mature than Auth0
- Need to build custom UI
- Already committed to Postgres (not Supabase DB)

### NextAuth.js (Alternative)
**Pros**:
- Open source, self-hosted
- Free
- Deep Next.js integration

**Cons**:
- More implementation work
- Manage security yourself
- No hosted UI
- Less enterprise features

### Recommendation
**Stick with Auth0** given:
1. Existing familiarity from discogs-player-2024
2. Proven pattern with Next.js + Express (adaptable to FastAPI)
3. Time savings vs building custom auth
4. Enterprise-ready for future growth
5. Strong security guarantees

## Implementation Checklist

### Phase 1: Auth0 Setup
- [ ] Create Auth0 account/tenant
- [ ] Create Regular Web Application
- [ ] Create API (backend audience)
- [ ] Configure callback URLs
- [ ] Create Auth0 Action for custom claims
- [ ] Test login flow in Auth0 dashboard

### Phase 2: Frontend
- [ ] Install `@auth0/nextjs-auth0`
- [ ] Create `lib/auth0.ts`
- [ ] Add `middleware.ts`
- [ ] Wrap layout with `UserProvider`
- [ ] Create `/api/auth/[auth0]/route.ts`
- [ ] Create tenant selection UI
- [ ] Test login flow

### Phase 3: Backend
- [ ] Add `python-jose` to dependencies
- [ ] Create `lib/auth.py` with JWT verification
- [ ] Create `services/auth_service.py`
- [ ] Add `@require_auth` decorator
- [ ] Create `/api/auth/me` endpoint
- [ ] Update job routes to require auth
- [ ] Test with Postman/curl

### Phase 4: Database
- [ ] Add `auth0_sub` column to User table
- [ ] Create migration `003_auth0_integration.sql`
- [ ] Create `TenantInvitation` table
- [ ] Add indexes
- [ ] Run migration

### Phase 5: Multi-Tenancy
- [ ] Implement tenant selection flow
- [ ] Create tenant switching logic
- [ ] Add tenant context to all queries
- [ ] Create invitation system
- [ ] Test data isolation

### Phase 6: Testing & Deployment
- [ ] Write unit tests
- [ ] Write integration tests
- [ ] Manual testing checklist
- [ ] Update Azure Pipeline with Auth0 secrets
- [ ] Deploy to production
- [ ] Monitor logs for auth errors

## References

### Documentation
- [Auth0 Next.js SDK](https://auth0.com/docs/quickstart/webapp/nextjs)
- [Auth0 API Authorization](https://auth0.com/docs/secure/api-authorization)
- [python-jose Documentation](https://python-jose.readthedocs.io/)
- [FastAPI Security](https://fastapi.tiangolo.com/tutorial/security/)

### Existing Implementation
- `discogs-player-2024/modules/client/lib/auth0.js` - Auth0 client setup
- `discogs-player-2024/modules/client/middleware.ts` - Route protection
- `discogs-player-2024/modules/client/pages/api/get-token.ts` - Token fetching
- `discogs-player-2024/modules/app/index.ts` - Express JWT middleware pattern

### Internal References
- `modules/agent/migrations/002_multitenancy.sql` - Tenant schema
- `modules/agent/src/models/User.py` - User model
- `modules/agent/src/models/Tenant.py` - Tenant model
