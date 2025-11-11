"""Import jobs from CSV file into the database.

Usage:
    # Local development (from modules/agent directory):
    uv run python scripts/import_jobs.py imports/jobs.csv --dry-run
    uv run python scripts/import_jobs.py imports/jobs.csv

    # Production (from server):
    python scripts/import_jobs.py imports/jobs.csv --dry-run
    python scripts/import_jobs.py imports/jobs.csv
"""

import csv
import sys
import os
from datetime import datetime
from pathlib import Path

# Add src to path
sys.path.insert(0, str(Path(__file__).parent.parent / "src"))

# Load environment variables
try:
    from dotenv import load_dotenv
    load_dotenv(override=True)
except ImportError:
    pass

from agent.repositories.job_repository import create_job_from_orm, find_all_jobs
from agent.models.Job import Job


def parse_date(date_str: str) -> datetime | None:
    """Parse date string to datetime."""
    if not date_str or date_str.strip() == "":
        return None
    
    date_str = date_str.strip()
    
    # Skip non-date values
    skip_values = ["did not apply", "to apply", "jen to apply", "need to finish"]
    if date_str.lower() in skip_values:
        return None
    
    formats = [
        "%m/%d/%Y",  # 11/1/2024
        "%m/%d",     # 11/1 (assume current year)
        "%m/%d/%y",  # 11/1/24
    ]
    
    def _try_parse_format(fmt: str) -> datetime | None:
        try:
            parsed = datetime.strptime(date_str, fmt)
            # If no year in format, assume current year
            if fmt == "%m/%d":
                parsed = parsed.replace(year=datetime.now().year)
            return parsed
        except ValueError:
            return None
    
    for fmt in formats:
        parsed = _try_parse_format(fmt)
        if parsed:
            return parsed
    
    return None


def normalize_status(status: str, date_str: str) -> str:
    """Normalize status to valid values."""
    if status and status.strip():
        status_lower = status.lower().strip()
        valid_statuses = ["applied", "interviewing", "rejected", "offer", "accepted"]
        
        for valid in valid_statuses:
            if valid.lower() == status_lower:
                return valid
    
    # Infer from date field
    if date_str:
        date_lower = date_str.lower()
        if "interview" in date_lower:
            return "interviewing"
    
    return "applied"


def should_skip_row(row: dict) -> bool:
    """Determine if row should be skipped."""
    # Try multiple possible field names (case-insensitive)
    def _get_field_insensitive(row: dict, *keys: str) -> str:
        for key in keys:
            if key in row:
                return row[key].strip() if row[key] else ""
            for row_key in row.keys():
                if row_key.lower() == key.lower():
                    return row[row_key].strip() if row[row_key] else ""
        return ""
    
    description = _get_field_insensitive(row, "Description", "description", "company")
    job_title = _get_field_insensitive(row, "Job title", "job title", "job_title", "title")
    date_applied = _get_field_insensitive(row, "date applied", "date", "date_applied").lower()
    
    # Handle case where description and job title are combined in description field
    # Format: "Company Name – Job Title" or "Company Name - Job Title"
    if description and not job_title:
        import re
        # Try to split on em-dash (–) or regular dash (-) if followed by space
        parts = re.split(r'\s*[–-]\s+', description, maxsplit=1)
        if len(parts) == 2:
            description = parts[0].strip()
            job_title = parts[1].strip()
        elif '–' in description:
            parts = description.rsplit('–', 1)
            if len(parts) == 2:
                description = parts[0].strip()
                job_title = parts[1].strip()
        elif ' - ' in description:
            parts = description.rsplit(' - ', 1)
            if len(parts) == 2:
                description = parts[0].strip()
                job_title = parts[1].strip()
    
    # Skip if missing required fields (after trying to split)
    if not description or not job_title:
        return True
    
    # Skip if explicitly marked as not applied
    skip_indicators = ["did not apply", "to apply", "jen to apply", "need to finish"]
    if date_applied in skip_indicators:
        return True
    
    return False


def check_duplicate(job: Job, existing_jobs: list[Job]) -> bool:
    """Check if job already exists in database."""
    for existing in existing_jobs:
        # Match on company, job_title, and date (if available)
        if (existing.company.lower() == job.company.lower() and
            existing.job_title.lower() == job.job_title.lower()):
            # If both have dates, they must match
            if job.date and existing.date:
                if job.date.date() == existing.date.date():
                    return True
            # If only one has a date, consider it a duplicate if company+title match
            elif job.date is None and existing.date is None:
                return True
    return False


