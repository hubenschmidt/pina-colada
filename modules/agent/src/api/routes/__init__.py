"""Routes layer for API endpoints."""

from api.routes.jobs import router as jobs_routes
from api.routes.leads import router as leads_routes
from api.routes.auth import router as auth_routes
from api.routes.users import router as users_routes
from api.routes.preferences import router as preferences_routes
from api.routes.organizations import router as organizations_routes
from api.routes.individuals import router as individuals_routes
from api.routes.industries import router as industries_routes

__all__ = [
    "jobs_routes",
    "leads_routes",
    "auth_routes",
    "users_routes",
    "preferences_routes",
    "organizations_routes",
    "individuals_routes",
    "industries_routes",
]

