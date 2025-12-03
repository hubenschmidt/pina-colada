"""Organization schemas for API validation."""

from datetime import datetime
from typing import List, Optional

from pydantic import BaseModel, field_validator

from lib.validators import validate_phone


class OrganizationCreate(BaseModel):
    name: str
    website: Optional[str] = None
    phone: Optional[str] = None
    employee_count: Optional[int] = None
    employee_count_range_id: Optional[int] = None
    funding_stage_id: Optional[int] = None
    description: Optional[str] = None
    account_id: Optional[int] = None
    industry_ids: Optional[List[int]] = None
    project_ids: Optional[List[int]] = None
    revenue_range_id: Optional[int] = None
    founding_year: Optional[int] = None
    headquarters_city: Optional[str] = None
    headquarters_state: Optional[str] = None
    headquarters_country: Optional[str] = None
    company_type: Optional[str] = None
    linkedin_url: Optional[str] = None
    crunchbase_url: Optional[str] = None

    @field_validator("phone")
    @classmethod
    def validate_phone_format(cls, v):
        return validate_phone(v)

    @field_validator("founding_year")
    @classmethod
    def validate_founding_year(cls, v):
        if v is None:
            return v
        current_year = datetime.now().year
        if v < 1800 or v > current_year:
            raise ValueError(f"founding_year must be between 1800 and {current_year}")
        return v


class OrganizationUpdate(BaseModel):
    name: Optional[str] = None
    website: Optional[str] = None
    phone: Optional[str] = None
    employee_count: Optional[int] = None
    employee_count_range_id: Optional[int] = None
    funding_stage_id: Optional[int] = None
    description: Optional[str] = None
    industry_ids: Optional[List[int]] = None
    project_ids: Optional[List[int]] = None
    revenue_range_id: Optional[int] = None
    founding_year: Optional[int] = None
    headquarters_city: Optional[str] = None
    headquarters_state: Optional[str] = None
    headquarters_country: Optional[str] = None
    company_type: Optional[str] = None
    linkedin_url: Optional[str] = None
    crunchbase_url: Optional[str] = None

    @field_validator("phone")
    @classmethod
    def validate_phone_format(cls, v):
        return validate_phone(v)

    @field_validator("founding_year")
    @classmethod
    def validate_founding_year(cls, v):
        if v is None:
            return v
        current_year = datetime.now().year
        if v < 1800 or v > current_year:
            raise ValueError(f"founding_year must be between 1800 and {current_year}")
        return v


class OrgContactCreate(BaseModel):
    individual_id: Optional[int] = None
    first_name: str
    last_name: str
    title: Optional[str] = None
    department: Optional[str] = None
    role: Optional[str] = None
    email: Optional[str] = None
    phone: Optional[str] = None
    is_primary: bool = False
    notes: Optional[str] = None

    @field_validator("phone")
    @classmethod
    def validate_phone_format(cls, v):
        return validate_phone(v)


class OrgContactUpdate(BaseModel):
    first_name: Optional[str] = None
    last_name: Optional[str] = None
    title: Optional[str] = None
    department: Optional[str] = None
    role: Optional[str] = None
    email: Optional[str] = None
    phone: Optional[str] = None
    is_primary: Optional[bool] = None
    notes: Optional[str] = None

    @field_validator("phone")
    @classmethod
    def validate_phone_format(cls, v):
        return validate_phone(v)


class OrgTechnologyCreate(BaseModel):
    technology_id: int
    source: Optional[str] = None
    confidence: Optional[float] = None


class FundingRoundCreate(BaseModel):
    round_type: str
    amount: Optional[int] = None
    announced_date: Optional[str] = None
    lead_investor: Optional[str] = None
    source_url: Optional[str] = None


class SignalCreate(BaseModel):
    signal_type: str
    headline: str
    description: Optional[str] = None
    signal_date: Optional[str] = None
    source: Optional[str] = None
    source_url: Optional[str] = None
    sentiment: Optional[str] = None
    relevance_score: Optional[float] = None
