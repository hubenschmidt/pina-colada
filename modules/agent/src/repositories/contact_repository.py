"""Repository layer for contact data access."""

import logging
from typing import List, Optional, Dict, Any
from sqlalchemy import select, text
from models.Contact import Contact
from models.Individual import Individual
from lib.db import async_get_session

logger = logging.getLogger(__name__)


async def find_contact_by_individual_and_org(
    individual_id: int, organization_id: Optional[int] = None
) -> Optional[Contact]:
    """Find contact by individual and organization."""
    async with async_get_session() as session:
        stmt = select(Contact).where(Contact.individual_id == individual_id)
        if organization_id is not None:
            stmt = stmt.where(Contact.organization_id == organization_id)
            result = await session.execute(stmt)
            return result.scalar_one_or_none()
        
        stmt = stmt.where(Contact.organization_id.is_(None))
        result = await session.execute(stmt)
        return result.scalar_one_or_none()


async def create_contact(data: Dict[str, Any]) -> Contact:
    """Create a new contact record."""
    async with async_get_session() as session:
        try:
            contact = Contact(**data)
            session.add(contact)
            await session.commit()
            await session.refresh(contact)
            return contact
        except Exception as e:
            await session.rollback()
            logger.error(f"Failed to create contact: {e}")
            raise


async def get_or_create_contact(
    individual_id: int,
    organization_id: Optional[int] = None,
    email: Optional[str] = None,
    phone: Optional[str] = None,
    title: Optional[str] = None,
    is_primary: bool = False
) -> Contact:
    """Get or create a contact for an individual at an organization."""
    existing = await find_contact_by_individual_and_org(individual_id, organization_id)
    if existing:
        return existing

    return await create_contact({
        "individual_id": individual_id,
        "organization_id": organization_id,
        "email": email,
        "phone": phone,
        "title": title,
        "is_primary": is_primary
    })


async def create_lead_contact(
    lead_id: int,
    contact_id: int,
    is_primary: bool = False,
    role_on_lead: Optional[str] = None
) -> None:
    """Create a Lead_Contact junction entry."""
    async with async_get_session() as session:
        try:
            await session.execute(
                text("""
                    INSERT INTO "Lead_Contact" (lead_id, contact_id, is_primary, role_on_lead, created_at)
                    VALUES (:lead_id, :contact_id, :is_primary, :role_on_lead, NOW())
                    ON CONFLICT (lead_id, contact_id) DO NOTHING
                """),
                {
                    "lead_id": lead_id,
                    "contact_id": contact_id,
                    "is_primary": is_primary,
                    "role_on_lead": role_on_lead
                }
            )
            await session.commit()
        except Exception as e:
            await session.rollback()
            logger.error(f"Failed to create lead_contact: {e}")
            raise


async def delete_lead_contacts_by_lead(lead_id: int) -> None:
    """Delete all Lead_Contact entries for a lead."""
    async with async_get_session() as session:
        try:
            await session.execute(
                text('DELETE FROM "Lead_Contact" WHERE lead_id = :lead_id'),
                {"lead_id": lead_id}
            )
            await session.commit()
        except Exception as e:
            await session.rollback()
            logger.error(f"Failed to delete lead_contacts: {e}")
            raise


async def find_contacts_by_lead(lead_id: int) -> List[Contact]:
    """Find all contacts for a lead."""
    async with async_get_session() as session:
        stmt = text("""
            SELECT c.* FROM "Contact" c
            JOIN "Lead_Contact" lc ON c.id = lc.contact_id
            WHERE lc.lead_id = :lead_id
            ORDER BY lc.is_primary DESC, c.id ASC
        """)
        result = await session.execute(stmt, {"lead_id": lead_id})
        rows = result.fetchall()

        contacts = []
        for row in rows:
            contact = await session.get(Contact, row.id)
            if contact:
                contacts.append(contact)
        return contacts


