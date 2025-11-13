"""
DealTracker SQLAlchemy Models

Complete ORM implementation following KISS/YAGNI principles:
- Flexible Status system with ordered workflows (Status + join tables)
- All statuses/stages/priorities stored in database for flexibility
- Joined Table Inheritance (Lead → Job)
- Association objects for many-to-many relationships
- Polymorphic Task associations
- Application-level validation instead of complex DB triggers
- Helper functions for polymorphic queries

See dealtracker_schema.md for complete documentation.
"""

from datetime import datetime, date
from typing import List, Optional
from sqlalchemy import (
    ForeignKey,
    Numeric,
    Date,
    CheckConstraint,
    UniqueConstraint,
    func,
)
from sqlalchemy.orm import (
    DeclarativeBase,
    Mapped,
    mapped_column,
    relationship,
    validates,
)


class Base(DeclarativeBase):
    """Base class for all models"""
    pass


# ==============================
# Core Models
# ==============================


class Status(Base):
    """Status model - central status/stage definitions"""
    __tablename__ = "Status"

    id: Mapped[int] = mapped_column(primary_key=True)
    name: Mapped[str] = mapped_column(unique=True)
    description: Mapped[Optional[str]]
    category: Mapped[Optional[str]]  # 'job', 'lead', 'deal', 'task_status', 'task_priority'
    is_terminal: Mapped[bool] = mapped_column(default=False)
    created_at: Mapped[datetime] = mapped_column(server_default=func.now())
    updated_at: Mapped[datetime] = mapped_column(server_default=func.now(), onupdate=func.now())


class Tenant(Base):
    """
    Tenant model - represents the app customer/company (multi-tenant isolation)
    See auth_spec.md for full multi-tenancy implementation details
    """
    __tablename__ = "Tenant"

    id: Mapped[int] = mapped_column(primary_key=True)
    name: Mapped[str]
    slug: Mapped[str] = mapped_column(unique=True)  # URL-safe identifier
    status: Mapped[str] = mapped_column(default="active")  # active, suspended, trial, cancelled
    plan: Mapped[str] = mapped_column(default="free")  # free, starter, professional, enterprise
    settings: Mapped[Optional[dict]]  # JSONB tenant-specific settings
    created_at: Mapped[datetime] = mapped_column(server_default=func.now())
    updated_at: Mapped[datetime] = mapped_column(server_default=func.now(), onupdate=func.now())

    # Relationships
    users: Mapped[List["User"]] = relationship(back_populates="tenant", cascade="all, delete-orphan")


class User(Base):
    """
    User model - belongs to one Tenant
    See auth_spec.md for authentication and authorization details
    """
    __tablename__ = "User"

    id: Mapped[int] = mapped_column(primary_key=True)
    tenant_id: Mapped[int] = mapped_column(ForeignKey("Tenant.id", ondelete="CASCADE"))
    email: Mapped[str]
    first_name: Mapped[Optional[str]]
    last_name: Mapped[Optional[str]]
    avatar_url: Mapped[Optional[str]]
    status: Mapped[str] = mapped_column(default="active")  # active, inactive, invited
    last_login_at: Mapped[Optional[datetime]]
    created_at: Mapped[datetime] = mapped_column(server_default=func.now())
    updated_at: Mapped[datetime] = mapped_column(server_default=func.now(), onupdate=func.now())

    __table_args__ = (
        UniqueConstraint("tenant_id", "email", name="uq_user_tenant_email"),
    )

    # Relationships
    tenant: Mapped["Tenant"] = relationship(back_populates="users")
    user_roles: Mapped[List["UserRole"]] = relationship(back_populates="user", cascade="all, delete-orphan")


class Role(Base):
    """
    Role model - system-defined roles
    See auth_spec.md for role-based access control details
    """
    __tablename__ = "Role"

    id: Mapped[int] = mapped_column(primary_key=True)
    name: Mapped[str] = mapped_column(unique=True)  # owner, admin, member, viewer
    description: Mapped[Optional[str]]
    created_at: Mapped[datetime] = mapped_column(server_default=func.now())

    # Relationships
    user_roles: Mapped[List["UserRole"]] = relationship(back_populates="role", cascade="all, delete-orphan")


