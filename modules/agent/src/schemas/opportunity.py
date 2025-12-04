"""Opportunity schemas for API validation."""

from typing import List, Optional

from pydantic import BaseModel, Field, field_validator


def empty_str_to_none(v):
    """Convert empty strings to None for optional numeric fields."""
    if v == "" or v is None:
        return None
    return v


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

    @field_validator("estimated_value", "probability", mode="before")
    @classmethod
    def convert_empty_to_none(cls, v):
        return empty_str_to_none(v)


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

    @field_validator("estimated_value", "probability", mode="before")
    @classmethod
    def convert_empty_to_none(cls, v):
        return empty_str_to_none(v)
