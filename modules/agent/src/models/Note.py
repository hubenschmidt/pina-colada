"""Note model for polymorphic notes across entities."""

from datetime import datetime
from sqlalchemy import Column, Integer, String, Text, DateTime, ForeignKey
from sqlalchemy.orm import relationship
from models import Base


class Note(Base):
    __tablename__ = "Note"

    id = Column(Integer, primary_key=True)
    tenant_id = Column(Integer, ForeignKey("Tenant.id", ondelete="CASCADE"), nullable=False)
    entity_type = Column(String(50), nullable=False)
    entity_id = Column(Integer, nullable=False)
    content = Column(Text, nullable=False)
    created_by = Column(Integer, ForeignKey("User.id", ondelete="SET NULL"), nullable=True)
    updated_by = Column(Integer, ForeignKey("User.id", ondelete="SET NULL"), nullable=True)
    created_at = Column(DateTime(timezone=True), default=datetime.utcnow)
    updated_at = Column(DateTime(timezone=True), default=datetime.utcnow, onupdate=datetime.utcnow)

    # Relationships
    tenant = relationship("Tenant", back_populates="notes")
    creator = relationship("User", foreign_keys=[created_by])
    updater = relationship("User", foreign_keys=[updated_by])

    def __repr__(self):
        return f"<Note(id={self.id}, entity_type={self.entity_type}, entity_id={self.entity_id})>"
