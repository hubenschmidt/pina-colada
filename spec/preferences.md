# User & Tenant Preferences - Technical Specification

## Context

Implement theming system for authenticated users with support for both user-level and tenant-level preferences. Theme toggle only available when `isAuthed === true`.

**Requirements:**
- Light theme (default, current CSS)
- Dark theme (lime green borders: `#00FF00`)
- User-level preferences (individual theme choice)
- Tenant-level preferences (shared across tenant, Admin/SuperAdmin only)
- Preference hierarchy: User → Tenant → System default (light)

---

## Database Schema

### Migration 005: Preferences Tables

#### Changes to Existing Tables

**Remove:**
- `Tenant.settings` (JSONB column) - to be replaced with dedicated table

#### New Tables

**TenantPreferences**
```python
class TenantPreferences(Base):
    __tablename__ = "TenantPreferences"

    id = Column(BigInteger, primary_key=True, autoincrement=True)
    tenant_id = Column(BigInteger, ForeignKey("Tenant.id", ondelete="CASCADE"),
                       nullable=False, unique=True, index=True)
    theme = Column(Text, nullable=False, default="light")  # "light" | "dark"
    created_at = Column(DateTime, nullable=False, server_default=func.now())
    updated_at = Column(DateTime, nullable=False,
                       server_default=func.now(), onupdate=func.now())

    # Relationships
    tenant = relationship("Tenant", back_populates="preferences", uselist=False)
```

**UserPreferences**
```python
class UserPreferences(Base):
    __tablename__ = "UserPreferences"

    id = Column(BigInteger, primary_key=True, autoincrement=True)
    user_id = Column(BigInteger, ForeignKey("User.id", ondelete="CASCADE"),
                     nullable=False, unique=True, index=True)
    theme = Column(Text, nullable=True)  # "light" | "dark" | null (inherit tenant)
    created_at = Column(DateTime, nullable=False, server_default=func.now())
    updated_at = Column(DateTime, nullable=False,
                       server_default=func.now(), onupdate=func.now())

    # Relationships
    user = relationship("User", back_populates="preferences", uselist=False)
```

#### Relationship Updates

**Tenant Model:**
```python
# Add to Tenant class
preferences = relationship("TenantPreferences", back_populates="tenant",
                          uselist=False, cascade="all, delete-orphan")
```

**User Model:**
```python
# Add to User class
preferences = relationship("UserPreferences", back_populates="user",
                          uselist=False, cascade="all, delete-orphan")
```

#### Migration Script: `005_create_preferences_tables.py`

```python
"""
Create TenantPreferences and UserPreferences tables, migrate Tenant.settings

Revision ID: 005
Revises: 004
"""

async def upgrade():
    # 1. Create TenantPreferences table
    op.create_table(
        'TenantPreferences',
        sa.Column('id', sa.BigInteger(), autoincrement=True, primary_key=True),
        sa.Column('tenant_id', sa.BigInteger(), nullable=False),
        sa.Column('theme', sa.Text(), nullable=False, server_default='light'),
        sa.Column('created_at', sa.DateTime(), server_default=sa.func.now()),
        sa.Column('updated_at', sa.DateTime(), server_default=sa.func.now(),
                  onupdate=sa.func.now()),
        sa.ForeignKeyConstraint(['tenant_id'], ['Tenant.id'], ondelete='CASCADE'),
        sa.UniqueConstraint('tenant_id')
    )
    op.create_index('ix_TenantPreferences_tenant_id', 'TenantPreferences', ['tenant_id'])

    # 2. Create UserPreferences table
    op.create_table(
        'UserPreferences',
        sa.Column('id', sa.BigInteger(), autoincrement=True, primary_key=True),
        sa.Column('user_id', sa.BigInteger(), nullable=False),
        sa.Column('theme', sa.Text(), nullable=True),
        sa.Column('created_at', sa.DateTime(), server_default=sa.func.now()),
        sa.Column('updated_at', sa.DateTime(), server_default=sa.func.now(),
                  onupdate=sa.func.now()),
        sa.ForeignKeyConstraint(['user_id'], ['User.id'], ondelete='CASCADE'),
        sa.UniqueConstraint('user_id')
    )
    op.create_index('ix_UserPreferences_user_id', 'UserPreferences', ['user_id'])

    # 3. Migrate existing Tenant.settings data (if any themes were stored)
    # Note: Current schema has settings={} by default, so likely no data to migrate
    connection = op.get_bind()
    # If Tenant.settings.theme existed, migrate here

    # 4. Create default TenantPreferences for all existing tenants
    connection.execute(
        """
        INSERT INTO "TenantPreferences" (tenant_id, theme, created_at, updated_at)
        SELECT id, 'light', NOW(), NOW()
        FROM "Tenant"
        """
    )

    # 5. Create default UserPreferences for all existing users (theme=null to inherit)
    connection.execute(
        """
        INSERT INTO "UserPreferences" (user_id, theme, created_at, updated_at)
        SELECT id, NULL, NOW(), NOW()
        FROM "User"
        """
    )

    # 6. Drop Tenant.settings column
    op.drop_column('Tenant', 'settings')

async def downgrade():
    # 1. Re-add Tenant.settings column
    op.add_column('Tenant',
        sa.Column('settings', JSONB, nullable=False, server_default='{}'))

    # 2. Drop tables
    op.drop_index('ix_UserPreferences_user_id')
    op.drop_table('UserPreferences')

    op.drop_index('ix_TenantPreferences_tenant_id')
    op.drop_table('TenantPreferences')
```

