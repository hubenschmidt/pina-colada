from sqlalchemy import Column, Text, DateTime, BigInteger, ForeignKey, CheckConstraint, func
from sqlalchemy.dialects.postgresql import UUID, JSONB
from sqlalchemy.orm import relationship
from models import Base


class Conversation(Base):
    """Chat conversation metadata."""

    __tablename__ = "Conversation"

    id = Column(BigInteger, primary_key=True, autoincrement=True)
    tenant_id = Column(BigInteger, ForeignKey("Tenant.id"), nullable=False)
    user_id = Column(BigInteger, ForeignKey("User.id"), nullable=False)
    thread_id = Column(UUID(as_uuid=True), nullable=False, unique=True)
    title = Column(Text, nullable=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), nullable=False)
    archived_at = Column(DateTime(timezone=True), nullable=True)

    messages = relationship("ConversationMessage", back_populates="conversation", cascade="all, delete-orphan")


class ConversationMessage(Base):
    """Individual message within a conversation."""

    __tablename__ = "Conversation_Message"

    id = Column(BigInteger, primary_key=True, autoincrement=True)
    conversation_id = Column(BigInteger, ForeignKey("Conversation.id", ondelete="CASCADE"), nullable=False)
    role = Column(Text, nullable=False)
    content = Column(Text, nullable=False)
    token_usage = Column(JSONB, nullable=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)

    conversation = relationship("Conversation", back_populates="messages")

    __table_args__ = (
        CheckConstraint(role.in_(["user", "assistant"]), name="ck_conversation_message_role"),
    )
