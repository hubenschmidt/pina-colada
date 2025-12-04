"""Technology schemas for API validation."""

from typing import Optional

from pydantic import BaseModel


class TechnologyCreate(BaseModel):
    name: str
    category: str
    vendor: Optional[str] = None
