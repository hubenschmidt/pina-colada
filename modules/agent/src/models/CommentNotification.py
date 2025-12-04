"""CommentNotification model for tracking comment activity notifications."""

from datetime import datetime
from sqlalchemy import Column, Integer, String, Boolean, DateTime, ForeignKey
from sqlalchemy.orm import
from models import Base


class CommentNotification(Base):
    __tablename__ = "Comment_Notification"

    id = Column(Integer, primary_key=True)
    tenant_id = Column(Integer, ForeignKey("Tenant.id", ondelete="CASCADE"), nullable=False)
    user_id = Column(Integer, ForeignKey("User.id", ondelete="CASCADE"), nullable=False)
    comment_id = Column(Integer, ForeignKey("Comment.id", ondelete="CASCADE"), nullable=False)
    notification_type = Column(String(20), nullable=False)  # direct_reply, thread_activity
    is_read = Column(Boolean, default=False)
    created_at = Column(DateTime(timezone=True), default=datetime.utcnow)
    updated_at = Column(DateTime(timezone=True), default=datetime.utcnow, onupdate=datetime.utcnow)
    created_by = Column(Integer, ForeignKey("User.id"), nullable=False)
    updated_by = Column(Integer, ForeignKey("User.id"), nullable=False)
    read_at = Column(DateTime(timezone=True), nullable=True)

    # Relationships
    tenant = relationship("Tenant", foreign_keys=[tenant_id])
    user = relationship("User", foreign_keys=[user_id])
    comment = relationship("Comment", foreign_keys=[comment_id])

    def __repr__(self):
        return f"<CommentNotification(id={self.id}, user_id={self.user_id}, type={self.notification_type})>"
