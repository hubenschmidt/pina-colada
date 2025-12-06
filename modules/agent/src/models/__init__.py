"""Shared base for SQLAlchemy models."""

from sqlalchemy.orm import declarative_base

# Shared declarative base for all models
Base = declarative_base()

# Export all models for convenience
# Import order matters for SQLAlchemy relationship resolution
# Models referenced by relationships must be imported before the referencing model
from models.TenantPreferences import TenantPreferences
from models.UserPreferences import UserPreferences
from models.Comment import Comment
from models.Tenant import Tenant
from models.User import User
from models.Role import Role
from models.UserRole import UserRole
from models.Status import Status
from models.Account import Account
from models.Organization import Organization
from models.Individual import Individual
from models.IndividualRelationship import IndividualRelationship
from models.OrganizationRelationship import OrganizationRelationship
from models.Contact import Contact, ContactAccount
from models.Project import Project
from models.Deal import Deal
from models.Lead import Lead
from models.Job import Job
from models.Opportunity import Opportunity
from models.Partnership import Partnership
from models.Task import Task
from models.Asset import Asset
from models.Document import Document
from models.EntityAsset import EntityAsset
from models.Tag import Tag
from models.AccountProject import AccountProject
from models.Industry import Industry
from models.Note import Note
from models.SalaryRange import SalaryRange
from models.EmployeeCountRange import EmployeeCountRange
from models.FundingStage import FundingStage
from models.RevenueRange import RevenueRange
from models.Technology import Technology
from models.OrganizationTechnology import OrganizationTechnology
from models.FundingRound import FundingRound
from models.CompanySignal import CompanySignal
from models.DataProvenance import DataProvenance
from models.SavedReport import SavedReport
from models.SavedReportProject import SavedReportProject
from models.LeadProject import LeadProject
from models.CommentNotification import CommentNotification
from models.AccountRelationship import AccountRelationship
from models.Reasoning import Reasoning

__all__ = [
    "Base",
    "TenantPreferences",
    "UserPreferences",
    "Comment",
    "Tenant",
    "User",
    "Role",
    "UserRole",
    "Status",
    "Account",
    "Organization",
    "Individual",
    "IndividualRelationship",
    "OrganizationRelationship",
    "Contact",
    "ContactAccount",
    "Project",
    "Deal",
    "Lead",
    "Job",
    "Opportunity",
    "Partnership",
    "Task",
    "Asset",
    "Document",
    "EntityAsset",
    "Tag",
    "AccountProject",
    "Industry",
    "Note",
    "SalaryRange",
    "EmployeeCountRange",
    "FundingStage",
    "RevenueRange",
    "Technology",
    "OrganizationTechnology",
    "FundingRound",
    "CompanySignal",
    "DataProvenance",
    "SavedReport",
    "SavedReportProject",
    "LeadProject",
    "CommentNotification",
    "AccountRelationship",
    "Reasoning",
]
