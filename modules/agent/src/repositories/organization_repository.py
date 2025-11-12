"""Repository layer for organization data access."""

import logging
from typing import List, Optional
from sqlalchemy import select, func
from models.Organization import Organization, OrganizationCreateData, OrganizationUpdateData, orm_to_dict, dict_to_orm, update_orm_from_dict
from lib.db import get_session

logger = logging.getLogger(__name__)


def find_all_organizations(tenant_id: Optional[int] = None) -> List[Organization]:
    """Find all organizations, optionally filtered by tenant."""
    session = get_session()
    try:
        stmt = select(Organization).order_by(Organization.name)
        if tenant_id is not None:
            stmt = stmt.where(Organization.tenant_id == tenant_id)
        return list(session.execute(stmt).scalars().all())
    finally:
        session.close()


def find_organization_by_id(org_id: int) -> Optional[Organization]:
    """Find organization by ID."""
    session = get_session()
    try:
        return session.get(Organization, org_id)
    finally:
        session.close()


def find_organization_by_name(name: str, tenant_id: Optional[int] = None) -> Optional[Organization]:
    """Find organization by name (case-insensitive), optionally scoped to tenant."""
    session = get_session()
    try:
        stmt = select(Organization).where(func.lower(Organization.name) == func.lower(name))
        if tenant_id is not None:
            stmt = stmt.where(Organization.tenant_id == tenant_id)
        return session.execute(stmt).scalar_one_or_none()
    finally:
        session.close()


def create_organization(data: OrganizationCreateData) -> Organization:
    """Create a new organization."""
    session = get_session()
    try:
        org = dict_to_orm(data)
        session.add(org)
        session.commit()
        session.refresh(org)
        return org
    except Exception as e:
        session.rollback()
        logger.error(f"Failed to create organization: {e}")
        raise
    finally:
        session.close()


def get_or_create_organization(name: str, tenant_id: Optional[int] = None) -> Organization:
    """Get or create organization by name."""
    existing = find_organization_by_name(name, tenant_id)
    if existing:
        return existing

    data: OrganizationCreateData = {
        "name": name,
        "tenant_id": tenant_id
    }
    return create_organization(data)


def update_organization(org_id: int, data: OrganizationUpdateData) -> Optional[Organization]:
    """Update an existing organization."""
    session = get_session()
    try:
        org = session.get(Organization, org_id)
        if not org:
            return None

        update_orm_from_dict(org, data)
        session.commit()
        session.refresh(org)
        return org
    except Exception as e:
        session.rollback()
        logger.error(f"Failed to update organization: {e}")
        raise
    finally:
        session.close()


def delete_organization(org_id: int) -> bool:
    """Delete an organization by ID."""
    session = get_session()
    try:
        org = session.get(Organization, org_id)
        if not org:
            return False
        session.delete(org)
        session.commit()
        return True
    except Exception as e:
        session.rollback()
        logger.error(f"Failed to delete organization: {e}")
        raise
    finally:
        session.close()


def search_organizations(query: str, tenant_id: Optional[int] = None) -> List[Organization]:
    """Search organizations by name (case-insensitive partial match)."""
    session = get_session()
    try:
        stmt = select(Organization).where(
            func.lower(Organization.name).contains(func.lower(query))
        ).order_by(Organization.name)
        if tenant_id is not None:
            stmt = stmt.where(Organization.tenant_id == tenant_id)
        return list(session.execute(stmt).scalars().all())
    finally:
        session.close()
