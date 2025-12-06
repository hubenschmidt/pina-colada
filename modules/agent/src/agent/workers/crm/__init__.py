"""CRM Worker - AI-assisted CRM with RAG and semi-automatic updates."""

from agent.workers.crm.crm_worker import create_crm_worker_node, route_from_crm_worker

__all__ = ["create_crm_worker_node", "route_from_crm_worker"]
