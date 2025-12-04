"""Controller layer for individual routing to services."""

from typing import List, Optional

from fastapi import Request

from lib.decorators import handle_http_exceptions
from serializers.common import to_paged_response, contact_to_dict
from serializers.individual import (
    ind_to_list_response,
    ind_to_search_response,
    ind_to_detail_response,
)
from schemas.individual import IndContactCreate, IndContactUpdate, IndividualCreate, IndividualUpdate
from services.individual_service import (
    get_individuals_paginated,
    search_individuals as search_individuals_service,
    get_individual as get_individual_service,
    create_individual as create_individual_service,
    update_individual as update_individual_service,
    delete_individual as delete_individual_service,
    get_individual_contacts as get_contacts_service,
    create_individual_contact as create_contact_service,
    update_individual_contact as update_contact_service,
    delete_individual_contact as delete_contact_service,
)

# Re-export for routes
__all__ = ["IndividualCreate", "IndividualUpdate", "IndContactCreate", "IndContactUpdate"]


# Individual CRUD

@handle_http_exceptions
async def get_individuals(
    request: Request,
    page: int,
    limit: int,
    order_by: str,
    order: str,
    search: Optional[str] = None,
) -> dict:
    """Get individuals with pagination."""
    tenant_id = request.state.tenant_id
    individuals, total = await get_individuals_paginated(
        tenant_id, page, limit, order_by, order, search
    )
    items = [ind_to_list_response(ind) for ind in individuals]
    return to_paged_response(total, page, limit, items)


@handle_http_exceptions
async def search_individuals(request: Request, query: str) -> List[dict]:
    """Search individuals by name or email."""
    tenant_id = request.state.tenant_id
    individuals = await search_individuals_service(query, tenant_id)
    return [ind_to_search_response(ind) for ind in individuals]


@handle_http_exceptions
async def get_individual(individual_id: int) -> dict:
    """Get individual by ID with contacts and research data."""
    ind, contacts = await get_individual_service(individual_id)
    return ind_to_detail_response(ind, contacts=contacts, include_research=True)


@handle_http_exceptions
async def create_individual(request: Request, data: IndividualCreate) -> dict:
    """Create a new individual."""
    tenant_id = request.state.tenant_id
    user_id = request.state.user_id
    ind_data = data.model_dump(exclude_none=True)
    industry_ids = ind_data.pop("industry_ids", None)
    project_ids = ind_data.pop("project_ids", None)
    ind_data["created_by"] = user_id
    ind_data["updated_by"] = user_id
    ind = await create_individual_service(ind_data, tenant_id, industry_ids, project_ids)
    return ind_to_detail_response(ind, contacts=[])


@handle_http_exceptions
async def update_individual(request: Request, individual_id: int, data: IndividualUpdate) -> dict:
    """Update an individual."""
    user_id = request.state.user_id
    ind_data = data.model_dump(exclude_unset=True)
    industry_ids = ind_data.pop("industry_ids", None)
    project_ids = ind_data.pop("project_ids", None)
    project_ids_provided = "project_ids" in data.model_dump(exclude_unset=True)
    ind_data["updated_by"] = user_id
    ind = await update_individual_service(individual_id, ind_data, industry_ids, project_ids, project_ids_provided)
    return ind_to_detail_response(ind)


@handle_http_exceptions
async def delete_individual(individual_id: int) -> dict:
    """Delete an individual."""
    await delete_individual_service(individual_id)
    return {"success": True}


# Contact management

@handle_http_exceptions
async def get_individual_contacts(individual_id: int) -> List[dict]:
    """Get contacts for an individual."""
    contacts = await get_contacts_service(individual_id)
    return [contact_to_dict(c) for c in contacts]


@handle_http_exceptions
async def create_individual_contact(request: Request, individual_id: int, data: IndContactCreate) -> dict:
    """Create a contact for an individual."""
    user_id = request.state.user_id
    contact = await create_contact_service(individual_id, data.model_dump(), user_id)
    return contact_to_dict(contact)


@handle_http_exceptions
async def update_individual_contact(
    request: Request,
    individual_id: int,
    contact_id: int,
    data: IndContactUpdate,
) -> dict:
    """Update a contact for an individual."""
    user_id = request.state.user_id
    contact = await update_contact_service(individual_id, contact_id, data.model_dump(exclude_unset=True), user_id)
    return contact_to_dict(contact)


@handle_http_exceptions
async def delete_individual_contact(individual_id: int, contact_id: int) -> dict:
    """Delete a contact from an individual."""
    await delete_contact_service(individual_id, contact_id)
    return {"success": True}


# Individual-to-Individual Relationship management

@handle_http_exceptions
async def create_individual_relationship(
    individual_id: int,
    to_individual_id: int,
    relationship_type: Optional[str] = None,
    notes: Optional[str] = None,
) -> dict:
    """Create a relationship to another individual."""
    from services.individual_service import create_individual_relationship as create_rel_service

    relationship = await create_rel_service(
        from_individual_id=individual_id,
        to_individual_id=to_individual_id,
        relationship_type=relationship_type,
        notes=notes,
    )
    return {"id": relationship.id, "to_individual_id": relationship.to_individual_id}


@handle_http_exceptions
async def delete_individual_relationship(individual_id: int, relationship_id: int) -> dict:
    """Delete a relationship."""
    from services.individual_service import delete_individual_relationship as delete_rel_service

    await delete_rel_service(individual_id, relationship_id)
    return {"success": True}
