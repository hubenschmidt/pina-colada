"""Shared base for SQLAlchemy models."""

from sqlalchemy.orm import declarative_base

# Shared declarative base for all models
Base = declarative_base()

# Export all models for convenience
from models.Tenant import Tenant, TenantData, TenantCreateData, TenantUpdateData
from models.User import User, UserData, UserCreateData, UserUpdateData
from models.Role import Role, RoleData, RoleCreateData, RoleUpdateData
from models.UserRole import UserRole, UserRoleData, UserRoleCreateData
from models.Status import Status, StatusData, StatusCreateData, StatusUpdateData
from models.Organization import Organization, OrganizationData, OrganizationCreateData, OrganizationUpdateData
from models.Individual import Individual, IndividualData, IndividualCreateData, IndividualUpdateData
from models.Contact import Contact, ContactData, ContactCreateData, ContactUpdateData
from models.Project import Project, ProjectData, ProjectCreateData, ProjectUpdateData
from models.Deal import Deal, DealData, DealCreateData, DealUpdateData
from models.Lead import Lead, LeadData, LeadCreateData, LeadUpdateData
from models.Job import Job, JobData, JobCreateData, JobUpdateData
from models.Opportunity import Opportunity, OpportunityData, OpportunityCreateData, OpportunityUpdateData
from models.Partnership import Partnership, PartnershipData, PartnershipCreateData, PartnershipUpdateData
from models.Task import Task, TaskData, TaskCreateData, TaskUpdateData

__all__ = [
    "Base",
    # Tenant
    "Tenant",
    "TenantData",
    "TenantCreateData",
    "TenantUpdateData",
    # User
    "User",
    "UserData",
    "UserCreateData",
    "UserUpdateData",
    # Role
    "Role",
    "RoleData",
    "RoleCreateData",
    "RoleUpdateData",
    # UserRole
    "UserRole",
    "UserRoleData",
    "UserRoleCreateData",
    # Status
    "Status",
    "StatusData",
    "StatusCreateData",
    "StatusUpdateData",
    # Organization
    "Organization",
    "OrganizationData",
    "OrganizationCreateData",
    "OrganizationUpdateData",
    # Individual
    "Individual",
    "IndividualData",
    "IndividualCreateData",
    "IndividualUpdateData",
    # Contact
    "Contact",
    "ContactData",
    "ContactCreateData",
    "ContactUpdateData",
    # Project
    "Project",
    "ProjectData",
    "ProjectCreateData",
    "ProjectUpdateData",
    # Deal
    "Deal",
    "DealData",
    "DealCreateData",
    "DealUpdateData",
    # Lead
    "Lead",
    "LeadData",
    "LeadCreateData",
    "LeadUpdateData",
    # Job (extends Lead)
    "Job",
    "JobData",
    "JobCreateData",
    "JobUpdateData",
    # Opportunity (extends Lead)
    "Opportunity",
    "OpportunityData",
    "OpportunityCreateData",
    "OpportunityUpdateData",
    # Partnership (extends Lead)
    "Partnership",
    "PartnershipData",
    "PartnershipCreateData",
    "PartnershipUpdateData",
    # Task
    "Task",
    "TaskData",
    "TaskCreateData",
    "TaskUpdateData",
]
