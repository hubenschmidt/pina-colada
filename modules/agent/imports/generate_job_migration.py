#!/usr/bin/env python3
"""
Generate SQL migration script from prod-data.csv
"""

import csv
from pathlib import Path

def escape_sql(value):
    """Escape single quotes for SQL"""
    if value is None or value == '(null)' or value == '':
        return None
    return value.replace("'", "''")

def generate_migration():
    csv_path = Path(__file__).parent.parent.parent.parent / 'prod-data.csv'

    with open(csv_path, 'r', encoding='utf-8') as f:
        reader = csv.DictReader(f, delimiter='\t')
        jobs = list(reader)

    sql_parts = []
    sql_parts.append("""-- ==============================
-- Production Data Migration: Migrate Job Records
-- ==============================
-- Run this AFTER 001_import_orgs.sql
-- This migrates all 188 job records from old format to new DealTracker schema
-- ==============================

DO $$
DECLARE
  default_deal_id BIGINT;
  default_deal_status_id BIGINT;
  new_lead_id BIGINT;
  org_id BIGINT;
  mapped_status_id BIGINT;
BEGIN
  -- Get default deal status for "Prospecting"
  SELECT id INTO default_deal_status_id
  FROM "Status"
  WHERE name = 'Prospecting' AND category = 'deal'
  LIMIT 1;

  -- Get or create the default deal
  SELECT id INTO default_deal_id
  FROM "Deal"
  WHERE name = 'Job Search 2025'
  LIMIT 1;

  IF default_deal_id IS NULL THEN
    INSERT INTO "Deal" (name, description, current_status_id, created_at, updated_at)
    VALUES (
      'Job Search 2025',
      'Job search applications for 2025',
      default_deal_status_id,
      '2025-11-11 11:22:46',
      NOW()
    )
    RETURNING id INTO default_deal_id;
    RAISE NOTICE 'Created Deal: Job Search 2025 (id=%)', default_deal_id;
  ELSE
    RAISE NOTICE 'Using existing Deal: Job Search 2025 (id=%)', default_deal_id;
  END IF;

""")

    for idx, job in enumerate(jobs, 1):
        company = escape_sql(job['company'])
        job_title = escape_sql(job['job_title'])
        status = job['status'].lower()
        notes = escape_sql(job['notes'])
        source = escape_sql(job['source']) or 'manual'
        job_url = escape_sql(job['job_url'])
        resume_date = escape_sql(job['resume'])
        salary_range = escape_sql(job['salary_range'])
        created_at = job['created_at']
        updated_at = job['updated_at']

        # Map status
        if status == 'applied':
            status_name = 'Applied'
        elif status == 'interviewing':
            status_name = 'Interviewing'
        elif status == 'rejected':
            status_name = 'Rejected'
        elif status == 'offer':
            status_name = 'Offer'
        elif status == 'accepted':
            status_name = 'Accepted'
        else:
            status_name = 'Applied'

        title = f"{company} - {job_title}"

        sql_parts.append(f"  -- Job {idx}: {title}\n")
        sql_parts.append(f"  SELECT id INTO org_id FROM \"Organization\" WHERE LOWER(name) = LOWER('{company}');\n")
        sql_parts.append(f"  mapped_status_id := (SELECT id FROM \"Status\" WHERE name = '{status_name}' AND category = 'job');\n")

        # Lead insert
        description = f"'{notes}'" if notes else "NULL"
        sql_parts.append(f"  INSERT INTO \"Lead\" (deal_id, type, title, description, source, current_status_id, created_at, updated_at)\n")
        sql_parts.append(f"  VALUES (default_deal_id, 'Job', '{title}', {description}, '{source}', mapped_status_id, '{created_at}', '{updated_at}')\n")
        sql_parts.append(f"  RETURNING id INTO new_lead_id;\n")

        # Job insert - build columns dynamically
        cols = ['id', 'organization_id', 'job_title']
        vals = ['new_lead_id', 'org_id', f"'{job_title}'"]

        if job_url:
            cols.append('job_url')
            vals.append(f"'{job_url}'")

        if notes:
            cols.append('notes')
            vals.append(f"'{notes}'")

        if resume_date:
            cols.append('resume_date')
            vals.append(f"'{resume_date}'")

        if salary_range:
            cols.append('salary_range')
            vals.append(f"'{salary_range}'")

        cols.extend(['created_at', 'updated_at'])
        vals.extend([f"'{created_at}'", f"'{updated_at}'"])

        sql_parts.append(f"  INSERT INTO \"Job\" ({', '.join(cols)})\n")
        sql_parts.append(f"  VALUES ({', '.join(vals)});\n\n")

        if idx % 50 == 0:
            sql_parts.append(f"  RAISE NOTICE 'Migrated {idx} jobs...';\n\n")

    sql_parts.append(f"""  RAISE NOTICE 'Migration complete! Migrated {len(jobs)} jobs';
END $$;

-- Verification queries
SELECT COUNT(*) as total_jobs FROM "Job";
SELECT COUNT(*) as total_leads_jobs FROM "Lead" WHERE type = 'Job';
SELECT
  s.name as status,
  COUNT(*) as count
FROM "Lead" l
JOIN "Status" s ON l.current_status_id = s.id
WHERE l.type = 'Job'
GROUP BY s.name
ORDER BY count DESC;
""")

    return ''.join(sql_parts)

if __name__ == '__main__':
    output_path = Path(__file__).parent / '002_migrate_jobs.sql'
    sql = generate_migration()
    output_path.write_text(sql)
    print(f"Generated {output_path}")
    print(f"Total lines: {len(sql.splitlines())}")
