"""Repository layer for project data access."""

from typing import Optional

from pydantic import BaseModel


# Pydantic models

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
