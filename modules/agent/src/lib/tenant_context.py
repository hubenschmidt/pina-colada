"""Tenant context for request-scoped tenant_id access."""

from contextvars import ContextVar
from typing import Optional

# Context variable for tenant_id - thread/async safe
_tenant_id_var: ContextVar[Optional[int]] = ContextVar("tenant_id", default=None)


def set_tenant_id(tenant_id: Optional[int]) -> None:
    """Set tenant_id for the current request context."""
    _tenant_id_var.set(tenant_id)


def get_tenant_id() -> Optional[int]:
    """Get tenant_id from the current request context."""
    return _tenant_id_var.get()
