"""Individual schemas for API validation."""

from typing import List, Optional

from pydantic import BaseModel, field_validator

from lib.validators import validate_phone


class IndividualCreate(BaseModel):
    first_name: str
    last_name: str
    email: Optional[str] = None
    phone: Optional[str] = None
    linkedin_url: Optional[str] = None
    title: Optional[str] = None
    description: Optional[str] = None
    account_id: Optional[int] = None
    industry_ids: Optional[List[int]] = None
    project_ids: Optional[List[int]] = None
    twitter_url: Optional[str] = None
    github_url: Optional[str] = None
    bio: Optional[str] = None
    seniority_level: Optional[str] = None
    department: Optional[str] = None
    is_decision_maker: Optional[bool] = None
    reports_to_id: Optional[int] = None

    @field_validator("phone")
    @classmethod
    def validate_phone_format(cls, v):
        return validate_phone(v)


class IndividualUpdate(BaseModel):
    first_name: Optional[str] = None
    last_name: Optional[str] = None
    email: Optional[str] = None
    phone: Optional[str] = None
    linkedin_url: Optional[str] = None
    title: Optional[str] = None
    description: Optional[str] = None
    industry_ids: Optional[List[int]] = None
    project_ids: Optional[List[int]] = None
    twitter_url: Optional[str] = None
    github_url: Optional[str] = None
    bio: Optional[str] = None
    seniority_level: Optional[str] = None
    department: Optional[str] = None
    is_decision_maker: Optional[bool] = None
    reports_to_id: Optional[int] = None

    @field_validator("phone")
    @classmethod
    def validate_phone_format(cls, v):
        return validate_phone(v)


class IndContactCreate(BaseModel):
    first_name: str
    last_name: str
    organization_id: Optional[int] = None
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


class IndContactUpdate(BaseModel):
    first_name: Optional[str] = None
    last_name: Optional[str] = None
    organization_id: Optional[int] = None
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
