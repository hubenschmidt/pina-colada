"""Fast-path nodes for token-optimized routing."""

from agent.nodes.fast_lookup import fast_lookup_node
from agent.nodes.format_response import create_format_response_node

__all__ = ["fast_lookup_node", "create_format_response_node"]