---

## Theme System

### Theme Definitions

**Light Theme (Default)**
- Current CSS as-is
- Background: `#ffffff`
- Foreground: `#171717`
- Borders: standard grays

**Dark Theme**
- Background: `#0a0a0a` (near-black)
- Foreground: `#e5e5e5` (light gray text)
- Borders: `1px solid #00FF00` (lime green)
- Accent: lime green (`#00FF00`)

### Preference Hierarchy

```
User opens app
  ↓
Check UserPreferences.theme
  ├─ If set → use user's theme
  └─ If null → check TenantPreferences.theme
       ├─ If set → use tenant's theme
       └─ Else → use system default ("light")
```

**Resolution Logic:**
```python
def resolve_theme(user: User) -> str:
    """Resolve effective theme for user"""
    if user.preferences and user.preferences.theme:
        return user.preferences.theme

    if user.tenant.preferences and user.tenant.preferences.theme:
        return user.tenant.preferences.theme

    return "light"  # System default
```

---

## Backend API

### Endpoints

**1. Get User Preferences**
```
GET /api/preferences/user
Authorization: Bearer <token>

Response:
{
  "theme": "dark" | "light" | null,
  "effective_theme": "dark" | "light",  // Resolved theme after hierarchy
  "can_edit_tenant": boolean  // true if Admin/SuperAdmin
}
```

**2. Update User Preferences**
```
PATCH /api/preferences/user
Authorization: Bearer <token>
Body: {
  "theme": "dark" | "light" | null  // null = inherit from tenant
}

Response:
{
  "theme": "dark",
  "effective_theme": "dark"
}
```

**3. Get Tenant Preferences** (Admin/SuperAdmin only)
```
GET /api/preferences/tenant
Authorization: Bearer <token>

Response:
{
  "theme": "light"
}

Error (403): User does not have permission to view tenant preferences
```

**4. Update Tenant Preferences** (Admin/SuperAdmin only)
```
PATCH /api/preferences/tenant
Authorization: Bearer <token>
Body: {
  "theme": "dark" | "light"
}

Response:
{
  "theme": "dark"
}

Error (403): User does not have permission to modify tenant preferences
```

### Implementation

**File:** `modules/agent/src/api/routes/preferences.py`

