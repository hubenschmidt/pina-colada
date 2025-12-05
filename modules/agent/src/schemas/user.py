"""User schemas for API validation."""

from typing import Optional

from pydantic import BaseModel


class SetSelectedProjectRequest(BaseModel):
    project_id: Optional[int] = None
