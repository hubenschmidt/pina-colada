"""Service layer for job business logic."""

import logging
from typing import Dict, List, Optional
from agent.models.job import AppliedJob
from agent.repositories.job_repository import JobRepository

logger = logging.getLogger(__name__)


def _normalize_identifier(company: str, title: str) -> str:
    """Normalize job identifier for comparison."""
    return f"{company.lower().strip()}|{title.lower().strip()}"


def _map_to_dict(job: AppliedJob) -> Dict[str, str]:
    """Map database model to dictionary."""
    return {
        "company": job.company or "Unknown Company",
        "title": job.job_title or "",
        "date_applied": str(job.application_date) if job.application_date else "Not specified",
        "link": job.job_url or "",
        "status": job.status or "applied",
        "location": job.location or "",
        "salary_range": job.salary_range or "",
        "notes": job.notes or "",
        "source": job.source or "manual",
        "id": str(job.id) if job.id else "",
    }


class JobService:
    """Service for job business logic."""
    
    def __init__(self, repository: JobRepository):
        """Initialize service with repository."""
        self._repository = repository
        self._cache: Optional[set[str]] = None
        self._details_cache: Optional[List[Dict[str, str]]] = None
    
    def get_all_jobs(self, refresh: bool = False) -> List[Dict[str, str]]:
        """Get all jobs as dictionaries."""
        if self._details_cache and not refresh:
            return self._details_cache.copy()
        
        jobs = self._repository.find_all()
        details = [_map_to_dict(job) for job in jobs]
        
        self._details_cache = details
        self._cache = {_normalize_identifier(j["company"], j["title"]) for j in details if j["title"]}
        
        logger.info(f"Loaded {len(details)} jobs from repository")
        return details
    
    def fetch_applied_jobs(self, refresh: bool = False) -> set[str]:
        """Fetch set of normalized job identifiers (for compatibility)."""
        return self.get_applied_identifiers(refresh=refresh)
    
    def get_applied_identifiers(self, refresh: bool = False) -> set[str]:
        """Get set of normalized job identifiers."""
        if self._cache and not refresh:
            return self._cache.copy()
        
        self.get_all_jobs(refresh=refresh)
        return self._cache.copy() if self._cache else set()
    
    def get_jobs_details(self, refresh: bool = False) -> List[Dict[str, str]]:
        """Get detailed list of applied jobs (for compatibility)."""
        return self.get_all_jobs(refresh=refresh)
    
    def is_job_applied(self, company: str, title: str) -> bool:
        """Check if a job has been applied to."""
        identifiers = self.get_applied_identifiers()
        identifier = _normalize_identifier(company, title)
        return identifier in identifiers
    
    def filter_jobs(self, jobs: List[Dict[str, str]]) -> List[Dict[str, str]]:
        """Filter out jobs that have already been applied to."""
        identifiers = self.get_applied_identifiers()
        
        if not identifiers:
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
    ) -> Optional[Dict[str, str]]:
        """Add a new job application (for compatibility)."""
        return self.add_job(
            company=company,
            job_title=job_title,
            job_url=job_url,
            location=location,
            salary_range=salary_range,
            notes=notes,
            status=status,
            source=source
        )
    
    def add_job(
        self,
        company: str,
        job_title: str,
        job_url: str = "",
        location: str = "",
        salary_range: str = "",
        notes: str = "",
        status: str = "applied",
        source: str = "agent"
    ) -> Optional[Dict[str, str]]:
        """Add a new job application."""
        job = AppliedJob(
            company=company,
            job_title=job_title,
            job_url=job_url,
            location=location,
            salary_range=salary_range,
            notes=notes,
            status=status,
            source=source
        )
        
        created = self._repository.create(job)
        logger.info(f"Added job application: {company} - {job_title}")
        
        self._cache = None
        self._details_cache = None
        
        return _map_to_dict(created)

