"""Opportunity schemas for API validation."""

from typing import List, Optional

from pydantic import BaseModel, Field


class OpportunityCreate(BaseModel):
    """Model for creating an opportunity."""
    account_type: str = "Organization"
    account: Optional[str] = None
    contacts: Optional[List[dict]] = None
    industry: Optional[List[str]] = None
    industry_ids: Optional[List[int]] = None
    title: str
    opportunity_name: str
    estimated_value: Optional[float] = None
    probability: Optional[int] = Field(default=None, ge=0, le=100)
    expected_close_date: Optional[str] = None
    description: Optional[str] = None
    status: str = "Qualifying"
    source: str = "manual"
    project_ids: Optional[List[int]] = None


class OpportunityUpdate(BaseModel):
    """Model for updating an opportunity."""
    account: Optional[str] = None
    contacts: Optional[List[dict]] = None
    title: Optional[str] = None
    opportunity_name: Optional[str] = None
    estimated_value: Optional[float] = None
    probability: Optional[int] = Field(default=None, ge=0, le=100)
    expected_close_date: Optional[str] = None
    description: Optional[str] = None
    status: Optional[str] = None
    source: Optional[str] = None
    project_ids: Optional[List[int]] = None
