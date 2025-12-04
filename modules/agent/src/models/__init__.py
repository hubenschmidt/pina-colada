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
from models.Document import Document
from models.EntityAsset import EntityAsset
from models.Tag import Tag
from models.AccountProject import AccountProject
from models.Account import Account
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
from models.LeadProject import LeadProject
from models.CommentNotification import CommentNotification
from models.AccountRelationship import AccountRelationship

__all__ = [
    "Base",
    "Tenant",
    "User",
    "Role",
    "User_Role",
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
    "Document",
    "Entity_Asset",
    "Tag",
    "Account_Project",
    "Account",
    "Industry",
    "Note",
    "Salary_Range",
    "Employee_Count_Range",
    "Funding_Stage",
    "Revenue_Range",
    "Technology",
    "Organization_Technology",
    "Funding_Round",
    "Company_Signal",
    "Data_Provenance",
    "Saved_Report",
    "Lead_Project",
    "Comment_Notification",
    "AccountRelationship",
]