```python
from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.ext.asyncio import AsyncSession
from models.User import User
from models.UserPreferences import UserPreferences
from models.TenantPreferences import TenantPreferences
from lib.auth import get_current_user
from lib.db import get_db

router = APIRouter(prefix="/api/preferences")

@router.get("/user")
async def get_user_preferences(
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """Get current user's preferences with resolved theme"""
    # Ensure preferences exist
    if not user.preferences:
        user_prefs = UserPreferences(user_id=user.id, theme=None)
        db.add(user_prefs)
        await db.commit()
        await db.refresh(user)

    effective_theme = resolve_theme(user)
    can_edit_tenant = user_has_admin_role(user)

    return {
        "theme": user.preferences.theme,
        "effective_theme": effective_theme,
        "can_edit_tenant": can_edit_tenant
    }

@router.patch("/user")
async def update_user_preferences(
    body: dict,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """Update current user's theme preference"""
    theme = body.get("theme")
    if theme not in ["light", "dark", None]:
        raise HTTPException(400, "Invalid theme value")

    if not user.preferences:
        user.preferences = UserPreferences(user_id=user.id)
        db.add(user.preferences)

    user.preferences.theme = theme
    await db.commit()
    await db.refresh(user)

    return {
        "theme": user.preferences.theme,
        "effective_theme": resolve_theme(user)
    }

@router.get("/tenant")
async def get_tenant_preferences(
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """Get tenant preferences (Admin/SuperAdmin only)"""
    if not user_has_admin_role(user):
        raise HTTPException(403, "Insufficient permissions")

    # Ensure tenant preferences exist
    if not user.tenant.preferences:
        tenant_prefs = TenantPreferences(tenant_id=user.tenant_id, theme="light")
        db.add(tenant_prefs)
        await db.commit()
        await db.refresh(user.tenant)

    return {"theme": user.tenant.preferences.theme}

@router.patch("/tenant")
async def update_tenant_preferences(
    body: dict,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """Update tenant theme preference (Admin/SuperAdmin only)"""
    if not user_has_admin_role(user):
        raise HTTPException(403, "Insufficient permissions")

    theme = body.get("theme")
    if theme not in ["light", "dark"]:
        raise HTTPException(400, "Invalid theme value")

    if not user.tenant.preferences:
        user.tenant.preferences = TenantPreferences(tenant_id=user.tenant_id)
        db.add(user.tenant.preferences)

    user.tenant.preferences.theme = theme
    await db.commit()

    return {"theme": user.tenant.preferences.theme}

def resolve_theme(user: User) -> str:
    """Resolve effective theme using hierarchy"""
    if user.preferences and user.preferences.theme:
        return user.preferences.theme
    if user.tenant.preferences and user.tenant.preferences.theme:
        return user.tenant.preferences.theme
    return "light"

def user_has_admin_role(user: User) -> bool:
    """Check if user has Admin or SuperAdmin role"""
    if not user.user_roles:
        return False
    role_names = [ur.role.name for ur in user.user_roles]
    return "Admin" in role_names or "SuperAdmin" in role_names
```

---

## Frontend Implementation

### User Context Updates

**File:** `modules/client/context/userContext.tsx`

Add theme to UserState:
```typescript
interface UserState {
  user: any | null;
  tenantName: string | null;
  bearerToken: string | null;
  isAuthed: boolean;
  theme: "light" | "dark";  // NEW
  canEditTenantTheme: boolean;  // NEW
}

const initialState: UserState = {
  user: null,
  tenantName: null,
  bearerToken: null,
  isAuthed: false,
  theme: "light",
  canEditTenantTheme: false
};
```

**File:** `modules/client/reducers/userReducer.ts`

Add actions:
```typescript
type UserAction =
  | { type: "SET_USER"; payload: any }
  | { type: "SET_TENANT_NAME"; payload: string }
  | { type: "SET_AUTHED"; payload: boolean }
  | { type: "SET_THEME"; payload: { theme: "light" | "dark"; canEditTenant: boolean } }
  | { type: "LOGOUT" };

export function userReducer(state: UserState, action: UserAction): UserState {
  switch (action.type) {
    // ... existing cases ...

    case "SET_THEME":
      return {
        ...state,
        theme: action.payload.theme,
        canEditTenantTheme: action.payload.canEditTenant
      };

    case "LOGOUT":
      return {
        ...initialState,
        theme: "light"  // Reset to light on logout
      };

    default:
      return state;
  }
}
```

### Auth State Manager Updates

**File:** `modules/client/components/AuthStateManager.tsx`

Fetch preferences on auth:
```typescript
useEffect(() => {
  async function initAuth() {
    if (user) {
      const userData = await getMe();
      dispatch({ type: "SET_USER", payload: userData });
      dispatch({ type: "SET_TENANT_NAME", payload: userData.tenant.name });
      dispatch({ type: "SET_AUTHED", payload: true });

      // NEW: Fetch user preferences
      const prefs = await fetch("/api/preferences/user", {
        headers: { Authorization: `Bearer ${bearerToken}` }
      }).then(r => r.json());

      dispatch({
        type: "SET_THEME",
        payload: {
          theme: prefs.effective_theme,
          canEditTenant: prefs.can_edit_tenant
        }
      });
    }
  }
  initAuth();
}, [user]);
```

