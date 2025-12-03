"""Repository layer for preferences data access."""

from pydantic import BaseModel


# Pydantic models

class UpdateUserPreferencesRequest(BaseModel):
    theme: str | None = None
    timezone: str | None = None


class UpdateTenantPreferencesRequest(BaseModel):
    theme: str
