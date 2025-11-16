# Production Data Import Scripts

This directory contains scripts to import production data into the new DealTracker schema.

## Prerequisites

1. **Backup your production database first!**
2. Run the base migration: `modules/agent/migrations/002_dealtracker.sql`

## Import Order

Run these scripts in **exact order**:

### 1. Import Organizations
```sql
-- modules/agent/imports/001_import_orgs.sql
```
Creates 176 unique Organization records from production data.

### 2. Migrate Jobs
```sql
-- modules/agent/imports/002_migrate_jobs.sql
```
Migrates all 187 job records from CSV export to the new schema structure:
- Creates Lead records for each job
- Creates Job records linked to Organizations
- Maps statuses (applied → Applied, interviewing → Interviewing, rejected → Rejected)
- Preserves all metadata (URLs, notes, salary_range, resume_date, timestamps)

## Verification Queries

After running the imports, verify your data:

```sql
-- Check organization count
SELECT COUNT(*) as total_organizations FROM "Organization";
-- Expected: 176

-- Check job count
SELECT COUNT(*) as total_jobs FROM "Job";
-- Expected: 187

-- Check Lead/Job relationship
SELECT COUNT(*) as total_leads_jobs FROM "Lead" WHERE type = 'Job';
-- Expected: 187

-- Check status distribution
SELECT
  s.name as status,
  COUNT(*) as count
FROM "Lead" l
JOIN "Status" s ON l.current_status_id = s.id
WHERE l.type = 'Job'
GROUP BY s.name
ORDER BY count DESC;
```

## Generator Script

The `generate_job_migration.py` script was used to generate `002_migrate_jobs.sql` from the CSV export. You can regenerate this file if needed:

```bash
python3 generate_job_migration.py
```

## Source Data

The import scripts were generated from: `prod-data.csv` (root directory)

This CSV was exported from production using:
```sql
SELECT
  id,
  company,
  job_title,
  status,
  notes,
  source,
  job_url,
  resume,
  salary_range,
  created_at,
  updated_at
FROM "Job"
ORDER BY created_at;
```