class UserRole(Base):
    """UserRole junction table - users can have multiple roles"""
    __tablename__ = "UserRole"

    user_id: Mapped[int] = mapped_column(ForeignKey("User.id", ondelete="CASCADE"), primary_key=True)
    role_id: Mapped[int] = mapped_column(ForeignKey("Role.id", ondelete="CASCADE"), primary_key=True)
    created_at: Mapped[datetime] = mapped_column(server_default=func.now())

    # Relationships
    user: Mapped["User"] = relationship(back_populates="user_roles")
    role: Mapped["Role"] = relationship(back_populates="user_roles")


class Project(Base):
    """Project model - top-level organizational unit"""
    __tablename__ = "Project"

    id: Mapped[int] = mapped_column(primary_key=True)
    tenant_id: Mapped[Optional[int]] = mapped_column(ForeignKey("Tenant.id", ondelete="CASCADE"))
    name: Mapped[str]
    description: Mapped[Optional[str]]
    owner_user_id: Mapped[Optional[int]]
    status: Mapped[Optional[str]]
    start_date: Mapped[Optional[date]] = mapped_column(Date)
    end_date: Mapped[Optional[date]] = mapped_column(Date)
    created_at: Mapped[datetime] = mapped_column(server_default=func.now())
    updated_at: Mapped[datetime] = mapped_column(server_default=func.now(), onupdate=func.now())

    # Relationships
    project_deals: Mapped[List["ProjectDeal"]] = relationship(back_populates="project", cascade="all, delete-orphan")
    project_leads: Mapped[List["ProjectLead"]] = relationship(back_populates="project", cascade="all, delete-orphan")
    project_contacts: Mapped[List["ProjectContact"]] = relationship(back_populates="project", cascade="all, delete-orphan")
    project_organizations: Mapped[List["ProjectOrganization"]] = relationship(back_populates="project", cascade="all, delete-orphan")
    project_individuals: Mapped[List["ProjectIndividual"]] = relationship(back_populates="project", cascade="all, delete-orphan")


class Deal(Base):
    """Deal model - represents a sales opportunity"""
    __tablename__ = "Deal"

    id: Mapped[int] = mapped_column(primary_key=True)
    tenant_id: Mapped[Optional[int]] = mapped_column(ForeignKey("Tenant.id", ondelete="CASCADE"))
    name: Mapped[str]
    description: Mapped[Optional[str]]
    owner_user_id: Mapped[Optional[int]]
    current_status_id: Mapped[Optional[int]] = mapped_column(ForeignKey("Status.id", ondelete="SET NULL"))
    value_amount: Mapped[Optional[float]] = mapped_column(Numeric(18, 2))
    value_currency: Mapped[str] = mapped_column(default="USD")
    probability: Mapped[Optional[float]] = mapped_column(Numeric(5, 2))
    expected_close_date: Mapped[Optional[date]] = mapped_column(Date)
    close_date: Mapped[Optional[date]] = mapped_column(Date)
    created_at: Mapped[datetime] = mapped_column(server_default=func.now())
    updated_at: Mapped[datetime] = mapped_column(server_default=func.now(), onupdate=func.now())

    # Relationships
    current_status: Mapped[Optional["Status"]] = relationship(foreign_keys=[current_status_id])
    leads: Mapped[List["Lead"]] = relationship(back_populates="deal", cascade="all, delete-orphan")
    project_deals: Mapped[List["ProjectDeal"]] = relationship(back_populates="deal", cascade="all, delete-orphan")


