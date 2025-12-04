"""Tenant schemas for API validation."""

from typing import Optional

from pydantic import BaseModel


class TenantCreate(BaseModel):
    """Model for tenant creation."""
    name: str
    slug: Optional[str] = None
    plan: str = "free"