def import_csv(file_path: str, dry_run: bool = False, skip_duplicates: bool = True) -> None:
    """Import jobs from CSV file."""
    
    # Resolve file path relative to script directory if needed
    script_dir = Path(__file__).parent.parent
    original_path = file_path
    
    if not os.path.isabs(file_path):
        potential_path = script_dir / file_path
        if potential_path.exists():
            file_path = str(potential_path)
    
    if not os.path.exists(file_path):
        print(f"Error: File not found: {file_path}")
        print(f"  Tried: {file_path}")
        if not os.path.isabs(original_path):
            print(f"  Also tried: {script_dir / original_path}")
        sys.exit(1)
    
    # Load existing jobs to check for duplicates
    existing_jobs = []
    if skip_duplicates and not dry_run:
        try:
            print("Loading existing jobs to check for duplicates...")
            existing_jobs = find_all_jobs()
            print(f"  Found {len(existing_jobs)} existing jobs")
        except Exception as e:
            print(f"⚠️  Warning: Could not load existing jobs: {e}")
            print("  Continuing without duplicate check...")
    
    imported = 0
    skipped = 0
    duplicates = 0
    errors = []
    
    with open(file_path, "r", encoding="utf-8") as f:
        reader = csv.DictReader(f)
        
        def _get_field(row: dict[str, str], *possible_keys: str) -> str:
            """Get field value trying multiple possible key names (case-insensitive)."""
            for key in possible_keys:
                # Try exact match first
                if key in row:
                    return row[key].strip() if row[key] else ""
                # Try case-insensitive match
                for row_key in row.keys():
                    if row_key.lower() == key.lower():
                        return row[row_key].strip() if row[row_key] else ""
            return ""
        
        def _process_row(row_num: int, row: dict[str, str]) -> Job | None:
            if should_skip_row(row):
                return None
            
            # Handle multiple CSV formats (case-insensitive)
            description = _get_field(row, "Description", "description", "company")
            job_title = _get_field(row, "Job title", "job title", "job_title", "title")
            date_str = _get_field(row, "date applied", "date", "date_applied", "application_date")
            job_url = _get_field(row, "link", "url", "job_url", "job_link")
            notes = _get_field(row, "note", "notes")
            resume_str = _get_field(row, "Resume", "resume")
            source = _get_field(row, "job board", "job_board", "source") or "manual"
            
            # Handle case where description and job title are combined in description field
            # Format: "Company Name – Job Title" or "Company Name - Job Title"
            if description and not job_title:
                # Try to split on em-dash (–) or regular dash (-) if followed by space
                import re
                # Match em-dash or dash followed by space
                parts = re.split(r'\s*[–-]\s+', description, maxsplit=1)
                if len(parts) == 2:
                    description = parts[0].strip()
                    job_title = parts[1].strip()
                # If still no job title, try splitting on last dash
                elif not job_title and '–' in description:
                    parts = description.rsplit('–', 1)
                    if len(parts) == 2:
                        description = parts[0].strip()
                        job_title = parts[1].strip()
                elif not job_title and ' - ' in description:
                    parts = description.rsplit(' - ', 1)
                    if len(parts) == 2:
                        description = parts[0].strip()
                        job_title = parts[1].strip()
            
            # Parse resume date (if provided)
            resume_date = None
            if resume_str:
                parsed_resume = parse_date(resume_str)
                if parsed_resume:
                    resume_date = parsed_resume
                else:
                    # If resume is not a date, add it to notes
            if notes:
                        notes = f"{notes} | Resume: {resume_str}"
                    else:
                        notes = f"Resume: {resume_str}"
            
            job = Job(
                company=description,
                job_title=job_title,
                job_url=job_url,
                salary_range=None,
                notes=notes or None,
                resume=resume_date,
                status="applied",
                source=source if source else "manual",
            )
            
            # Parse application date (now called "date")
            if date_str:
                parsed_date = parse_date(date_str)
                if parsed_date:
                    job.date = parsed_date
            
            return job
        
        for row_num, row in enumerate(reader, start=2):
            try:
                job = _process_row(row_num, row)
                if not job:
                    skipped += 1
                    continue
                
                # Check for duplicates
                if skip_duplicates and check_duplicate(job, existing_jobs):
                    duplicates += 1
                    print(f"⊘ Row {row_num}: {job.company} - {job.job_title} (duplicate, skipping)")
                    continue
                
                    if not dry_run:
                        create_job_from_orm(job)
                    # Add to existing jobs list to prevent duplicates within the same import
                    existing_jobs.append(job)
                    
                    imported += 1
                    print(f"✓ Row {row_num}: {job.company} - {job.job_title}")
            
            except Exception as e:
                skipped += 1
                error_msg = f"Row {row_num}: {str(e)}"
                errors.append(error_msg)
                print(f"✗ {error_msg}")
    
    print(f"\n{'[DRY RUN] ' if dry_run else ''}Import complete:")
    print(f"  Imported: {imported}")
    print(f"  Skipped: {skipped}")
    if duplicates > 0:
        print(f"  Duplicates (skipped): {duplicates}")
    
    if errors:
        print(f"\nErrors ({len(errors)}):")
        for error in errors[:10]:
            print(f"  - {error}")
        if len(errors) > 10:
            print(f"  ... and {len(errors) - 10} more")


if __name__ == "__main__":
    import argparse
    
    parser = argparse.ArgumentParser(
        description="Import jobs from CSV file into the database",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  # Dry run (preview without saving)
  uv run python scripts/import_jobs.py imports/jobs.csv --dry-run
  
  # Actual import
  uv run python scripts/import_jobs.py imports/jobs.csv
  
  # Import without duplicate checking
  uv run python scripts/import_jobs.py imports/jobs.csv --no-skip-duplicates
        """
    )
    parser.add_argument("csv_file", help="Path to CSV file (relative to modules/agent or absolute)")
    parser.add_argument("--dry-run", action="store_true", help="Preview import without saving to database")
    parser.add_argument("--no-skip-duplicates", action="store_true", help="Allow duplicate imports (not recommended)")
    
    args = parser.parse_args()
    
    print("=" * 80)
    print(f"{'[DRY RUN] ' if args.dry_run else ''}Importing jobs from: {args.csv_file}")
    print("=" * 80)
    
    import_csv(
        args.csv_file,
        dry_run=args.dry_run,
        skip_duplicates=not args.no_skip_duplicates
    )

