from sqlalchemy import Column, Text, DateTime, BigInteger, ForeignKey, func
from sqlalchemy.orm import relationship
from models import Base

"""Role model for tenant-scoped roles."""




class Role(Base):
    """Role SQLAlchemy model (tenant-scoped roles)."""

    __tablename__ = "Role"

    id = Column(BigInteger, primary_key=True, autoincrement=True)
    tenant_id = Column(BigInteger, ForeignKey("Tenant.id"), nullable=False)
    name = Column(Text, nullable=False)
    description = Column(Text, nullable=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), nullable=False)
    created_by = Column(BigInteger, ForeignKey("User.id"), nullable=False)
    updated_by = Column(BigInteger, ForeignKey("User.id"), nullable=False)

    # Relationships
    tenant = relationship("Tenant", back_populates="roles", foreign_keys=[tenant_id])
    user_roles = relationship("UserRole", back_populates="role", foreign_keys="UserRole.role_id")
