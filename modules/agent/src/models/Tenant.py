from sqlalchemy import Column, Text, DateTime, BigInteger, func, CheckConstraint
from sqlalchemy.dialects.postgresql import JSONB
from sqlalchemy.orm import relationship
from models import Base

"""Tenant model for multi-tenant isolation."""




class Tenant(Base):
    """Tenant SQLAlchemy model (app customer/company)."""

    __tablename__ = "Tenant"

    id = Column(BigInteger, primary_key=True, autoincrement=True)
    name = Column(Text, nullable=False)
    slug = Column(Text, nullable=False, unique=True)
    status = Column(Text, nullable=False, default='active')
    plan = Column(Text, nullable=False, default='free')
    settings = Column(JSONB, nullable=False, default={}, server_default='{}')
    created_at = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), nullable=False)

    # Relationships
    users = relationship("User", back_populates="tenant")
    roles = relationship("Role", back_populates="tenant")
    deals = relationship("Deal", back_populates="tenant")
    organizations = relationship("Organization", back_populates="tenant")
    individuals = relationship("Individual", back_populates="tenant")
    projects = relationship("Project", back_populates="tenant")

    __table_args__ = (
        CheckConstraint("status IN ('active', 'suspended', 'trial', 'cancelled')", name='tenant_status_check'),
        CheckConstraint("plan IN ('free', 'starter', 'professional', 'enterprise')", name='tenant_plan_check'),
    )
