"""Import jobs from CSV file into the database."""

import csv
import sys
import os
from datetime import datetime
from pathlib import Path

# Add src to path
sys.path.insert(0, str(Path(__file__).parent.parent / "src"))

from agent.repositories.job_repository import JobRepository
from agent.models.job import AppliedJob


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
    
    for fmt in formats:
        try:
            parsed = datetime.strptime(date_str, fmt)
            # If no year in format, assume current year
            if fmt == "%m/%d":
                parsed = parsed.replace(year=datetime.now().year)
            return parsed
        except ValueError:
            continue
    
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
    company = row.get("Description", "").strip()
    job_title = row.get("Job title", "").strip()
    date_applied = row.get("date applied", "").strip().lower()
    
    # Skip if missing required fields
    if not company or not job_title:
        return True
    
    # Skip if explicitly marked as not applied
    skip_indicators = ["did not apply", "to apply", "jen to apply", "need to finish"]
    if date_applied in skip_indicators:
        return True
    
    return False


def import_csv(file_path: str, dry_run: bool = False) -> None:
    """Import jobs from CSV file."""
    repository = JobRepository()
    
    if not os.path.exists(file_path):
        print(f"Error: File not found: {file_path}")
        sys.exit(1)
    
    imported = 0
    skipped = 0
    errors = []
    
    with open(file_path, "r", encoding="utf-8") as f:
        reader = csv.DictReader(f)
        
        for row_num, row in enumerate(reader, start=2):
            try:
                if should_skip_row(row):
                    skipped += 1
                    continue
                
                company = row.get("Description", "").strip()
                job_title = row.get("Job title", "").strip()
                date_applied_str = row.get("date applied", "").strip()
                job_url = row.get("link", "").strip() or None
                notes = row.get("note", "").strip() or None
                source = row.get("job board", "").strip() or "manual"
                
                # Combine notes with other relevant info
                note_parts = []
                if notes:
                    note_parts.append(notes)
                if row.get("Resume", "").strip():
                    note_parts.append(f"Resume: {row.get('Resume', '').strip()}")
                combined_notes = " | ".join(note_parts) if note_parts else None
                
                job = AppliedJob(
                    company=company,
                    job_title=job_title,
                    job_url=job_url,
                    location=None,
                    salary_range=None,
                    notes=combined_notes,
                    status="applied",
                    source=source if source else "manual",
                )
                
                # Parse application date
                if date_applied_str:
                    parsed_date = parse_date(date_applied_str)
                    if parsed_date:
                        job.application_date = parsed_date
                
                if not dry_run:
                    repository.create(job)
                
                imported += 1
                print(f"✓ Row {row_num}: {company} - {job_title}")
            
            except Exception as e:
                skipped += 1
                error_msg = f"Row {row_num}: {str(e)}"
                errors.append(error_msg)
                print(f"✗ {error_msg}")
    
    print(f"\n{'[DRY RUN] ' if dry_run else ''}Import complete:")
    print(f"  Imported: {imported}")
    print(f"  Skipped: {skipped}")
    
    if errors:
        print(f"\nErrors ({len(errors)}):")
        for error in errors[:10]:
            print(f"  - {error}")
        if len(errors) > 10:
            print(f"  ... and {len(errors) - 10} more")


if __name__ == "__main__":
    import argparse
    
    parser = argparse.ArgumentParser(description="Import jobs from CSV file")
    parser.add_argument("csv_file", help="Path to CSV file")
    parser.add_argument("--dry-run", action="store_true", help="Preview import without saving")
    
    args = parser.parse_args()
    
    print(f"{'[DRY RUN] ' if args.dry_run else ''}Importing jobs from: {args.csv_file}")
    print("-" * 80)
    
    import_csv(args.csv_file, dry_run=args.dry_run)