class Lead(Base):
    """
    Lead model - base table for Joined Table Inheritance

    Subclasses: Job, Opportunity, Partnership, etc.
    """
    __tablename__ = "Lead"

    id: Mapped[int] = mapped_column(primary_key=True)
    deal_id: Mapped[int] = mapped_column(ForeignKey("Deal.id", ondelete="CASCADE"))
    type: Mapped[str]  # Discriminator
    title: Mapped[str]
    description: Mapped[Optional[str]]
    source: Mapped[Optional[str]]
    current_status_id: Mapped[Optional[int]] = mapped_column(ForeignKey("Status.id", ondelete="SET NULL"))
    owner_user_id: Mapped[Optional[int]]
    created_at: Mapped[datetime] = mapped_column(server_default=func.now())
    updated_at: Mapped[datetime] = mapped_column(server_default=func.now(), onupdate=func.now())

    __mapper_args__ = {
        "polymorphic_identity": "lead",
        "polymorphic_on": "type",
    }

    # Relationships
    deal: Mapped["Deal"] = relationship(back_populates="leads")
    current_status: Mapped[Optional["Status"]] = relationship(foreign_keys=[current_status_id])
    lead_contacts: Mapped[List["LeadContact"]] = relationship(back_populates="lead", cascade="all, delete-orphan")
    lead_organizations: Mapped[List["LeadOrganization"]] = relationship(back_populates="lead", cascade="all, delete-orphan")
    lead_individuals: Mapped[List["LeadIndividual"]] = relationship(back_populates="lead", cascade="all, delete-orphan")
    project_leads: Mapped[List["ProjectLead"]] = relationship(back_populates="lead", cascade="all, delete-orphan")

    @validates('lead_organizations', 'lead_individuals')
    def validate_has_party(self, key, value):
        """Ensure Lead has at least one Individual or Organization"""
        # This will be called after the relationship is set
        # More complex validation should happen at commit time via events
        return value


class Job(Lead):
    """Job model - extends Lead for job applications"""
    __tablename__ = "Job"

    id: Mapped[int] = mapped_column(ForeignKey("Lead.id", ondelete="CASCADE"), primary_key=True)
    organization_id: Mapped[int] = mapped_column(ForeignKey("Organization.id", ondelete="CASCADE"))
    job_title: Mapped[str]
    job_url: Mapped[Optional[str]]
    notes: Mapped[Optional[str]]
    resume_date: Mapped[Optional[datetime]]
    salary_range: Mapped[Optional[str]]
    created_at: Mapped[datetime] = mapped_column(server_default=func.now())
    updated_at: Mapped[datetime] = mapped_column(server_default=func.now(), onupdate=func.now())

    __mapper_args__ = {
        "polymorphic_identity": "Job",
    }

    # Relationship
    organization: Mapped["Organization"] = relationship()


class JobStatus(Base):
    """JobStatus model - configures which statuses are valid for Jobs"""
    __tablename__ = "Job_Status"

    id: Mapped[int] = mapped_column(primary_key=True)
    status_id: Mapped[int] = mapped_column(ForeignKey("Status.id", ondelete="CASCADE"), unique=True)
    display_order: Mapped[int] = mapped_column(unique=True)
    is_default: Mapped[bool] = mapped_column(default=False)
    created_at: Mapped[datetime] = mapped_column(server_default=func.now())

    # Relationship
    status: Mapped["Status"] = relationship()


class LeadStatus(Base):
    """LeadStatus model - configures which statuses are valid for Leads"""
    __tablename__ = "Lead_Status"

    id: Mapped[int] = mapped_column(primary_key=True)
    status_id: Mapped[int] = mapped_column(ForeignKey("Status.id", ondelete="CASCADE"), unique=True)
    display_order: Mapped[int] = mapped_column(unique=True)
    is_default: Mapped[bool] = mapped_column(default=False)
    created_at: Mapped[datetime] = mapped_column(server_default=func.now())

    # Relationship
    status: Mapped["Status"] = relationship()


class DealStatus(Base):
    """DealStatus model - configures which statuses are valid for Deals"""
    __tablename__ = "Deal_Status"

    id: Mapped[int] = mapped_column(primary_key=True)
    status_id: Mapped[int] = mapped_column(ForeignKey("Status.id", ondelete="CASCADE"), unique=True)
    display_order: Mapped[int] = mapped_column(unique=True)
    is_default: Mapped[bool] = mapped_column(default=False)
    created_at: Mapped[datetime] = mapped_column(server_default=func.now())

    # Relationship
    status: Mapped["Status"] = relationship()


