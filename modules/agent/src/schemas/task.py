"""Task schemas for API validation."""

from datetime import date
from decimal import Decimal
from typing import Optional

from pydantic import BaseModel, field_validator


FIBONACCI_VALUES = (1, 2, 3, 5, 8, 13, 21)


class TaskCreate(BaseModel):
    title: str
    description: Optional[str] = None
    taskable_type: Optional[str] = None
    taskable_id: Optional[int] = None
    current_status_id: Optional[int] = None
    priority_id: Optional[int] = None
    start_date: Optional[date] = None
    due_date: Optional[date] = None
    estimated_hours: Optional[Decimal] = None
    actual_hours: Optional[Decimal] = None
    complexity: Optional[int] = None
    sort_order: Optional[int] = None
    assigned_to_individual_id: Optional[int] = None

    @field_validator("complexity")
    @classmethod
    def validate_complexity(cls, v):
        if v is not None and v not in FIBONACCI_VALUES:
            raise ValueError(f"complexity must be one of {FIBONACCI_VALUES}")
        return v


class TaskUpdate(BaseModel):
    title: Optional[str] = None
    description: Optional[str] = None
    taskable_type: Optional[str] = None
    taskable_id: Optional[int] = None
    current_status_id: Optional[int] = None
    priority_id: Optional[int] = None
    start_date: Optional[date] = None
    due_date: Optional[date] = None
    estimated_hours: Optional[Decimal] = None
    actual_hours: Optional[Decimal] = None
    complexity: Optional[int] = None
    sort_order: Optional[int] = None
    completed_at: Optional[str] = None
    assigned_to_individual_id: Optional[int] = None

    @field_validator("complexity")
    @classmethod
    def validate_complexity(cls, v):
        if v is not None and v not in FIBONACCI_VALUES:
            raise ValueError(f"complexity must be one of {FIBONACCI_VALUES}")
        return v
