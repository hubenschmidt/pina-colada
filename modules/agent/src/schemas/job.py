"""Job schemas for API validation."""

from typing import List, Optional

from pydantic import BaseModel


class JobCreate(BaseModel):
    account_type: str = "Organization"
    account: Optional[str] = None
    contact_name: Optional[str] = None
    contacts: Optional[List[dict]] = None
    industry: Optional[List[str]] = None
    industry_ids: Optional[List[int]] = None
    job_title: str
    date: Optional[str] = None
    job_url: Optional[str] = None
    salary_range: Optional[str] = None  # Legacy, kept for backwards compat
    salary_range_id: Optional[int] = None
    description: Optional[str] = None
    resume: Optional[str] = None
    status: str = "applied"
    source: str = "manual"
    project_ids: Optional[List[int]] = None


class JobUpdate(BaseModel):
    account: Optional[str] = None
    contacts: Optional[List[dict]] = None
    job_title: Optional[str] = None
    date: Optional[str] = None
    job_url: Optional[str] = None
    salary_range: Optional[str] = None  # Legacy, kept for backwards compat
    salary_range_id: Optional[int] = None
    description: Optional[str] = None
    resume: Optional[str] = None
    status: Optional[str] = None
    source: Optional[str] = None
    lead_status_id: Optional[str] = None
    project_ids: Optional[List[int]] = None
