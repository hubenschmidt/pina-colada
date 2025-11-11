# Supabase Integration Setup Guide

This guide covers setting up Supabase for job application tracking, replacing the Google Sheets integration.

## Table of Contents

- [Why Supabase?](#why-supabase)
- [Prerequisites](#prerequisites)
- [Part 1: Supabase Project Setup](#part-1-supabase-project-setup)
- [Part 2: Database Migration](#part-2-database-migration)
- [Part 3: Backend Configuration](#part-3-backend-configuration)
- [Part 4: Frontend Configuration](#part-4-frontend-configuration)
- [Part 5: Data Migration from Google Sheets](#part-5-data-migration-from-google-sheets)
- [Troubleshooting](#troubleshooting)

---

## Why Supabase?

Supabase provides several advantages over Google Sheets:

- **Better Data Structure**: PostgreSQL database with proper types and relationships
- **Real-time Updates**: Instant sync between backend agent and frontend UI
- **Improved Querying**: SQL-based filtering and searching
- **Free Tier**: Generous limits (500MB storage, unlimited API requests)
- **React UI**: Built-in UI for managing applications
- **Scalability**: Better performance with large datasets

---

## Prerequisites

- Supabase account (free tier is sufficient)
- Python 3.12+ with pip
- Node.js 18+ with npm
- Access to your current Google Sheets data (if migrating)

---

## Part 1: Supabase Project Setup

### 1.1 Create Supabase Account

1. Go to [https://supabase.com](https://supabase.com)
2. Click "Start your project"
3. Sign up with GitHub, Google, or email

### 1.2 Create a New Project

1. Click "New Project"
2. Fill in project details:
   - **Name**: `pina-colada-jobs` (or your preferred name)
   - **Database Password**: Generate a strong password (save this!)
   - **Region**: Choose closest to your location
   - **Pricing Plan**: Free tier is sufficient
3. Click "Create new project"
4. Wait 1-2 minutes for project provisioning

### 1.3 Get API Credentials

1. Navigate to **Settings** > **API**
2. Copy these values:
   - **Project URL**: `https://xxxxx.supabase.co`
   - **anon public**: Your public key (safe for frontend)
   - **service_role**: Your service role key (keep secret!)

---

## Part 2: Database Migration

You have four options for running the database migrations:

### Option 0: Docker Compose (Easiest for Dev)

If using Docker Compose, migrations run automatically on container startup:

1. Add Supabase env vars to `modules/agent/.env`:
   ```bash
   SUPABASE_URL="https://your-project-ref.supabase.co"
   SUPABASE_SERVICE_KEY="your-service-role-key"
   SUPABASE_DB_PASSWORD="your-db-password"
   ```

2. Start services:
   ```bash
   docker-compose up
   ```

3. Watch logs for migration output:
   ```bash
   docker-compose logs -f agent
   ```

That's it! Migrations run automatically. See `DOCKER_MIGRATIONS.md` for details.

### Option A: Supabase Dashboard (Easiest)

1. Navigate to **SQL Editor** in your Supabase dashboard
2. Click "New Query"
3. Copy the contents of `supabase_migrations/001_initial_schema.sql`
4. Paste into the SQL editor
5. Click "Run" to execute

### Option B: Python Migration Script (Recommended)

1. Install dependencies:
   ```bash
   cd modules/agent
   pip install supabase psycopg2-binary
   ```

2. Set environment variables:
   ```bash
   export SUPABASE_URL="https://your-project-ref.supabase.co"
   export SUPABASE_SERVICE_KEY="your-service-role-key"
   export SUPABASE_DB_PASSWORD="your-database-password"
   ```

3. Run migration script:
   ```bash
   python scripts/apply_migrations.py
   ```

### Option C: Supabase CLI

1. Install Supabase CLI:
   ```bash
   npm install -g supabase
   ```

2. Login and link project:
   ```bash
   supabase login
   supabase link --project-ref your-project-ref
   ```

3. Apply migrations:
   ```bash
   supabase db push
   ```

### Verify Migration

Check that the `applied_jobs` table was created:

1. Go to **Table Editor** in Supabase dashboard
2. You should see the `applied_jobs` table with these columns:
   - id (uuid)
   - company (text)
   - job_title (text)
   - application_date (timestamp)
   - status (text)
   - job_url (text)
   - location (text)
   - salary_range (text)
   - notes (text)
   - source (text)
   - created_at (timestamp)
   - updated_at (timestamp)

---

## Part 3: Backend Configuration

### 3.1 Install Python Dependencies

```bash
cd modules/agent
pip install supabase>=2.0.0
```

Or use the updated pyproject.toml:

```bash
pip install -e .
```

### 3.2 Configure Environment Variables

Create or update `modules/agent/.env`:

```bash
# Supabase Configuration
USE_SUPABASE=true
SUPABASE_URL="https://your-project-ref.supabase.co"
SUPABASE_SERVICE_KEY="your-service-role-key"

# Optional: For migration scripts
SUPABASE_DB_PASSWORD="your-database-password"
```

### 3.3 Test Backend Connection

```python
from agent.services.supabase_client import get_applied_jobs_tracker

tracker = get_applied_jobs_tracker()
jobs = tracker.fetch_applied_jobs()
print(f"Found {len(jobs)} jobs")
```

---

## Part 4: Frontend Configuration

### 4.1 Install JavaScript Dependencies

```bash
cd modules/client
npm install @supabase/supabase-js
```

### 4.2 Configure Environment Variables

Create `modules/client/.env.local`:

```bash
NEXT_PUBLIC_SUPABASE_URL="https://your-project-ref.supabase.co"
NEXT_PUBLIC_SUPABASE_ANON_KEY="your-anon-public-key"

# Job Tracker Password (Server-side only for production)
# In development, uses hardcoded password for convenience
JOBS_PASSWORD="your-secure-password"
```

**Important**:
- Use the `anon` key (not the `service_role` key) for the frontend!
- `JOBS_PASSWORD` is server-side only (no `NEXT_PUBLIC_` prefix)
- In production, password is validated via API route (not exposed to browser)
- In development (localhost), uses hardcoded password for convenience

### 4.3 Build and Run

```bash
npm run build
npm run dev
```

Navigate to `http://localhost:3001/jobs` to see the Job Tracker UI.

---

## Part 5: Data Migration from Google Sheets

If you have existing job applications in Google Sheets, migrate them to Supabase.

### 5.1 Prerequisites

- Google Sheets integration still configured
- `gspread` and `google-auth` installed
- `GOOGLE_SHEETS_CREDENTIALS_JSON` or `GOOGLE_SHEETS_CREDENTIALS_PATH` set

### 5.2 Run Migration Script

**Dry Run (recommended first):**

```bash
cd modules/agent
python scripts/migrate_sheets_to_supabase.py --dry-run
```

This will show what would be migrated without making changes.

**Actual Migration:**

```bash
python scripts/migrate_sheets_to_supabase.py
```

### 5.3 Verify Migration

1. Check Supabase dashboard **Table Editor** > `applied_jobs`
2. Verify row count matches Google Sheets
3. Spot-check a few entries for accuracy

### 5.4 Switch to Supabase

Once verified, update your `.env`:

```bash
USE_SUPABASE=true
```

The system will now read from and write to Supabase instead of Google Sheets.

---

## Database Schema

### applied_jobs Table

```sql
CREATE TABLE applied_jobs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company TEXT NOT NULL,
    job_title TEXT NOT NULL,
    application_date TIMESTAMP DEFAULT NOW(),
    status TEXT DEFAULT 'applied' CHECK (status IN ('applied', 'interviewing', 'rejected', 'offer', 'accepted')),
    job_url TEXT,
    location TEXT,
    salary_range TEXT,
    notes TEXT,
    source TEXT DEFAULT 'manual' CHECK (source IN ('manual', 'agent')),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

### Indexes

- `idx_company_title`: Fast lookups by company and job title
- `idx_application_date`: Sorted by application date
- `idx_status`: Filter by status

### Row Level Security (RLS)

The migrations enable RLS with permissive policies for now. For production, consider:

1. Adding authentication
2. Restricting policies to authenticated users
3. Adding user-specific data isolation

---

## Frontend Usage

### Authentication

The `/jobs` route is protected with password authentication:

**Development (localhost):**
1. Navigate to `http://localhost:3001/jobs`
2. Enter the password: `pinacolada2024`
3. Password validated client-side (for convenience)
4. Access granted

**Production:**
1. Navigate to `https://yoursite.com/jobs`
2. Enter the password
3. Password sent to server API route for validation
4. Server checks against `JOBS_PASSWORD` env variable
5. Access granted if valid
6. Click "Logout" button (top right) to log out

**Changing the Password:**

- **Development**: Edit `DEV_PASSWORD` in `lib/auth.ts`
- **Production**: Set `JOBS_PASSWORD` in your `.env.local` or hosting platform:

```bash
JOBS_PASSWORD="mySecurePassword123"
```

**Security Note**:
- **Development**: Password is hardcoded for convenience (suitable for local dev)
- **Production**: Password is server-side only, never exposed to browser
- Suitable for personal/single-user use
- For production with multiple users, consider:
- Implementing Supabase Auth with email/password
- Adding Row Level Security based on user ID
- Using server-side session management

### Job Tracker UI Features

After logging in, you can access the Job Tracker:

- **View Applications**: See all job applications in a card layout
- **Add New**: Click "Add New Application" to manually add jobs
- **Edit**: Click the edit icon to update job details
- **Delete**: Click delete icon twice to confirm deletion
- **Filter**: Filter by status (applied, interviewing, rejected, offer, accepted)
- **Search**: Search by company, title, location, or notes
- **Real-time Updates**: See changes instantly when the agent adds jobs

### Real-time Sync

The frontend automatically subscribes to database changes. When your agent adds a job application, it appears instantly in the UI without refreshing.

---

## Agent Integration

The agent now uses Supabase for job filtering:

```python
# In worker_tools.py
from agent.services.supabase_client import get_applied_jobs_tracker

tracker = get_applied_jobs_tracker()

# Check if job was applied to
if tracker.is_job_applied("Google", "Software Engineer"):
    print("Already applied!")

# Filter job search results
filtered_jobs = tracker.filter_jobs(all_jobs)

# Add new application (optional)
tracker.add_applied_job(
    company="Google",
    job_title="Software Engineer",
    job_url="https://...",
    source="agent"
)
```

---

## Troubleshooting

### Backend Issues

**Error: "SUPABASE_URL or SUPABASE_SERVICE_KEY not set"**

Solution: Verify your `.env` file has the correct variables set.

**Error: "Failed to fetch applied jobs from Supabase"**

Solutions:
1. Check that migrations were applied successfully
2. Verify API keys are correct
3. Check Supabase project status in dashboard
4. Review network/firewall settings

**Connection Timeout**

Solutions:
1. Check your internet connection
2. Verify Supabase project is active (free tier pauses after inactivity)
3. Try a different network

### Frontend Issues

**Error: "Supabase environment variables not configured"**

Solution: Create `.env.local` with `NEXT_PUBLIC_SUPABASE_URL` and `NEXT_PUBLIC_SUPABASE_ANON_KEY`.

**Jobs not loading**

Solutions:
1. Open browser DevTools > Console for errors
2. Check Network tab for failed requests
3. Verify Row Level Security policies allow reads
4. Confirm `applied_jobs` table exists

**Real-time updates not working**

Solutions:
1. Check that Realtime is enabled in Supabase dashboard (Database > Replication)
2. Enable replication for `applied_jobs` table
3. Verify no ad blockers are blocking WebSocket connections

### Migration Issues

**Duplicate jobs after migration**

The migration script checks for duplicates by company + job_title. If duplicates still occur:

1. Manually deduplicate in Supabase dashboard
2. Run migration with `--dry-run` first to preview

**Some jobs missing after migration**

Check:
1. Google Sheets rows with "to apply" or "did not apply" are intentionally skipped
2. Rows without a job title are skipped
3. Review migration script logs for errors

---

## Feature Flag: Switching Between Google Sheets and Supabase

The system supports running with either backend:

### Use Supabase (default)

```bash
USE_SUPABASE=true
```

### Use Google Sheets (fallback)

```bash
USE_SUPABASE=false
```

This allows for:
- Testing Supabase without breaking existing Google Sheets integration
- Gradual migration
- Fallback in case of issues

---

## Cost Considerations

### Free Tier Limits

- **Database**: 500 MB storage
- **API Requests**: Unlimited (with rate limits)
- **Bandwidth**: 5 GB egress per month
- **Realtime Connections**: 200 concurrent

### Typical Usage

For job tracking:
- Each job record: ~1 KB
- 1000 jobs: ~1 MB
- API calls: ~100/day
- Well within free tier limits

---

## Next Steps

1. ✅ Complete this setup guide
2. ✅ Migrate existing Google Sheets data
3. ✅ Test the `/jobs` UI
4. ✅ Verify agent job filtering works
5. Consider adding authentication to `/jobs` route
6. Set up backups (Supabase has automatic daily backups)
7. Monitor usage in Supabase dashboard

---

## Support

If you encounter issues:

1. Check the [Troubleshooting](#troubleshooting) section
2. Review Supabase logs (Logs & Analytics in dashboard)
3. Check GitHub issues for this project
4. Consult [Supabase Documentation](https://supabase.com/docs)

---

## Additional Resources

- [Supabase Quickstart](https://supabase.com/docs/guides/getting-started/quickstarts/python)
- [Supabase Python Client](https://supabase.com/docs/reference/python/introduction)
- [Supabase JavaScript Client](https://supabase.com/docs/reference/javascript/introduction)
- [Row Level Security Guide](https://supabase.com/docs/guides/auth/row-level-security)
