"""
Google Sheets service for tracking applied jobs.

Reads the job application tracker sheet to prevent duplicate recommendations.
"""

import logging
import os
import traceback
from typing import List, Set, Optional, Dict, Any
import json

try:
    import gspread
    from google.oauth2.service_account import Credentials
except ImportError:
    gspread = None
    Credentials = None

logger = logging.getLogger(__name__)

SCOPES = ["https://www.googleapis.com/auth/spreadsheets.readonly"]
DEFAULT_SHEET_ID = "1booMeM9GMAbrF6DgV9hsKo5_dxwYcD5kr049kAULAIA"


def _get_column_value(record: Dict[str, Any], possible_names: List[str]) -> str:
    """Get value from record using any of the possible column names (case-insensitive)."""
    for name in possible_names:
        for key, value in record.items():
            if key.lower() == name.lower():
                return str(value).strip() if value else ""
    return ""


def _normalize_job_identifier(company: str, title: str) -> str:
    """Normalize job identifier for comparison."""
    return f"{company.lower().strip()}|{title.lower().strip()}"


def _is_pending_application(date_applied: str) -> bool:
    """Check if date indicates pending application."""
    if not date_applied:
        return False
    date_lower = date_applied.lower().strip()
    # Only filter out if it's EXACTLY "to apply" or "did not apply" or starts with those phrases
    # Don't filter out dates that contain these words as part of a date
    if date_lower == "to apply" or date_lower == "did not apply":
        return True
    if date_lower.startswith("to apply") or date_lower.startswith("did not apply"):
        return True
    return False


def _create_credentials_from_json(credentials_json: str) -> Optional[Any]:
    """Create credentials from JSON string."""
    if not Credentials:
        return None
    
    try:
        credentials_info = json.loads(credentials_json)
        return Credentials.from_service_account_info(credentials_info, scopes=SCOPES)
    except json.JSONDecodeError as e:
        logger.error(f"GOOGLE_SHEETS_CREDENTIALS_JSON is not valid JSON: {e}")
        return None


def _create_credentials_from_file(credentials_path: str) -> Optional[Any]:
    """Create credentials from file path."""
    if not Credentials:
        return None
    
    if not os.path.exists(credentials_path):
        logger.warning(f"Credentials file not found at {credentials_path}")
        return None
    
    return Credentials.from_service_account_file(credentials_path, scopes=SCOPES)


def _initialize_client() -> Optional[Any]:
    """Initialize gspread client from environment variables or file."""
    if not gspread:
        logger.error("gspread not installed. Run: pip install gspread google-auth")
        return None
    
    credentials_json = os.getenv("GOOGLE_SHEETS_CREDENTIALS_JSON")
    if credentials_json:
        credentials = _create_credentials_from_json(credentials_json)
        if credentials:
            client = gspread.authorize(credentials)
            logger.info("✓ Google Sheets client initialized from environment variable")
            return client
    
    credentials_path = os.getenv("GOOGLE_SHEETS_CREDENTIALS_PATH")
    if not credentials_path:
        logger.warning("Neither GOOGLE_SHEETS_CREDENTIALS_JSON nor GOOGLE_SHEETS_CREDENTIALS_PATH set")
        return None
    
    credentials = _create_credentials_from_file(credentials_path)
    if not credentials:
        return None
    
    client = gspread.authorize(credentials)
    logger.info("✓ Google Sheets client initialized from credentials file")
    return client


def _get_worksheet(spreadsheet: Any) -> Any:
    """Get worksheet, preferring 'applied' tab."""
    try:
        worksheet = spreadsheet.worksheet("applied")
        logger.info("✓ Found 'applied' worksheet")
        return worksheet
    except Exception:
        logger.info("No 'applied' worksheet found, using first sheet")
        return spreadsheet.sheet1


def _parse_job_record(record: Dict[str, Any]) -> Optional[Dict[str, str]]:
    """Parse a single job record from sheet row."""
    company = _get_column_value(record, ["Description", "Company", "description", "company"])
    title = _get_column_value(record, ["Job title", "Job Title", "job title", "Title", "title"])
    
    # Require at least a title to count as a job application
    # Company can be empty if title exists
    if not title:
        return None
    
    date_applied = _get_column_value(record, ["date applied", "Date Applied", "date_applied", "Date", "date"])
    link = _get_column_value(record, ["link", "Link", "url", "URL", "job link", "Job Link"])
    
    if _is_pending_application(date_applied):
        logger.debug(f"Skipping row: {company or 'N/A'} - {title} (status: {date_applied})")
        return None
    
    return {
        "company": company if company else "Unknown Company",
        "title": title,
        "date_applied": date_applied if date_applied else "Not specified",
        "link": link if link else ""
    }


