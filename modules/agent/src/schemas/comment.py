"""Comment schemas for API validation."""

from pydantic import BaseModel


class CommentCreate(BaseModel):
    commentable_type: str
    commentable_id: int
    content: str
    parent_comment_id: int | None = None


class CommentUpdate(BaseModel):
    content: str
