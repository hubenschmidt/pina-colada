"""Report schemas for API validation."""

from typing import Any, List, Literal, Optional

from pydantic import BaseModel


class ReportFilter(BaseModel):
    field: str
    operator: Literal["eq", "neq", "gt", "gte", "lt", "lte", "contains", "starts_with", "is_null", "is_not_null", "in"]
    value: Any = None


class Aggregation(BaseModel):
    function: Literal["count", "sum", "avg", "min", "max"]
    field: str
    alias: str


class ReportQueryRequest(BaseModel):
    primary_entity: Literal["organizations", "individuals", "contacts", "leads", "notes"]
    columns: List[str]
    joins: Optional[List[str]] = None
    filters: Optional[List[ReportFilter]] = None
    group_by: Optional[List[str]] = None
    aggregations: Optional[List[Aggregation]] = None
    limit: int = 100
    offset: int = 0
    project_id: Optional[int] = None


class SavedReportCreate(BaseModel):
    name: str
    description: Optional[str] = None
    query_definition: dict
    project_ids: Optional[List[int]] = None


class SavedReportUpdate(BaseModel):
    name: Optional[str] = None
    description: Optional[str] = None
    query_definition: Optional[dict] = None
    project_ids: Optional[List[int]] = None
