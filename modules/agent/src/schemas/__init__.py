"""Pydantic schemas for API request/response validation."""

from schemas.comment import CommentCreate, CommentUpdate
from schemas.contact import ContactCreate, ContactUpdate
from schemas.document import DocumentUpdate, EntityLink
from schemas.individual import IndContactCreate, IndContactUpdate, IndividualCreate, IndividualUpdate
from schemas.industry import IndustryCreate
from schemas.job import JobCreate, JobUpdate
from schemas.note import NoteCreate, NoteUpdate
from schemas.notification import MarkEntityReadRequest, MarkReadRequest
from schemas.opportunity import OpportunityCreate, OpportunityUpdate
from schemas.organization import (
    FundingRoundCreate,
    OrgContactCreate,
    OrgContactUpdate,
    OrgTechnologyCreate,
    OrganizationCreate,
    OrganizationUpdate,
    SignalCreate,
)
from schemas.partnership import PartnershipCreate, PartnershipUpdate
from schemas.preferences import UpdateTenantPreferencesRequest, UpdateUserPreferencesRequest
from schemas.project import ProjectCreate, ProjectUpdate
from schemas.provenance import ProvenanceCreate
from schemas.report import Aggregation, ReportFilter, ReportQueryRequest, SavedReportCreate, SavedReportUpdate
from schemas.task import TaskCreate, TaskUpdate
from schemas.technology import TechnologyCreate
from schemas.tenant import TenantCreate
from schemas.user import SetSelectedProjectRequest

__all__ = [
    # Comment
    "CommentCreate",
    "CommentUpdate",
    # Contact
    "ContactCreate",
    "ContactUpdate",
    # Document
    "DocumentUpdate",
    "EntityLink",
    # Individual
    "IndividualCreate",
    "IndividualUpdate",
    "IndContactCreate",
    "IndContactUpdate",
    # Industry
    "IndustryCreate",
    # Job
    "JobCreate",
    "JobUpdate",
    # Note
    "NoteCreate",
    "NoteUpdate",
    # Notification
    "MarkReadRequest",
    "MarkEntityReadRequest",
    # Opportunity
    "OpportunityCreate",
    "OpportunityUpdate",
    # Organization
    "OrganizationCreate",
    "OrganizationUpdate",
    "OrgContactCreate",
    "OrgContactUpdate",
    "OrgTechnologyCreate",
    "FundingRoundCreate",
    "SignalCreate",
    # Partnership
    "PartnershipCreate",
    "PartnershipUpdate",
    # Preferences
    "UpdateUserPreferencesRequest",
    "UpdateTenantPreferencesRequest",
    # Project
    "ProjectCreate",
    "ProjectUpdate",
    # Provenance
    "ProvenanceCreate",
    # Report
    "ReportFilter",
    "Aggregation",
    "ReportQueryRequest",
    "SavedReportCreate",
    "SavedReportUpdate",
    # Task
    "TaskCreate",
    "TaskUpdate",
    # Technology
    "TechnologyCreate",
    # Tenant
    "TenantCreate",
    # User
    "SetSelectedProjectRequest",
]
