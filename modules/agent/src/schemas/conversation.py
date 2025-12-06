"""Conversation schemas for API validation."""

from datetime import datetime
from typing import Optional
from uuid import UUID

from pydantic import BaseModel


class ConversationCreate(BaseModel):
    thread_id: UUID
    title: Optional[str] = None


class ConversationUpdate(BaseModel):
    title: str


class MessageCreate(BaseModel):
    role: str
    content: str
    token_usage: Optional[dict] = None


class MessageResponse(BaseModel):
    id: int
    role: str
    content: str
    token_usage: Optional[dict]
    created_at: datetime

    class Config:
        from_attributes = True


class UserBrief(BaseModel):
    id: int
    email: Optional[str] = None

    class Config:
        from_attributes = True


class ConversationResponse(BaseModel):
    id: int
    thread_id: UUID
    title: Optional[str]
    created_at: datetime
    updated_at: datetime
    archived_at: Optional[datetime]

    class Config:
        from_attributes = True


class ConversationWithUserResponse(ConversationResponse):
    """Conversation response with user details for full listing."""
    created_by_id: Optional[int] = None
    updated_by_id: Optional[int] = None
    created_by: Optional[UserBrief] = None
    updated_by: Optional[UserBrief] = None


class ConversationWithMessagesResponse(BaseModel):
    conversation: ConversationResponse
    messages: list[MessageResponse]