class TaskStatus(Base):
    """TaskStatus model - configures which statuses are valid for Tasks"""
    __tablename__ = "Task_Status"

    id: Mapped[int] = mapped_column(primary_key=True)
    status_id: Mapped[int] = mapped_column(ForeignKey("Status.id", ondelete="CASCADE"), unique=True)
    display_order: Mapped[int] = mapped_column(unique=True)
    is_default: Mapped[bool] = mapped_column(default=False)
    created_at: Mapped[datetime] = mapped_column(server_default=func.now())

    # Relationship
    status: Mapped["Status"] = relationship()


class TaskPriority(Base):
    """TaskPriority model - configures which priorities are valid for Tasks"""
    __tablename__ = "Task_Priority"

    id: Mapped[int] = mapped_column(primary_key=True)
    status_id: Mapped[int] = mapped_column(ForeignKey("Status.id", ondelete="CASCADE"), unique=True)
    display_order: Mapped[int] = mapped_column(unique=True)
    is_default: Mapped[bool] = mapped_column(default=False)
    created_at: Mapped[datetime] = mapped_column(server_default=func.now())

    # Relationship
    status: Mapped["Status"] = relationship()


class Organization(Base):
    """Organization model - represents companies"""
    __tablename__ = "Organization"

    id: Mapped[int] = mapped_column(primary_key=True)
    tenant_id: Mapped[Optional[int]] = mapped_column(ForeignKey("Tenant.id", ondelete="CASCADE"))
    name: Mapped[str]
    website: Mapped[Optional[str]]
    phone: Mapped[Optional[str]]
    industry: Mapped[Optional[str]]
    employee_count: Mapped[Optional[int]]
    description: Mapped[Optional[str]]
    notes: Mapped[Optional[str]]
    created_at: Mapped[datetime] = mapped_column(server_default=func.now())
    updated_at: Mapped[datetime] = mapped_column(server_default=func.now(), onupdate=func.now())

    # Note: Uniqueness enforced by migration index: idx_organization_name_lower_tenant ON (tenant_id, LOWER(name))

    # Relationships
    contacts: Mapped[List["Contact"]] = relationship(back_populates="organization", foreign_keys="Contact.organization_id")
    lead_organizations: Mapped[List["LeadOrganization"]] = relationship(back_populates="organization", cascade="all, delete-orphan")
    project_organizations: Mapped[List["ProjectOrganization"]] = relationship(back_populates="organization", cascade="all, delete-orphan")


class Individual(Base):
    """Individual model - represents a person"""
    __tablename__ = "Individual"

    id: Mapped[int] = mapped_column(primary_key=True)
    tenant_id: Mapped[Optional[int]] = mapped_column(ForeignKey("Tenant.id", ondelete="CASCADE"))
    first_name: Mapped[str]
    last_name: Mapped[str]
    email: Mapped[Optional[str]]
    phone: Mapped[Optional[str]]
    linkedin_url: Mapped[Optional[str]]
    title: Mapped[Optional[str]]
    notes: Mapped[Optional[str]]
    created_at: Mapped[datetime] = mapped_column(server_default=func.now())
    updated_at: Mapped[datetime] = mapped_column(server_default=func.now(), onupdate=func.now())

    # Relationships
    contacts: Mapped[List["Contact"]] = relationship(back_populates="individual", cascade="all, delete-orphan")
    lead_individuals: Mapped[List["LeadIndividual"]] = relationship(back_populates="individual", cascade="all, delete-orphan")
    project_individuals: Mapped[List["ProjectIndividual"]] = relationship(back_populates="individual", cascade="all, delete-orphan")