class AppliedJobsTracker:
    """Track jobs that have already been applied to via Google Sheets."""

    def __init__(self):
        """Initialize the tracker."""
        self.sheet_id = os.getenv("GOOGLE_SHEETS_ID", DEFAULT_SHEET_ID)
        self._client: Optional[Any] = None
        self._applied_jobs_cache: Set[str] = set()
        self._jobs_details_cache: List[Dict[str, str]] = []
        self._sheet_name_cache: Optional[str] = None

    def _get_client(self) -> Optional[Any]:
        """Get or create gspread client (lazy initialization)."""
        if self._client is not None:
            return self._client
        
        self._client = _initialize_client()
        return self._client

    def fetch_applied_jobs(self, refresh: bool = False) -> Set[str]:
        """Fetch list of companies/job titles from applied jobs sheet."""
        if self._applied_jobs_cache and not refresh:
            return self._applied_jobs_cache

        client = self._get_client()
        if not client:
            logger.warning("Google Sheets client unavailable - returning empty set")
            return set()

        try:
            spreadsheet = client.open_by_key(self.sheet_id)
            self._sheet_name_cache = spreadsheet.title
            logger.info(f"✓ Connected to Google Sheet: {self._sheet_name_cache}")
            
            sheet = _get_worksheet(spreadsheet)
            records = sheet.get_all_records()
            logger.info(f"✓ Found {len(records)} rows in sheet")
            
            if records:
                logger.info(f"✓ Column names: {list(records[0].keys())}")

            applied_jobs = set()
            jobs_details = []
            skipped_count = 0
            
            for record in records:
                job_data = _parse_job_record(record)
                if not job_data:
                    skipped_count += 1
                    continue
                
                identifier = _normalize_job_identifier(job_data["company"], job_data["title"])
                applied_jobs.add(identifier)
                jobs_details.append(job_data)

            logger.info(f"✓ Loaded {len(applied_jobs)} applied jobs from Google Sheets")
            logger.info(f"✓ Skipped {skipped_count} rows (empty or pending)")
            logger.info(f"✓ Total job details: {len(jobs_details)}")
            self._applied_jobs_cache = applied_jobs
            self._jobs_details_cache = jobs_details
            return applied_jobs

        except Exception as e:
            logger.error(f"Failed to fetch applied jobs from Google Sheets: {e}")
            logger.error(traceback.format_exc())
            return self._applied_jobs_cache
    
    def get_sheet_name(self) -> Optional[str]:
        """Get the name of the Google Sheet."""
        if self._sheet_name_cache:
            return self._sheet_name_cache
        
        client = self._get_client()
        if not client:
            return None
        
        try:
            spreadsheet = client.open_by_key(self.sheet_id)
            self._sheet_name_cache = spreadsheet.title
            return self._sheet_name_cache
        except Exception as e:
            logger.error(f"Failed to get sheet name: {e}")
            return None
    
    def get_jobs_details(self, refresh: bool = False) -> List[Dict[str, str]]:
        """Get detailed list of applied jobs."""
        if not self._jobs_details_cache or refresh:
            self.fetch_applied_jobs(refresh=refresh)
        return self._jobs_details_cache.copy()

    def is_job_applied(self, company: str, title: str) -> bool:
        """Check if a job has already been applied to."""
        if not self._applied_jobs_cache:
            self.fetch_applied_jobs()
        
        identifier = _normalize_job_identifier(company, title)
        return identifier in self._applied_jobs_cache

    def filter_jobs(self, jobs: List[Dict[str, str]]) -> List[Dict[str, str]]:
        """Filter out jobs that have already been applied to."""
        if not self._applied_jobs_cache:
            self.fetch_applied_jobs()

        if not self._applied_jobs_cache:
            logger.warning("No applied jobs loaded - returning all jobs")
            return jobs

        filtered = [
            job for job in jobs
            if not self.is_job_applied(job.get("company", ""), job.get("title", ""))
        ]
        
        filtered_count = len(jobs) - len(filtered)
        if filtered_count > 0:
            logger.info(f"Filtered {filtered_count} jobs, {len(filtered)} remaining")
        
        return filtered


_tracker_instance: Optional[AppliedJobsTracker] = None


def get_applied_jobs_tracker() -> AppliedJobsTracker:
    """Get or create the global applied jobs tracker instance."""
    global _tracker_instance
    if _tracker_instance is None:
        _tracker_instance = AppliedJobsTracker()
    return _tracker_instance
