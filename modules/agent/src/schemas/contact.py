"""Contact schemas for API validation."""

from typing import List, Optional

from pydantic import BaseModel

from lib.validators import validate_phone


class ContactCreate(BaseModel):
    first_name: str
    last_name: str
    email: Optional[str] = None
    phone: Optional[str] = None
    title: Optional[str] = None
    department: Optional[str] = None
    role: Optional[str] = None
    notes: Optional[str] = None
    individual_ids: Optional[List[int]] = None
    organization_ids: Optional[List[int]] = None
    is_primary: bool = False

    @classmethod
    def validate_phone_format(cls, v):
        return validate_phone(v) if v else v


class ContactUpdate(BaseModel):
    first_name: Optional[str] = None
    last_name: Optional[str] = None
    email: Optional[str] = None
    phone: Optional[str] = None
    title: Optional[str] = None
    department: Optional[str] = None
    role: Optional[str] = None
    notes: Optional[str] = None
    individual_ids: Optional[List[int]] = None
    organization_ids: Optional[List[int]] = None
    is_primary: Optional[bool] = None

    @classmethod
    def validate_phone_format(cls, v):
        return validate_phone(v) if v else v
