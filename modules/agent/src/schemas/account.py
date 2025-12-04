"""Account schemas for API validation."""

from typing import Optional

from pydantic import BaseModel


class AccountRelationshipCreate(BaseModel):
    to_account_id: int
    relationship_type: Optional[str] = None
    notes: Optional[str] = None
