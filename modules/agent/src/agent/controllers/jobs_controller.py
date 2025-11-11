"""Controller for jobs API endpoints."""

import logging
from typing import List, Optional
from fastapi import APIRouter, HTTPException
from pydantic import BaseModel
from agent.repositories.job_repository import JobRepository

logger = logging.getLogger(__name__)

router = APIRouter(prefix="/api/jobs", tags=["jobs"])


class JobCreate(BaseModel):
    """Job creation request model."""
    company: str
    job_title: str
    job_url: Optional[str] = None
    location: Optional[str] = None
    salary_range: Optional[str] = None
    notes: Optional[str] = None
    status: str = "applied"
    source: str = "manual"


class JobUpdate(BaseModel):
    """Job update request model."""
    company: Optional[str] = None
    job_title: Optional[str] = None
    job_url: Optional[str] = None
    location: Optional[str] = None
    salary_range: Optional[str] = None
    notes: Optional[str] = None
    status: Optional[str] = None
    source: Optional[str] = None


class JobResponse(BaseModel):
    """Job response model."""
    id: str
    company: str
    job_title: str
    application_date: str
    status: str
    job_url: Optional[str]
    location: Optional[str]
    salary_range: Optional[str]
    notes: Optional[str]
    source: str
    created_at: str
    updated_at: str


@router.get("/", response_model=List[JobResponse])
async def get_jobs():
    """Get all jobs."""
    repository = JobRepository()
    jobs = repository.find_all()
    
    return [
        JobResponse(
            id=str(job.id),
            company=job.company,
            job_title=job.job_title,
            application_date=job.application_date.isoformat() if job.application_date else "",
            status=job.status or "applied",
            job_url=job.job_url,
            location=job.location,
            salary_range=job.salary_range,
            notes=job.notes,
            source=job.source or "manual",
            created_at=job.created_at.isoformat() if job.created_at else "",
            updated_at=job.updated_at.isoformat() if job.updated_at else ""
        )
        for job in jobs
    ]


@router.post("/", response_model=JobResponse)
async def create_job(job_data: JobCreate):
    """Create a new job."""
    from agent.models.job import AppliedJob
    
    job = AppliedJob(
        company=job_data.company,
        job_title=job_data.job_title,
        job_url=job_data.job_url,
        location=job_data.location,
        salary_range=job_data.salary_range,
        notes=job_data.notes,
        status=job_data.status,
        source=job_data.source
    )
    
    repository = JobRepository()
    created = repository.create(job)
    
    return JobResponse(
        id=str(created.id),
        company=created.company,
        job_title=created.job_title,
        application_date=created.application_date.isoformat() if created.application_date else "",
        status=created.status or "applied",
        job_url=created.job_url,
        location=created.location,
        salary_range=created.salary_range,
        notes=created.notes,
        source=created.source or "manual",
        created_at=created.created_at.isoformat() if created.created_at else "",
        updated_at=created.updated_at.isoformat() if created.updated_at else ""
    )


@router.get("/{job_id}", response_model=JobResponse)
async def get_job(job_id: str):
    """Get a job by ID."""
    repository = JobRepository()
    job = repository.find_by_id(job_id)
    
    if not job:
        raise HTTPException(status_code=404, detail="Job not found")
    
    return JobResponse(
        id=str(job.id),
        company=job.company,
        job_title=job.job_title,
        application_date=job.application_date.isoformat() if job.application_date else "",
        status=job.status or "applied",
        job_url=job.job_url,
        location=job.location,
        salary_range=job.salary_range,
        notes=job.notes,
        source=job.source or "manual",
        created_at=job.created_at.isoformat() if job.created_at else "",
        updated_at=job.updated_at.isoformat() if job.updated_at else ""
    )


@router.put("/{job_id}", response_model=JobResponse)
async def update_job(job_id: str, job_data: JobUpdate):
    """Update a job."""
    repository = JobRepository()
    job = repository.find_by_id(job_id)
    
    if not job:
        raise HTTPException(status_code=404, detail="Job not found")
    
    if job_data.company is not None:
        job.company = job_data.company
    if job_data.job_title is not None:
        job.job_title = job_data.job_title
    if job_data.job_url is not None:
        job.job_url = job_data.job_url
    if job_data.location is not None:
        job.location = job_data.location
    if job_data.salary_range is not None:
        job.salary_range = job_data.salary_range
    if job_data.notes is not None:
        job.notes = job_data.notes
    if job_data.status is not None:
        job.status = job_data.status
    if job_data.source is not None:
        job.source = job_data.source
    
    updated = repository.update(job)
    
    return JobResponse(
        id=str(updated.id),
        company=updated.company,
        job_title=updated.job_title,
        application_date=updated.application_date.isoformat() if updated.application_date else "",
        status=updated.status or "applied",
        job_url=updated.job_url,
        location=updated.location,
        salary_range=updated.salary_range,
        notes=updated.notes,
        source=updated.source or "manual",
        created_at=updated.created_at.isoformat() if updated.created_at else "",
        updated_at=updated.updated_at.isoformat() if updated.updated_at else ""
    )


@router.delete("/{job_id}")
async def delete_job(job_id: str):
    """Delete a job."""
    repository = JobRepository()
    deleted = repository.delete(job_id)
    
    if not deleted:
        raise HTTPException(status_code=404, detail="Job not found")
    
    return {"success": True}

