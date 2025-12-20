"""Routes layer for API endpoints."""

from api.routes.jobs import router as jobs_routes
from api.routes.leads import router as leads_routes
from api.routes.auth import router as auth_routes
from api.routes.users import router as users_routes
from api.routes.preferences import router as preferences_routes
from api.routes.organizations import router as organizations_routes
from api.routes.individuals import router as individuals_routes
from api.routes.industries import router as industries_routes
from api.routes.salary_ranges import router as salary_ranges_routes
from api.routes.employee_count_ranges import router as employee_count_ranges_routes
from api.routes.funding_stages import router as funding_stages_routes
from api.routes.notes import router as notes_routes
from api.routes.contacts import router as contacts_routes
from api.routes.accounts import router as accounts_routes
from api.routes.revenue_ranges import router as revenue_ranges_routes
from api.routes.technologies import router as technologies_routes
from api.routes.provenance import router as provenance_routes
from api.routes.reports import router as reports_routes
from api.routes.projects import router as projects_routes
from api.routes.opportunities import router as opportunities_routes
from api.routes.partnerships import router as partnerships_routes
from api.routes.tasks import router as tasks_routes
from api.routes.comments import router as comments_routes
from api.routes.notifications import router as notifications_routes
from api.routes.documents import router as documents_routes
from api.routes.tags import router as tags_routes
from api.routes.conversations import router as conversations_routes
from api.routes.usage import router as usage_routes
from api.routes.costs import router as costs_routes
from api.routes.metrics import router as metrics_routes

__all__ = [
    "jobs_routes",
    "leads_routes",
    "auth_routes",
    "users_routes",
    "preferences_routes",
    "organizations_routes",
    "individuals_routes",
    "industries_routes",
    "salary_ranges_routes",
    "employee_count_ranges_routes",
    "funding_stages_routes",
    "notes_routes",
    "contacts_routes",
    "accounts_routes",
    "revenue_ranges_routes",
    "technologies_routes",
    "provenance_routes",
    "reports_routes",
    "projects_routes",
    "opportunities_routes",
    "partnerships_routes",
    "tasks_routes",
    "comments_routes",
    "notifications_routes",
    "documents_routes",
    "tags_routes",
    "conversations_routes",
    "usage_routes",
    "costs_routes",
    "metrics_routes",
]

