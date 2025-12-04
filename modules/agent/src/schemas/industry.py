"""Industry schemas for API validation."""

from pydantic import BaseModel


class IndustryCreate(BaseModel):
    name: str