class Contact(Base):
    """Contact model - person in a role/context"""
    __tablename__ = "Contact"

    id: Mapped[int] = mapped_column(primary_key=True)
    individual_id: Mapped[int] = mapped_column(ForeignKey("Individual.id", ondelete="CASCADE"))
    organization_id: Mapped[Optional[int]] = mapped_column(ForeignKey("Organization.id", ondelete="SET NULL"))
    title: Mapped[Optional[str]]
    department: Mapped[Optional[str]]
    role: Mapped[Optional[str]]
    email: Mapped[Optional[str]]
    phone: Mapped[Optional[str]]
    is_primary: Mapped[bool] = mapped_column(default=False)
    notes: Mapped[Optional[str]]
    created_at: Mapped[datetime] = mapped_column(server_default=func.now())
    updated_at: Mapped[datetime] = mapped_column(server_default=func.now(), onupdate=func.now())

    # Relationships
    individual: Mapped["Individual"] = relationship(back_populates="contacts")
    organization: Mapped[Optional["Organization"]] = relationship(back_populates="contacts", foreign_keys=[organization_id])
    lead_contacts: Mapped[List["LeadContact"]] = relationship(back_populates="contact", cascade="all, delete-orphan")
    project_contacts: Mapped[List["ProjectContact"]] = relationship(back_populates="contact", cascade="all, delete-orphan")


class Task(Base):
    """Task model - polymorphic association to any entity"""
    __tablename__ = "Task"

    id: Mapped[int] = mapped_column(primary_key=True)
    taskable_type: Mapped[Optional[str]]
    taskable_id: Mapped[Optional[int]]
    title: Mapped[str]
    description: Mapped[Optional[str]]
    current_status_id: Mapped[Optional[int]] = mapped_column(ForeignKey("Status.id", ondelete="SET NULL"))
    priority_id: Mapped[Optional[int]] = mapped_column(ForeignKey("Status.id", ondelete="SET NULL"))
    due_date: Mapped[Optional[date]] = mapped_column(Date)
    completed_at: Mapped[Optional[datetime]]
    assigned_to_user_id: Mapped[Optional[int]]
    created_at: Mapped[datetime] = mapped_column(server_default=func.now())
    updated_at: Mapped[datetime] = mapped_column(server_default=func.now(), onupdate=func.now())

    # Relationships
    current_status: Mapped[Optional["Status"]] = relationship(foreign_keys=[current_status_id])
    priority: Mapped[Optional["Status"]] = relationship(foreign_keys=[priority_id])


# ==============================
# Association Objects
# ==============================


class LeadContact(Base):
    """Lead ↔ Contact association"""
    __tablename__ = "Lead_Contact"

    lead_id: Mapped[int] = mapped_column(ForeignKey("Lead.id", ondelete="CASCADE"), primary_key=True)
    contact_id: Mapped[int] = mapped_column(ForeignKey("Contact.id", ondelete="CASCADE"), primary_key=True)
    role_on_lead: Mapped[Optional[str]]
    is_primary: Mapped[bool] = mapped_column(default=False)
    created_at: Mapped[datetime] = mapped_column(server_default=func.now())

    # Relationships
    lead: Mapped["Lead"] = relationship(back_populates="lead_contacts")
    contact: Mapped["Contact"] = relationship(back_populates="lead_contacts")


class LeadOrganization(Base):
    """Lead ↔ Organization association"""
    __tablename__ = "Lead_Organization"

    lead_id: Mapped[int] = mapped_column(ForeignKey("Lead.id", ondelete="CASCADE"), primary_key=True)
    organization_id: Mapped[int] = mapped_column(ForeignKey("Organization.id", ondelete="CASCADE"), primary_key=True)
    relationship: Mapped[Optional[str]]
    is_primary: Mapped[bool] = mapped_column(default=False)
    created_at: Mapped[datetime] = mapped_column(server_default=func.now())

    # Relationships
    lead: Mapped["Lead"] = relationship(back_populates="lead_organizations")
    organization: Mapped["Organization"] = relationship(back_populates="lead_organizations")


class LeadIndividual(Base):
    """Lead ↔ Individual association"""
    __tablename__ = "Lead_Individual"

    lead_id: Mapped[int] = mapped_column(ForeignKey("Lead.id", ondelete="CASCADE"), primary_key=True)
    individual_id: Mapped[int] = mapped_column(ForeignKey("Individual.id", ondelete="CASCADE"), primary_key=True)
    relationship: Mapped[Optional[str]]
    is_primary: Mapped[bool] = mapped_column(default=False)
    created_at: Mapped[datetime] = mapped_column(server_default=func.now())

    # Relationships
    lead: Mapped["Lead"] = relationship(back_populates="lead_individuals")
    individual: Mapped["Individual"] = relationship(back_populates="lead_individuals")


