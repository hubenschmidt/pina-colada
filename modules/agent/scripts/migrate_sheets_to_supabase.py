#!/usr/bin/env python3
"""Migrate job application data from Google Sheets to Supabase.

This script reads all job applications from the Google Sheets tracker
and imports them into the Supabase applied_jobs table.
"""

import sys
from pathlib import Path
from typing import List, Dict, Any

# Add parent directory to path to import from agent package
sys.path.insert(0, str(Path(__file__).parent.parent / "src"))

import logging
from dotenv import load_dotenv

load_dotenv(override=True)

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

try:
    from supabase import create_client, Client
    import os
    SUPABASE_AVAILABLE = True
except ImportError:
    print("Error: supabase-py not installed. Run: pip install supabase")
    sys.exit(1)

try:
    from agent.services.google_sheets import get_applied_jobs_tracker as get_sheets_tracker
    SHEETS_AVAILABLE = True
except ImportError:
    SHEETS_AVAILABLE = False
    logger.warning("Google Sheets service not available")


def get_supabase_client() -> Client:
    """Create Supabase client from environment variables."""
    supabase_url = os.getenv("SUPABASE_URL")
    supabase_key = os.getenv("SUPABASE_SERVICE_KEY")

    if not supabase_url or not supabase_key:
        logger.error("SUPABASE_URL or SUPABASE_SERVICE_KEY not set")
        sys.exit(1)

    return create_client(supabase_url, supabase_key)


def map_sheets_job_to_supabase(job: Dict[str, str]) -> Dict[str, Any]:
    """Map Google Sheets job format to Supabase format."""
    return {
        "company": job.get("company", "Unknown Company"),
        "job_title": job.get("title", ""),
        "job_url": job.get("link", ""),
        "notes": f"Migrated from Google Sheets. Applied: {job.get('date_applied', 'Unknown')}",
        "status": "applied",
        "source": "manual"
    }


def check_if_job_exists(client: Client, company: str, job_title: str) -> bool:
    """Check if job already exists in Supabase."""
    try:
        response = client.table("applied_jobs").select("id").eq(
            "company", company
        ).eq("job_title", job_title).execute()

        return len(response.data) > 0
    except Exception as e:
        logger.error(f"Error checking for existing job: {e}")
        return False


def migrate_job(client: Client, job: Dict[str, str], dry_run: bool = False) -> bool:
    """Migrate a single job from Google Sheets to Supabase."""
    supabase_job = map_sheets_job_to_supabase(job)

    if not supabase_job["job_title"]:
        logger.warning(f"Skipping job with no title: {job}")
        return False

    # Check for duplicates
    if check_if_job_exists(client, supabase_job["company"], supabase_job["job_title"]):
        logger.info(f"Job already exists: {supabase_job['company']} - {supabase_job['job_title']}")
        return False

    if dry_run:
        logger.info(f"[DRY RUN] Would insert: {supabase_job['company']} - {supabase_job['job_title']}")
        return True

    try:
        response = client.table("applied_jobs").insert(supabase_job).execute()

        if response.data:
            logger.info(f"✓ Migrated: {supabase_job['company']} - {supabase_job['job_title']}")
            return True

        logger.error(f"Failed to insert: {supabase_job['company']} - {supabase_job['job_title']}")
        return False

    except Exception as e:
        logger.error(f"Error migrating job: {e}")
        return False


def main():
    """Main entry point."""
    print("="*60)
    print("Google Sheets → Supabase Migration Tool")
    print("="*60)

    if not SHEETS_AVAILABLE:
        print("\nError: Google Sheets service not available")
        print("Make sure gspread and google-auth are installed")
        sys.exit(1)

    # Parse command line arguments
    dry_run = "--dry-run" in sys.argv

    if dry_run:
        print("\n⚠️  DRY RUN MODE - No data will be written to Supabase")

    print("\n1. Fetching jobs from Google Sheets...")
    try:
        sheets_tracker = get_sheets_tracker()
        sheets_jobs = sheets_tracker.get_jobs_details(refresh=True)

        if not sheets_jobs:
            print("\n❌ No jobs found in Google Sheets")
            return

        print(f"✓ Found {len(sheets_jobs)} jobs in Google Sheets")

    except Exception as e:
        print(f"\n❌ Failed to fetch from Google Sheets: {e}")
        import traceback
        print(traceback.format_exc())
        return

    print("\n2. Connecting to Supabase...")
    try:
        supabase_client = get_supabase_client()
        print("✓ Connected to Supabase")

    except Exception as e:
        print(f"\n❌ Failed to connect to Supabase: {e}")
        return

    print("\n3. Migrating jobs...")
    success_count = 0
    skip_count = 0
    error_count = 0

    for job in sheets_jobs:
        result = migrate_job(supabase_client, job, dry_run=dry_run)

        if result:
            success_count += 1
        elif check_if_job_exists(supabase_client, job.get("company", ""), job.get("title", "")):
            skip_count += 1
        else:
            error_count += 1

    print("\n" + "="*60)
    print("Migration Summary")
    print("="*60)
    print(f"Total jobs in Google Sheets: {len(sheets_jobs)}")
    print(f"Successfully migrated: {success_count}")
    print(f"Skipped (already exist): {skip_count}")
    print(f"Errors: {error_count}")

    if dry_run:
        print("\n⚠️  This was a DRY RUN. Run without --dry-run to actually migrate data.")
    else:
        print("\n✓ Migration complete!")


if __name__ == "__main__":
    main()
