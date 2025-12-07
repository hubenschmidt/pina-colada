"""Role checking utilities for authorization."""

from functools import wraps
from typing import Callable, List

from fastapi import HTTPException, Request
from sqlalchemy import and_, select
from lib.db import async_get_session
from models.Role import Role
from models.UserRole import UserRole


async def user_has_role(user_id: int, role_name: str) -> bool:
    """Check if user has a specific role."""
    async with async_get_session() as session:
        stmt = (
            select(UserRole)
            .join(Role, UserRole.role_id == Role.id)
            .where(
                and_(
                    UserRole.user_id == user_id,
                    Role.name == role_name,
                )
            )
        )
        result = await session.execute(stmt)
        return result.scalar_one_or_none() is not None


async def get_user_roles(user_id: int) -> List[str]:
    """Get all role names for a user."""
    async with async_get_session() as session:
        stmt = (
            select(Role.name)
            .join(UserRole, UserRole.role_id == Role.id)
            .where(UserRole.user_id == user_id)
        )
        result = await session.execute(stmt)
        return [row[0] for row in result.all()]


def require_role(role_name: str):
    """Decorator to require a specific role for route access."""
    def decorator(func: Callable):
        @wraps(func)
        async def wrapper(request: Request, *args, **kwargs):
            user_id = getattr(request.state, "user_id", None)
            if not user_id:
                raise HTTPException(status_code=401, detail="Not authenticated")

            has_role = await user_has_role(user_id, role_name)
            if not has_role:
                raise HTTPException(status_code=403, detail="Insufficient permissions")

            return await func(request, *args, **kwargs)
        return wrapper
    return decorator