class ProjectDeal(Base):
    """Project ↔ Deal association"""
    __tablename__ = "Project_Deal"

    project_id: Mapped[int] = mapped_column(ForeignKey("Project.id", ondelete="CASCADE"), primary_key=True)
    deal_id: Mapped[int] = mapped_column(ForeignKey("Deal.id", ondelete="CASCADE"), primary_key=True)
    relationship: Mapped[Optional[str]]
    created_at: Mapped[datetime] = mapped_column(server_default=func.now())

    # Relationships
    project: Mapped["Project"] = relationship(back_populates="project_deals")
    deal: Mapped["Deal"] = relationship(back_populates="project_deals")


class ProjectLead(Base):
    """Project ↔ Lead association"""
    __tablename__ = "Project_Lead"

    project_id: Mapped[int] = mapped_column(ForeignKey("Project.id", ondelete="CASCADE"), primary_key=True)
    lead_id: Mapped[int] = mapped_column(ForeignKey("Lead.id", ondelete="CASCADE"), primary_key=True)
    relationship: Mapped[Optional[str]]
    created_at: Mapped[datetime] = mapped_column(server_default=func.now())

    # Relationships
    project: Mapped["Project"] = relationship(back_populates="project_leads")
    lead: Mapped["Lead"] = relationship(back_populates="project_leads")


class ProjectContact(Base):
    """Project ↔ Contact association"""
    __tablename__ = "Project_Contact"

    project_id: Mapped[int] = mapped_column(ForeignKey("Project.id", ondelete="CASCADE"), primary_key=True)
    contact_id: Mapped[int] = mapped_column(ForeignKey("Contact.id", ondelete="CASCADE"), primary_key=True)
    role: Mapped[Optional[str]]
    is_primary: Mapped[bool] = mapped_column(default=False)
    created_at: Mapped[datetime] = mapped_column(server_default=func.now())

    # Relationships
    project: Mapped["Project"] = relationship(back_populates="project_contacts")
    contact: Mapped["Contact"] = relationship(back_populates="project_contacts")


class ProjectOrganization(Base):
    """Project ↔ Organization association"""
    __tablename__ = "Project_Organization"

    project_id: Mapped[int] = mapped_column(ForeignKey("Project.id", ondelete="CASCADE"), primary_key=True)
    organization_id: Mapped[int] = mapped_column(ForeignKey("Organization.id", ondelete="CASCADE"), primary_key=True)
    relationship: Mapped[Optional[str]]
    is_primary: Mapped[bool] = mapped_column(default=False)
    created_at: Mapped[datetime] = mapped_column(server_default=func.now())

    # Relationships
    project: Mapped["Project"] = relationship(back_populates="project_organizations")
    organization: Mapped["Organization"] = relationship(back_populates="project_organizations")


class ProjectIndividual(Base):
    """Project ↔ Individual association"""
    __tablename__ = "Project_Individual"

    project_id: Mapped[int] = mapped_column(ForeignKey("Project.id", ondelete="CASCADE"), primary_key=True)
    individual_id: Mapped[int] = mapped_column(ForeignKey("Individual.id", ondelete="CASCADE"), primary_key=True)
    role: Mapped[Optional[str]]
    is_primary: Mapped[bool] = mapped_column(default=False)
    created_at: Mapped[datetime] = mapped_column(server_default=func.now())

    # Relationships
    project: Mapped["Project"] = relationship(back_populates="project_individuals")
    individual: Mapped["Individual"] = relationship(back_populates="project_individuals")


# ==============================
# Helper Functions for Polymorphic Task Queries
# ==============================


def get_tasks_for_entity(session, entity_type: str, entity_id: int):
    """
    Query tasks for any entity type.

    Example:
        get_tasks_for_entity(session, 'Deal', deal.id)
        get_tasks_for_entity(session, 'Job', job.id)
    """
    return session.query(Task).filter_by(
        taskable_type=entity_type,
        taskable_id=entity_id
    ).all()
