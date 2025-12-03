from sqlalchemy import Column, Text, DateTime, BigInteger, ForeignKey, func
from sqlalchemy.orm import relationship
from models import Base
from models.Industry import Account_Industry


class Account(Base):
    """Account for classifying Organizations and Individuals."""

    __tablename__ = "Account"

    id = Column(BigInteger, primary_key=True, autoincrement=True)
    tenant_id = Column(BigInteger, ForeignKey("Tenant.id", ondelete="CASCADE"), nullable=True)
    name = Column(Text, nullable=False)
    created_at = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), nullable=False)
    created_by = Column(BigInteger, ForeignKey("User.id"), nullable=False)
    updated_by = Column(BigInteger, ForeignKey("User.id"), nullable=False)

    # Relationships
    tenant = relationship("Tenant", back_populates="accounts")
    organizations = relationship("Organization", back_populates="account")
    individuals = relationship("Individual", back_populates="account")
    leads = relationship("Lead", back_populates="account")
    industries = relationship("Industry", secondary=Account_Industry, back_populates="accounts")
    projects = relationship("Project", secondary="Account_Project", backref="accounts")
