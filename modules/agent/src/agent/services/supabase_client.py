"""Supabase service for tracking applied jobs.

Replaces Google Sheets integration with Supabase database for better
data structure, querying capabilities, and real-time updates.
"""

import logging
import os
from typing import Dict, List, Optional, Set, Any

logger = logging.getLogger(__name__)

try:
    from supabase import create_client, Client
    SUPABASE_AVAILABLE = True
except ImportError:
    SUPABASE_AVAILABLE = False
    logger.warning("supabase-py not installed. Run: pip install supabase")


def _normalize_job_identifier(company: str, title: str) -> str:
    """Normalize job identifier for comparison."""
    return f"{company.lower().strip()}|{title.lower().strip()}"


def _get_supabase_client() -> Optional[Client]:
    """Create Supabase client from environment variables."""
    if not SUPABASE_AVAILABLE:
        return None

    supabase_url = os.getenv("SUPABASE_URL")
    supabase_key = os.getenv("SUPABASE_SERVICE_KEY")

    if not supabase_url or not supabase_key:
        logger.warning("SUPABASE_URL or SUPABASE_SERVICE_KEY not set in environment")
        return None

    return create_client(supabase_url, supabase_key)


def _map_db_record_to_job(record: Dict[str, Any]) -> Dict[str, str]:
    """Map database record to job dictionary."""
    return {
        "company": record.get("company", "Unknown Company"),
        "title": record.get("job_title", ""),
        "date_applied": record.get("application_date", "Not specified"),
        "link": record.get("job_url", ""),
        "status": record.get("status", "applied"),
        "location": record.get("location", ""),
        "salary_range": record.get("salary_range", ""),
        "notes": record.get("notes", ""),
        "source": record.get("source", "manual"),
        "id": record.get("id", "")
    }


class AppliedJobsTracker:
    """Track jobs that have already been applied to via Supabase."""

    def __init__(self):
        """Initialize the tracker."""
        self._client: Optional[Client] = None
        self._applied_jobs_cache: Set[str] = set()
        self._jobs_details_cache: List[Dict[str, str]] = []

    def _get_client(self) -> Optional[Client]:
        """Get or create Supabase client (lazy initialization)."""
        if self._client is not None:
            return self._client

        self._client = _get_supabase_client()
        if self._client:
            logger.info("✓ Supabase client initialized")
        return self._client

    def fetch_applied_jobs(self, refresh: bool = False) -> Set[str]:
        """Fetch list of companies/job titles from applied jobs table."""
        if self._applied_jobs_cache and not refresh:
            return self._applied_jobs_cache

        client = self._get_client()
        if not client:
            logger.warning("Supabase client unavailable - returning empty set")
            return set()

        try:
            response = client.table("applied_jobs").select("*").execute()
            records = response.data

            logger.info(f"✓ Found {len(records)} rows in applied_jobs table")

            applied_jobs = set()
            jobs_details = []

            for record in records:
                job_data = _map_db_record_to_job(record)

                if not job_data["title"]:
                    continue

                identifier = _normalize_job_identifier(job_data["company"], job_data["title"])
                applied_jobs.add(identifier)
                jobs_details.append(job_data)

            logger.info(f"✓ Loaded {len(applied_jobs)} applied jobs from Supabase")
            logger.info(f"✓ Total job details: {len(jobs_details)}")
            self._applied_jobs_cache = applied_jobs
            self._jobs_details_cache = jobs_details
            return applied_jobs

        except Exception as e:
            logger.error(f"Failed to fetch applied jobs from Supabase: {e}")
            return self._applied_jobs_cache

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

    def add_applied_job(
        self,
        company: str,
        job_title: str,
        job_url: str = "",
        location: str = "",
        salary_range: str = "",
        notes: str = "",
        status: str = "applied",
        source: str = "agent"
    ) -> Optional[Dict[str, Any]]:
        """Add a new job application to the database."""
        client = self._get_client()
        if not client:
            logger.error("Cannot add job - Supabase client unavailable")
            return None

        try:
            data = {
                "company": company,
                "job_title": job_title,
                "job_url": job_url,
                "location": location,
                "salary_range": salary_range,
                "notes": notes,
                "status": status,
                "source": source
            }

            response = client.table("applied_jobs").insert(data).execute()

            if response.data:
                logger.info(f"✓ Added job application: {company} - {job_title}")
                # Invalidate cache to force refresh on next fetch
                self._applied_jobs_cache = set()
                self._jobs_details_cache = []
                return response.data[0]

            return None

        except Exception as e:
            logger.error(f"Failed to add job application: {e}")
            return None


_tracker_instance: Optional[AppliedJobsTracker] = None


def get_applied_jobs_tracker() -> AppliedJobsTracker:
    """Get or create the global applied jobs tracker instance."""
    global _tracker_instance
    if _tracker_instance is None:
        _tracker_instance = AppliedJobsTracker()
    return _tracker_instance
