from sqlalchemy import Column, Text, DateTime, BigInteger, Date, ForeignKey, func, Index
from sqlalchemy.orm import relationship
from models import Base

"""Project model."""




class Project(Base):
    """Project SQLAlchemy model."""

    __tablename__ = "Project"

    id = Column(BigInteger, primary_key=True, autoincrement=True)
    tenant_id = Column(BigInteger, ForeignKey("Tenant.id", ondelete="CASCADE"), nullable=True)
    name = Column(Text, nullable=False)
    description = Column(Text, nullable=True)
    owner_user_id = Column(BigInteger, nullable=True)
    status = Column(Text, nullable=True)
    current_status_id = Column(BigInteger, ForeignKey("Status.id", ondelete="SET NULL"), nullable=True)
    start_date = Column(Date, nullable=True)
    end_date = Column(Date, nullable=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), nullable=False)

    # Relationships
    tenant = relationship("Tenant", back_populates="projects")
    current_status = relationship("Status", back_populates="projects")
    deals = relationship("Deal", back_populates="project")
    leads = relationship("Lead", secondary="Lead_Project", back_populates="projects")

    __table_args__ = (
        Index('idx_project_tenant_id', 'tenant_id'),
        Index('idx_project_current_status_id', 'current_status_id'),
    )
