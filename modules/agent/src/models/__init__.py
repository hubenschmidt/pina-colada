"""Shared base for SQLAlchemy models."""

from sqlalchemy.orm import declarative_base

# Shared declarative base for all models
Base = declarative_base()

# Export all models for convenience
from models.Tenant import Tenant
from models.User import User
from models.Role import Role
from models.UserRole import UserRole
from models.Status import Status
from models.Organization import Organization
from models.Individual import Individual
from models.Contact import Contact
from models.Project import Project
from models.Deal import Deal
from models.Lead import Lead
from models.Job import Job
from models.Opportunity import Opportunity
from models.Partnership import Partnership
from models.Task import Task
from models.Asset import Asset
from models.Tag import Tag
from models.Account import Account
from models.Industry import Industry
from models.Note import Note

__all__ = [
    "Base",
    "Tenant",
    "User",
    "Role",
    "UserRole",
    "Status",
    "Organization",
    "Individual",
    "Contact",
    "Project",
    "Deal",
    "Lead",
    "Job",
    "Opportunity",
    "Partnership",
    "Task",
    "Asset",
    "Tag",
    "Account",
    "Industry",
    "Note",
]
