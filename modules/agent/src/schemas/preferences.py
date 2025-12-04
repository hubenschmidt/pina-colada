"""Preferences schemas for API validation."""

from pydantic import BaseModel


class UpdateUserPreferencesRequest(BaseModel):
    theme: str | None = None
    timezone: str | None = None


class UpdateTenantPreferencesRequest(BaseModel):
    theme: str
