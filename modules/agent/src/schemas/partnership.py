"""Partnership schemas for API validation."""

from typing import List, Optional

from pydantic import BaseModel


class PartnershipCreate(BaseModel):
    account_type: str = "Organization"
    account: Optional[str] = None
    contacts: Optional[List[dict]] = None
    industry: Optional[List[str]] = None
    industry_ids: Optional[List[int]] = None
    partnership_name: str
    partnership_type: Optional[str] = None
    start_date: Optional[str] = None
    end_date: Optional[str] = None
    description: Optional[str] = None
    status: str = "Exploring"
    source: str = "manual"
    project_ids: Optional[List[int]] = None


class PartnershipUpdate(BaseModel):
    account: Optional[str] = None
    contacts: Optional[List[dict]] = None
    partnership_name: Optional[str] = None
    partnership_type: Optional[str] = None
    start_date: Optional[str] = None
    end_date: Optional[str] = None
    description: Optional[str] = None
    status: Optional[str] = None
    source: Optional[str] = None
    project_ids: Optional[List[int]] = None
