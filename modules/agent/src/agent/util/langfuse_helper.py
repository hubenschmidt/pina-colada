"""
Langfuse helper utility - conditionally initializes Langfuse only if credentials are available.
"""

import logging
import os

logger = logging.getLogger(__name__)


def get_langfuse_handler():
    """
    Get Langfuse callback handler if credentials are available, otherwise return None.
    
    This prevents warnings when Langfuse is not configured in production.
    """
    try:
        # Check if Langfuse credentials are available
        public_key = os.getenv("LANGFUSE_PUBLIC_KEY")
        secret_key = os.getenv("LANGFUSE_SECRET_KEY")
        langfuse_host = os.getenv("LANGFUSE_HOST")
        
        if not public_key or not secret_key:
            logger.debug("Langfuse credentials not found, skipping Langfuse initialization")
            return None
        
        from langfuse.langchain import CallbackHandler
        
        # Initialize with explicit credentials to prevent warnings
        handler_kwargs = {
            "public_key": public_key,
            "secret_key": secret_key,
        }
        
        if langfuse_host:
            handler_kwargs["host"] = langfuse_host
        
        return CallbackHandler(**handler_kwargs)
    except Exception as e:
        logger.debug(f"Failed to initialize Langfuse handler: {e}")
        return None

