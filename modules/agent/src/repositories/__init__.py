"""Repository layer for data access."""

# Job repository
from repositories.job_repository import (
    find_all_jobs,
    count_jobs,
    find_job_by_id,
    create_job,
    update_job,
    delete_job,
    find_job_by_company_and_title,
    find_all_statuses as find_all_job_statuses,
    find_jobs_with_status,
)

# Status repository
from repositories.status_repository import (
    find_all_statuses,
    find_statuses_by_category,
    find_status_by_id,
    find_status_by_name,
    create_status,
    update_status,
    delete_status,
)

# Organization repository
from repositories.organization_repository import (
    find_all_organizations,
    find_organization_by_id,
    find_organization_by_name,
    create_organization,
    get_or_create_organization,
    update_organization,
    delete_organization,
    search_organizations,
)

# Deal repository
from repositories.deal_repository import (
    find_all_deals,
    find_deal_by_id,
    find_deal_by_name,
    create_deal,
    get_or_create_deal,
    update_deal,
    delete_deal,
)

# Lead repository
from repositories.lead_repository import (
    find_all_leads,
    find_lead_by_id,
    create_lead,
    update_lead,
    delete_lead,
    find_leads_by_status,
)

__all__ = [
    # Job repository
    "find_all_jobs",
    "count_jobs",
    "find_job_by_id",
    "create_job",
    "update_job",
    "delete_job",
    "find_job_by_company_and_title",
    "find_all_job_statuses",
    "find_jobs_with_status",
    # Status repository
    "find_all_statuses",
    "find_statuses_by_category",
    "find_status_by_id",
    "find_status_by_name",
    "create_status",
    "update_status",
    "delete_status",
    # Organization repository
    "find_all_organizations",
    "find_organization_by_id",
    "find_organization_by_name",
    "create_organization",
    "get_or_create_organization",
    "update_organization",
    "delete_organization",
    "search_organizations",
    # Deal repository
    "find_all_deals",
    "find_deal_by_id",
    "find_deal_by_name",
    "create_deal",
    "get_or_create_deal",
    "update_deal",
    "delete_deal",
    # Lead repository
    "find_all_leads",
    "find_lead_by_id",
    "create_lead",
    "update_lead",
    "delete_lead",
    "find_leads_by_status",
]
