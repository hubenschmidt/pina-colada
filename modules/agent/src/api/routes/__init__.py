"""Routes layer for API endpoints."""

from api.routes.jobs import router as jobs_routes
from api.routes.leads import router as leads_routes
from api.routes.auth import router as auth_routes
from api.routes.users import router as users_routes

__all__ = ["jobs_routes", "leads_routes", "auth_routes", "users_routes"]