### Theme Application

**File:** `modules/client/app/layout.tsx`

Apply theme via CSS class on `<html>`:
```typescript
export default function RootLayout({ children }) {
  const { theme } = useUser();

  return (
    <html lang="en" className={theme === "dark" ? "dark" : ""}>
      <body>{children}</body>
    </html>
  );
}
```

**File:** `modules/client/app/globals.css`

Add dark theme CSS variables:
```css
:root {
  --background: #ffffff;
  --foreground: #171717;
  --border-color: #e5e5e5;
}

html.dark {
  --background: #0a0a0a;
  --foreground: #e5e5e5;
  --border-color: #00FF00;  /* Lime green */
}

/* Apply variables to elements */
body {
  background-color: var(--background);
  color: var(--foreground);
}

.border,
[class*="border"],
input,
select,
textarea,
.card {
  border-color: var(--border-color);
}

/* Header bottom border (already lime in light mode) */
header {
  border-bottom: 1px solid var(--border-color);
}
```

**Alternative: Tailwind Dark Mode**

If using Tailwind extensively, enable dark mode in `tailwind.config.js`:
```javascript
module.exports = {
  darkMode: 'class',  // Enable class-based dark mode
  // ... rest of config
};
```

Then use Tailwind dark: variants:
```tsx
<div className="bg-white dark:bg-[#0a0a0a] border-gray-300 dark:border-[#00FF00]">
```

### User Settings Page

**File:** `modules/client/app/settings/page.tsx` (NEW)

```typescript
"use client";

import { useUser } from "@/context/userContext";
import { useState } from "react";

export default function SettingsPage() {
  const { theme, canEditTenantTheme, dispatch } = useUser();
  const [userTheme, setUserTheme] = useState<string | null>(null);
  const [tenantTheme, setTenantTheme] = useState<string>("light");

  useEffect(() => {
    // Fetch current preferences
    async function fetchPrefs() {
      const userPrefs = await fetch("/api/preferences/user").then(r => r.json());
      setUserTheme(userPrefs.theme);

      if (canEditTenantTheme) {
        const tenantPrefs = await fetch("/api/preferences/tenant").then(r => r.json());
        setTenantTheme(tenantPrefs.theme);
      }
    }
    fetchPrefs();
  }, []);

  async function updateUserTheme(newTheme: "light" | "dark" | null) {
    const res = await fetch("/api/preferences/user", {
      method: "PATCH",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ theme: newTheme })
    }).then(r => r.json());

    setUserTheme(newTheme);
    dispatch({
      type: "SET_THEME",
      payload: { theme: res.effective_theme, canEditTenant: canEditTenantTheme }
    });
  }

  async function updateTenantTheme(newTheme: "light" | "dark") {
    await fetch("/api/preferences/tenant", {
      method: "PATCH",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ theme: newTheme })
    });

    setTenantTheme(newTheme);

    // If user theme is null (inheriting), update effective theme
    if (userTheme === null) {
      dispatch({
        type: "SET_THEME",
        payload: { theme: newTheme, canEditTenant: canEditTenantTheme }
      });
    }
  }

  return (
    <div className="settings-page">
      <h1>User Settings</h1>

      <section>
        <h2>Personal Theme</h2>
        <label>
          <input
            type="radio"
            checked={userTheme === "light"}
            onChange={() => updateUserTheme("light")}
          />
          Light
        </label>
        <label>
          <input
            type="radio"
            checked={userTheme === "dark"}
            onChange={() => updateUserTheme("dark")}
          />
          Dark
        </label>
        <label>
          <input
            type="radio"
            checked={userTheme === null}
            onChange={() => updateUserTheme(null)}
          />
          Inherit from organization
        </label>
      </section>

      {canEditTenantTheme && (
        <section>
          <h2>Organization Theme (Admin)</h2>
          <p>This affects all users who inherit the organization theme.</p>
          <label>
            <input
              type="radio"
              checked={tenantTheme === "light"}
              onChange={() => updateTenantTheme("light")}
            />
            Light
          </label>
          <label>
            <input
              type="radio"
              checked={tenantTheme === "dark"}
              onChange={() => updateTenantTheme("dark")}
            />
            Dark
          </label>
        </section>
      )}
    </div>
  );
}
```

