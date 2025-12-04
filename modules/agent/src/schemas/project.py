"""Project schemas for API validation."""

from typing import Optional

from pydantic import BaseModel


class ProjectCreate(BaseModel):
    name: str
    description: Optional[str] = None
    status: Optional[str] = None
    current_status_id: Optional[int] = None
    start_date: Optional[str] = None
    end_date: Optional[str] = None


class ProjectUpdate(BaseModel):
    name: Optional[str] = None
    description: Optional[str] = None
    status: Optional[str] = None
    current_status_id: Optional[int] = None
    start_date: Optional[str] = None
    end_date: Optional[str] = None
