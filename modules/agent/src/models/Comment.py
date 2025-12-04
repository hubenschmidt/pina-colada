"""Comment model for polymorphic comments across entities."""

from datetime import datetime
from sqlalchemy import Column, Integer, String, Text, DateTime, ForeignKey
from sqlalchemy.orm import relationship
from models import Base


class Comment(Base):
    __tablename__ = "Comment"

    id = Column(Integer, primary_key=True)
    tenant_id = Column(Integer, ForeignKey("Tenant.id", ondelete="CASCADE"), nullable=False)
    commentable_type = Column(String(50), nullable=False)
    commentable_id = Column(Integer, nullable=False)
    content = Column(Text, nullable=False)
    created_by = Column(Integer, ForeignKey("User.id"), nullable=False)
    updated_by = Column(Integer, ForeignKey("User.id"), nullable=False)
    parent_comment_id = Column(Integer, ForeignKey("Comment.id", ondelete="CASCADE"), nullable=True)
    created_at = Column(DateTime(timezone=True), default=datetime.utcnow)
    updated_at = Column(DateTime(timezone=True), default=datetime.utcnow, onupdate=datetime.utcnow)

    # Relationships
    tenant = relationship("Tenant", back_populates="comments")
    creator = relationship("User", foreign_keys=[created_by])
    parent = relationship("Comment", remote_side=[id], backref="replies")

    def __repr__(self):
        return f"<Comment(id={self.id}, commentable_type={self.commentable_type}, commentable_id={self.commentable_id})>"