async def find_contacts_by_individual(individual_id: int) -> List[Contact]:
    """Find all contacts for an individual."""
    async with async_get_session() as session:
        stmt = select(Contact).where(Contact.individual_id == individual_id).order_by(
            Contact.is_primary.desc(), Contact.id
        )
        result = await session.execute(stmt)
        return list(result.scalars().all())


async def find_contacts_by_organization(organization_id: int) -> List[Contact]:
    """Find all contacts for an organization."""
    async with async_get_session() as session:
        stmt = select(Contact).where(Contact.organization_id == organization_id).order_by(
            Contact.is_primary.desc(), Contact.id
        )
        result = await session.execute(stmt)
        return list(result.scalars().all())


async def find_contact_by_id(contact_id: int) -> Optional[Contact]:
    """Find a contact by ID."""
    async with async_get_session() as session:
        return await session.get(Contact, contact_id)


async def update_contact(contact_id: int, data: Dict[str, Any]) -> Optional[Contact]:
    """Update an existing contact."""
    async with async_get_session() as session:
        try:
            contact = await session.get(Contact, contact_id)
            if not contact:
                return None

            for key, value in data.items():
                if hasattr(contact, key):
                    setattr(contact, key, value)

            await session.commit()
            await session.refresh(contact)
            return contact
        except Exception as e:
            await session.rollback()
            logger.error(f"Failed to update contact: {e}")
            raise


async def delete_contact(contact_id: int) -> bool:
    """Delete a contact by ID."""
    async with async_get_session() as session:
        try:
            contact = await session.get(Contact, contact_id)
            if not contact:
                return False
            await session.delete(contact)
            await session.commit()
            return True
        except Exception as e:
            await session.rollback()
            logger.error(f"Failed to delete contact: {e}")
            raise


async def search_contacts_and_individuals(query: str, tenant_id: Optional[int] = None) -> List[Dict[str, Any]]:
    """
    Search both Individuals and existing Contacts.
    Returns unified results with individual info for linking.
    """
    async with async_get_session() as session:
        search_pattern = f"%{query}%"

        if tenant_id is not None:
            sql = text("""
                SELECT DISTINCT
                    i.id as individual_id,
                    i.first_name,
                    i.last_name,
                    i.email,
                    i.phone,
                    'individual' as source
                FROM "Individual" i
                LEFT JOIN "Account" a ON i.account_id = a.id
                WHERE (
                    i.first_name ILIKE :pattern
                    OR i.last_name ILIKE :pattern
                    OR i.email ILIKE :pattern
                    OR CONCAT(i.first_name, ' ', i.last_name) ILIKE :pattern
                )
                AND a.tenant_id = :tenant_id
                ORDER BY i.first_name, i.last_name
                LIMIT 20
            """)
            result = await session.execute(sql, {"pattern": search_pattern, "tenant_id": tenant_id})
            rows = result.fetchall()
            return [
                {
                    "individual_id": row.individual_id,
                    "first_name": row.first_name,
                    "last_name": row.last_name,
                    "email": row.email,
                    "phone": row.phone,
                    "source": row.source,
                }
                for row in rows
            ]
        
        sql = text("""
            SELECT DISTINCT
                i.id as individual_id,
                i.first_name,
                i.last_name,
                i.email,
                i.phone,
                'individual' as source
            FROM "Individual" i
            WHERE (
                i.first_name ILIKE :pattern
                OR i.last_name ILIKE :pattern
                OR i.email ILIKE :pattern
                OR CONCAT(i.first_name, ' ', i.last_name) ILIKE :pattern
            )
            ORDER BY i.first_name, i.last_name
            LIMIT 20
        """)
        result = await session.execute(sql, {"pattern": search_pattern})
        rows = result.fetchall()
        return [
            {
                "individual_id": row.individual_id,
                "first_name": row.first_name,
                "last_name": row.last_name,
                "email": row.email,
                "phone": row.phone,
                "source": row.source,
            }
            for row in rows
        ]
