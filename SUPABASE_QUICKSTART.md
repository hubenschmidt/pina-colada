# Supabase Integration - Quick Start Guide

Get up and running with Supabase in 10 minutes.

## Step 1: Create Supabase Project (3 min)

1. Go to https://supabase.com and sign up
2. Click "New Project"
3. Fill in:
   - Name: `pina-colada-jobs`
   - Database Password: (generate and save)
   - Region: Choose closest to you
4. Click "Create new project" and wait ~2 minutes

## Step 2: Get Your Credentials (1 min)

1. Go to **Settings** → **API**
2. Copy these three values:
   ```
   Project URL: https://xxxxx.supabase.co
   anon public: eyJhbG...
   service_role: eyJhbG...
   ```

## Step 3: Run Database Migration (2 min)

Open Supabase Dashboard → **SQL Editor** → "New Query"

Copy/paste the contents of `supabase_migrations/001_initial_schema.sql` and click **Run**.

You should see "Success. No rows returned" message.

## Step 4: Configure Backend (2 min)

```bash
cd modules/agent

# Create .env file (or update existing)
cat >> .env << EOF
USE_SUPABASE=true
SUPABASE_URL="https://xxxxx.supabase.co"
SUPABASE_SERVICE_KEY="your-service-role-key"
EOF

# Install Supabase
pip install supabase
```

## Step 5: Configure Frontend (2 min)

```bash
cd modules/client

# Create .env.local file
cat >> .env.local << EOF
NEXT_PUBLIC_SUPABASE_URL="https://xxxxx.supabase.co"
NEXT_PUBLIC_SUPABASE_ANON_KEY="your-anon-public-key"
# For production, set JOBS_PASSWORD (no NEXT_PUBLIC_ prefix)
JOBS_PASSWORD="your-secure-password"
EOF

# Install dependencies
npm install
```

**Note**:
- Development uses hardcoded password: `pinacolada2024`
- Production uses `JOBS_PASSWORD` env var (server-side only, not exposed to browser)

## Step 6: Test It! (1 min)

### Test Backend:
```bash
cd modules/agent
python -c "from agent.services.supabase_client import get_applied_jobs_tracker; tracker = get_applied_jobs_tracker(); jobs = tracker.fetch_applied_jobs(); print(f'✅ Backend working! Found {len(jobs)} jobs')"
```

### Test Frontend:
```bash
cd modules/client
npm run dev
# Open http://localhost:3001/jobs
```

You should see a login screen.
- **Development password**: `pinacolada2024`
- Enter it and you'll see the Job Tracker UI!
- In production, it will use the `JOBS_PASSWORD` from your server environment

## Step 7: Migrate Your Google Sheets Data (Optional)

If you have existing data in Google Sheets:

```bash
cd modules/agent

# Dry run first (shows what will happen)
python scripts/migrate_sheets_to_supabase.py --dry-run

# If it looks good, run for real
python scripts/migrate_sheets_to_supabase.py
```

## ✅ You're Done!

- **Backend**: Agent will now read/write to Supabase
- **Frontend**: Visit http://localhost:3001/jobs to manage applications
- **Real-time**: Changes sync instantly between agent and UI

## Need Help?

- Full setup guide: See `SUPABASE_SETUP.md`
- Implementation details: See `IMPLEMENTATION_SUMMARY.md`
- Troubleshooting: See SUPABASE_SETUP.md → Troubleshooting section

## Quick Commands Reference

```bash
# Backend: Run migrations
cd modules/agent
python scripts/apply_migrations.py

# Backend: Migrate from Google Sheets
python scripts/migrate_sheets_to_supabase.py --dry-run
python scripts/migrate_sheets_to_supabase.py

# Frontend: Start dev server
cd modules/client
npm run dev

# Frontend: Build for production
npm run build
npm start
```

## Switching Back to Google Sheets (Emergency Fallback)

If you need to temporarily switch back:

```bash
# In modules/agent/.env
USE_SUPABASE=false
```

Restart your agent and it will use Google Sheets instead.
