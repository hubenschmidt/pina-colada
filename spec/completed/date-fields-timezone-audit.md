# Date Fields Timezone Audit

## Overview
This document audits all user-controlled date fields in the codebase for timezone handling consistency.

## Current Implementation (Updated 2025-12-01)
- **Database**: PostgreSQL configured for UTC via migration `044_set_database_utc.sql`
- **Storage**: All timestamps stored in UTC with `TIMESTAMPTZ` columns
- **API Format**: ISO8601 with `Z` suffix (e.g., `2025-12-01T15:30:00Z`)
- **Display**: Converted to `MM/DD/YYYY` via `format_display_date()`
- **User Preference**: `UserPreferences.timezone` stores user's timezone (default: `America/New_York`)

## Date Fields Inventory

| Entity | Field | Column Type | Handling | Notes |
|--------|-------|-------------|----------|-------|
| Job | `resume_date` | TIMESTAMPTZ | UTC | User-selected date with UTC time |
| Partnership | `start_date` | DATE | Date-only | No timezone needed |
| Partnership | `end_date` | DATE | Date-only | No timezone needed |
| Task | `start_date` | DATE | Date-only | No timezone needed |
| Task | `due_date` | DATE | Date-only | No timezone needed |
| Task | `completed_at` | TIMESTAMPTZ | UTC | Timestamp of completion |
| Project | `start_date` | DATE | Date-only | No timezone needed |
| Project | `end_date` | DATE | Date-only | No timezone needed |
| Deal | `expected_close_date` | DATE | Date-only | No timezone needed |
| Deal | `close_date` | DATE | Date-only | No timezone needed |
| Opportunity | `expected_close_date` | DATE | Date-only | No timezone needed |
| FundingRound | `announced_date` | DATE | Date-only | Historical reference |
| CompanySignal | `signal_date` | DATE | Date-only | Historical reference |

## Best Practices (Implemented)

### 1. Store UTC, Display Local
- Accept timezone preference from user during onboarding
- Store all datetimes in UTC in database
- Convert to user's timezone for display in frontend
- Accept time input in user's timezone, convert to UTC in backend

### 2. Date Format Consistency
- Database exports: ISO8601 with `Z` suffix
- API responses: ISO8601 with `Z` suffix
- Display: `MM/DD/YYYY` for US locale

### 3. DST Edge Cases
- DST causes 1 hour to "disappear" (spring) or "repeat" (fall)
- A 01:30 event can occur before a 01:20 event (50 mins apart actual)
- For scheduling features, use timezone-aware scheduler (Quartz, Google Scheduler)
- Test specifically for DST transition dates

### 4. Server-Side Conversion
- **Always** do timezone conversion on server side
- Never trust client timezone handling
- Use `zoneinfo` (Python 3.9+) or `pytz` for conversions

## Implementation Reference

```python
from datetime import datetime, timezone
from zoneinfo import ZoneInfo

# Store (frontend → backend → DB)
def to_utc(local_dt: datetime, user_tz: str) -> datetime:
    user_zone = ZoneInfo(user_tz)
    if local_dt.tzinfo is None:
        local_dt = local_dt.replace(tzinfo=user_zone)
    return local_dt.astimezone(timezone.utc)

# Display (DB → backend → frontend)
def to_local(utc_dt: datetime, user_tz: str) -> datetime:
    user_zone = ZoneInfo(user_tz)
    return utc_dt.astimezone(user_zone)

# Format for API response
def format_utc(dt: datetime) -> str:
    return dt.strftime("%Y-%m-%dT%H:%M:%SZ")
```

## Migration Applied
- `044_set_database_utc.sql`: Sets database timezone to UTC, updates trigger function

---
*Updated: 2025-12-01*
