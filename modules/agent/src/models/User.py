from sqlalchemy import Column, Text, DateTime, BigInteger, ForeignKey, func, Index, UniqueConstraint, CheckConstraint
from sqlalchemy.orm import relationship
from models import Base

"""User model for multi-tenant users."""




class User(Base):
    """User SQLAlchemy model (belongs to one Tenant)."""

    __tablename__ = "User"

    id = Column(BigInteger, primary_key=True, autoincrement=True)
    tenant_id = Column(BigInteger, ForeignKey("Tenant.id", ondelete="CASCADE"), nullable=True)
    auth0_sub = Column(Text, unique=True, nullable=True)
    email = Column(Text, nullable=False)
    first_name = Column(Text, nullable=True)
    last_name = Column(Text, nullable=True)
    avatar_url = Column(Text, nullable=True)
    status = Column(Text, nullable=False, default='active')
    last_login_at = Column(DateTime(timezone=True), nullable=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), nullable=False)

    # Relationships
    tenant = relationship("Tenant", back_populates="users")
    user_roles = relationship("UserRole", back_populates="user")

    __table_args__ = (
        UniqueConstraint('tenant_id', 'email', name='user_tenant_email_unique'),
        Index('idx_user_tenant_id', 'tenant_id'),
        Index('idx_user_email', 'email'),
        Index('idx_user_auth0_sub', 'auth0_sub'),
        CheckConstraint("status IN ('active', 'inactive', 'invited')", name='user_status_check'),
    )
