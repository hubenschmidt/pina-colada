"""Service layer for user business logic."""

from typing import Optional

from fastapi import HTTPException

from lib.db import async_get_session
from models.Tenant import Tenant
from models.Individual import Individual
from repositories.user_repository import find_user_by_email


async def get_user_tenant(email: str) -> dict:
    """Get tenant info for a user by email."""
    user = await find_user_by_email(email)
    if not user:
        raise HTTPException(status_code=404, detail="User not found")

    if not user.tenant_id:
        raise HTTPException(status_code=404, detail="No tenant associated with user")

    async with async_get_session() as session:
        tenant = await session.get(Tenant, user.tenant_id)
        if not tenant:
            raise HTTPException(status_code=404, detail="Tenant not found")

        individual = None
        if user.individual_id:
            individual = await session.get(Individual, user.individual_id)

        return {
            "tenant": tenant,
            "individual": individual,
        }