---

## Implementation Checklist

### Phase 1: Database Migration
- [ ] Create `modules/agent/src/models/TenantPreferences.py`
- [ ] Create `modules/agent/src/models/UserPreferences.py`
- [ ] Update `modules/agent/src/models/Tenant.py` (add relationship, remove settings column)
- [ ] Update `modules/agent/src/models/User.py` (add relationship)
- [ ] Create migration `005_create_preferences_tables.py`
- [ ] Run migration: `python -m agent.scripts.run_migrations`
- [ ] Verify tables created in PostgreSQL

### Phase 2: Backend API
- [ ] Create `modules/agent/src/api/routes/preferences.py`
- [ ] Implement 4 endpoints (GET/PATCH for user, GET/PATCH for tenant)
- [ ] Add role check helper: `user_has_admin_role()`
- [ ] Add theme resolver: `resolve_theme()`
- [ ] Register router in main app
- [ ] Test endpoints with curl/Postman

### Phase 3: Frontend State
- [ ] Update `UserState` interface in `context/userContext.tsx`
- [ ] Add `SET_THEME` action to `reducers/userReducer.ts`
- [ ] Update `AuthStateManager.tsx` to fetch preferences on login
- [ ] Test theme state in React DevTools

### Phase 4: Frontend Styling
- [ ] Add dark theme CSS variables to `app/globals.css`
- [ ] Apply theme class to `<html>` in `app/layout.tsx`
- [ ] Update component CSS modules to use `var(--border-color)`
- [ ] Test theme switching in browser

### Phase 5: Settings Page
- [ ] Create `app/settings/page.tsx`
- [ ] Implement user theme toggle (light/dark/inherit)
- [ ] Implement tenant theme toggle (Admin/SuperAdmin only)
- [ ] Add navigation link to settings page
- [ ] Test permission checks for tenant preferences

### Phase 6: Polish
- [ ] Test theme persistence across page refreshes
- [ ] Test theme inheritance hierarchy
- [ ] Test Admin/SuperAdmin tenant preference editing
- [ ] Verify dark theme lime green borders match header
- [ ] Update any hardcoded colors to use CSS variables

---

## Future Enhancements (Out of Scope)

### Frontend Preferences (User-level)
- Language/locale selection
- Notification preferences (in-app, email, push)
- Display density (compact/comfortable/spacious)
- Default view (dashboard, leads, jobs)
- Keyboard shortcuts customization

### Backend Preferences (User-level)
- Email digest frequency (daily/weekly/never)
- Data retention preferences
- Export format preferences (CSV/JSON/Excel)

### Tenant-Level Preferences
- Default role for new users
- Allowed integrations
- Data sharing policies
- Branding (logo, colors, custom domain)

**Implementation Pattern:**
When adding new preferences, follow same structure:
1. Add column to `UserPreferences` or `TenantPreferences`
2. Add API endpoint to get/update
3. Add UI control in settings page
4. Add to state management

---

## Testing Strategy

### Unit Tests
- Test `resolve_theme()` with all hierarchy scenarios
- Test `user_has_admin_role()` for all role types
- Test API endpoints with mock users (regular, admin, super admin)

### Integration Tests
- Test full theme preference flow (API → DB → API)
- Test theme inheritance (user null → uses tenant)
- Test permission enforcement (non-admin cannot edit tenant)
- Test migration (existing tenants get default preferences)

### E2E Tests
- User logs in → sees correct theme
- User changes theme → persists across refresh
- Admin changes tenant theme → affects users who inherit
- Non-admin cannot access tenant preferences endpoint

---

## Success Metrics

- [ ] User can toggle theme in settings page
- [ ] Theme persists across browser refresh
- [ ] Admin can set tenant-wide theme
- [ ] Users with `theme: null` inherit tenant theme
- [ ] Dark theme has lime green borders (`#00FF00`)
- [ ] Theme state loads on authentication
- [ ] No flash of unstyled content on page load
- [ ] Migration completes without data loss
