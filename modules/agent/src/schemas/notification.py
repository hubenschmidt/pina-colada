"""Notification schemas for API validation."""

from typing import List

from pydantic import BaseModel


class MarkReadRequest(BaseModel):
    notification_ids: List[int]


class MarkEntityReadRequest(BaseModel):
    entity_type: str
    entity_id: int
